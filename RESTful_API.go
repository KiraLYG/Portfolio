package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Task описывает структуру задачи.
type Task struct {
	ID          int       `json:"id"`          // Уникальный идентификатор задачи
	Title       string    `json:"title"`       // Заголовок задачи
	Description string    `json:"description"` // Описание задачи
	Completed   bool      `json:"completed"`   // Статус выполнения
	CreatedAt   time.Time `json:"created_at"`  // Дата и время создания задачи
}

// tasks хранит список задач в памяти.
var tasks []Task

// nextID используется для генерации нового ID задачи.
var nextID int = 1

// mutex для защиты доступа к срезу tasks.
var mutex = &sync.Mutex{}

func main() {
	// Регистрируем обработчики для наших маршрутов.
	http.HandleFunc("/tasks", tasksHandler)     // для GET /tasks и POST /tasks
	http.HandleFunc("/tasks/", taskHandler)       // для GET, PUT, DELETE с конкретным ID (например, /tasks/1)

	fmt.Println("Сервер запущен на порту 8080")
	// Запускаем HTTP сервер на порту 8080.
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// tasksHandler обрабатывает запросы к коллекции задач.
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// Получение списка задач
	case http.MethodGet:
		getTasks(w, r)
	// Создание новой задачи
	case http.MethodPost:
		createTask(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// taskHandler обрабатывает запросы для работы с конкретной задачей.
func taskHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из URL. Ожидается формат /tasks/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Неверный ID задачи", http.StatusBadRequest)
		return
	}

	switch r.Method {
	// Получение задачи по ID
	case http.MethodGet:
		getTask(w, r, id)
	// Обновление задачи по ID
	case http.MethodPut:
		updateTask(w, r, id)
	// Удаление задачи по ID
	case http.MethodDelete:
		deleteTask(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// getTasks возвращает список всех задач.
func getTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mutex.Lock()
	defer mutex.Unlock()
	json.NewEncoder(w).Encode(tasks)
}

// getTask возвращает одну задачу по ID.
func getTask(w http.ResponseWriter, r *http.Request, id int) {
	w.Header().Set("Content-Type", "application/json")
	mutex.Lock()
	defer mutex.Unlock()
	for _, t := range tasks {
		if t.ID == id {
			json.NewEncoder(w).Encode(t)
			return
		}
	}
	http.Error(w, "Задача не найдена", http.StatusNotFound)
}

// createTask создаёт новую задачу.
func createTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var task Task
	// Декодируем JSON из тела запроса.
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Неверные данные", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	// Присваиваем задаче уникальный ID и заполняем дату создания.
	task.ID = nextID
	nextID++
	task.CreatedAt = time.Now()
	tasks = append(tasks, task)
	mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// updateTask обновляет данные задачи по ID.
func updateTask(w http.ResponseWriter, r *http.Request, id int) {
	w.Header().Set("Content-Type", "application/json")
	var updatedTask Task
	// Декодируем обновлённые данные задачи.
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, "Неверные данные", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	for i, t := range tasks {
		if t.ID == id {
			// Обновляем поля задачи. ID и CreatedAt остаются неизменными.
			tasks[i].Title = updatedTask.Title
			tasks[i].Description = updatedTask.Description
			tasks[i].Completed = updatedTask.Completed
			json.NewEncoder(w).Encode(tasks[i])
			return
		}
	}
	http.Error(w, "Задача не найдена", http.StatusNotFound)
}

// deleteTask удаляет задачу по ID.
func deleteTask(w http.ResponseWriter, r *http.Request, id int) {
	mutex.Lock()
	defer mutex.Unlock()
	for i, t := range tasks {
		if t.ID == id {
			// Удаляем задачу из среза.
			tasks = append(tasks[:i], tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Задача не найдена", http.StatusNotFound)
}
