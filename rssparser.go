package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// RSS описывает корневую структуру RSS-ленты.
type RSS struct {
	Channel Channel `xml:"channel"`
}

// Channel описывает канал RSS-ленты, содержащий заголовок и список элементов.
type Channel struct {
	Title string `xml:"title"`
	Items []Item `xml:"item"`
}

// Item описывает отдельную статью (элемент RSS-ленты).
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func main() {
	// Проверяем, передан ли URL RSS-ленты в аргументах командной строки.
	if len(os.Args) < 2 {
		fmt.Println("Использование: rssparser <URL RSS-ленты>")
		os.Exit(1)
	}
	rssURL := os.Args[1]

	// Создаем HTTP-запрос с заголовком User-Agent.
	req, err := http.NewRequest("GET", rssURL, nil)
	if err != nil {
		fmt.Printf("Ошибка создания запроса: %v\n", err)
		os.Exit(1)
	}
	// Устанавливаем User-Agent, чтобы сервер воспринимал запрос как исходящий из браузера.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyRSSParser/1.0)")

	// Отправляем запрос.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Ошибка при выполнении запроса: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа.
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Не удалось получить данные, статус: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// Читаем тело ответа.
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Ошибка чтения данных: %v\n", err)
		os.Exit(1)
	}

	// Парсим XML-данные в структуру RSS.
	var rss RSS
	err = xml.Unmarshal(data, &rss)
	if err != nil {
		fmt.Printf("Ошибка парсинга XML: %v\n", err)
		os.Exit(1)
	}

	// Если канал пустой или не содержит статей, выводим сообщение.
	if rss.Channel.Title == "" && len(rss.Channel.Items) == 0 {
		fmt.Println("Не удалось найти статьи в RSS-ленте. Возможно, формат ленты отличается от ожидаемого.")
		os.Exit(1)
	}

	// Выводим заголовок канала и список заголовков статей.
	fmt.Printf("Заголовки статей из RSS-ленты '%s':\n", rss.Channel.Title)
	for i, item := range rss.Channel.Items {
		fmt.Printf("%d. %s\n", i+1, item.Title)
	}
}
