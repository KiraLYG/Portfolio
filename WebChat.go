package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Client представляет одного подключённого клиента.
type Client struct {
	hub  *Hub              // Ссылка на центральный "хаб" для обмена сообщениями.
	conn *websocket.Conn   // WebSocket соединение.
	send chan []byte       // Канал для отправки сообщений клиенту.
}

// Hub управляет всеми клиентами и рассылкой сообщений.
type Hub struct {
	clients    map[*Client]bool // Зарегистрированные клиенты.
	broadcast  chan []byte      // Канал для входящих сообщений, которые нужно разослать.
	register   chan *Client     // Канал для регистрации новых клиентов.
	unregister chan *Client     // Канал для удаления клиентов.
}

// newHub создаёт и возвращает новый Hub.
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// run запускает основной цикл обработки событий Hub.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			// Рассылаем сообщение всем подключённым клиентам.
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// upgrader поднимает HTTP соединение до WebSocket.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Разрешаем подключения с любых источников.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// readPump читает сообщения от клиента и отправляет их в Hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	// Обработка pong-сообщений для поддержания соединения.
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка чтения: %v", err)
			}
			break
		}
		// Отправляем полученное сообщение на рассылку.
		c.hub.broadcast <- message
	}
}

// writePump отправляет сообщения из канала send клиенту.
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Канал закрыт, отправляем сообщение о закрытии.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Проверяем, есть ли ещё сообщения в канале и отправляем их одним пакетом.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// Отправляем ping для поддержания соединения.
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs обрабатывает HTTP-запрос на подключение к WebSocket.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Обновляем HTTP соединение до WebSocket.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка при обновлении до WebSocket:", err)
		return
	}
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	hub.register <- client

	// Запускаем горутины для чтения и записи.
	go client.writePump()
	go client.readPump()
}

// HTML-шаблон для клиентской части чата.
var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Чат на WebSocket</title>
	<style>
		body { font-family: Arial, sans-serif; }
		#chat { height: 300px; overflow-y: scroll; border: 1px solid #ccc; padding: 10px; }
		#message { width: 80%; }
	</style>
</head>
<body>
	<h1>Чат на WebSocket</h1>
	<div id="chat"></div>
	<input id="message" type="text" placeholder="Введите сообщение..." autofocus>
	<button onclick="sendMessage()">Отправить</button>

	<script>
		var conn = new WebSocket("ws://" + window.location.host + "/ws");
		var chat = document.getElementById("chat");
		conn.onmessage = function(evt) {
			var messages = evt.data.split('\n');
			messages.forEach(function(message) {
				var div = document.createElement("div");
				div.textContent = message;
				chat.appendChild(div);
				chat.scrollTop = chat.scrollHeight;
			});
		};

		function sendMessage() {
			var input = document.getElementById("message");
			if (!input.value) {
				return;
			}
			conn.send(input.value);
			input.value = "";
		}

		// Отправка сообщения по нажатию Enter.
		document.getElementById("message").addEventListener("keyup", function(event) {
			if (event.key === "Enter") {
				sendMessage();
			}
		});
	</script>
</body>
</html>
`))

// serveHome отдает HTML страницу с клиентом чата.
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Страница не найдена", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTemplate.Execute(w, r.Host)
}

func main() {
	// Флаг для указания адреса сервера (по умолчанию :8080).
	addr := flag.String("addr", ":8080", "http service address")
	flag.Parse()

	hub := newHub()
	go hub.run()

	// Маршруты:
	// "/" отдает HTML клиента.
	http.HandleFunc("/", serveHome)
	// "/ws" обрабатывает WebSocket подключения.
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	fmt.Println("Сервер запущен на", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

