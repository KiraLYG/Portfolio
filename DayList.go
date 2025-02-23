package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Структура заметки
type Note struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

var notes []Note

// Основной файл хранения заметок
const notesFile = "notes.json"

// Функция загрузки файла
func loadNotes() error {
	data, err := ioutil.ReadFile(notesFile)
	if err != nil {
		if os.IsNotExist(err) {
			notes = []Note{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &notes)
}

// Функция сохранения файла
func saveNotes() error {
	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(notesFile, data, 0644)
}

// Функция добавления заметки
func addNote(content string) {
	id := 1
	if len(notes) > 0 {
		id = notes[len(notes)-1].ID + 1
	}
	note := Note{
		ID:        id,
		Content:   content,
		CreatedAt: time.Now(),
	}
	notes = append(notes, note)
	fmt.Printf("Заметка добавлена с ID %d\n", note.ID)
}

// Функция просмотра всех заметок
func listNotes() {
	if len(notes) == 0 {
		fmt.Println("Заметок не найдено.")
		return
	}
	for _, note := range notes {
		fmt.Printf("ID: %d\nСодержание: %s\nДата создания: %s\n\n",
			note.ID, note.Content, note.CreatedAt.Format(time.RFC1123))
	}
}

// Функция удаления заметки
func deleteNote(id int) {
	index := -1
	for i, note := range notes {
		if note.ID == id {
			index = i
			break
		}
	}
	if index == -1 {
		fmt.Printf("Заметка с ID %d не найдена.\n", id)
		return
	}
	notes = append(notes[:index], notes[index+1:]...)
	fmt.Printf("Заметка с ID %d удалена.\n", id)
}

// Основная логика
func main() {
	// Загружаем заметки из файла
	if err := loadNotes(); err != nil {
		fmt.Printf("Ошибка загрузки заметок: %v\n", err)
		os.Exit(1)
	}

	// Если аргументов меньше 2, выводим использование
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run DayList.go [add|list|delete] [аргументы...]")
		os.Exit(1)
	}

	// Определяем, где находится команда
	var command string
	// Если os.Args[1] выглядит как абсолютный путь (что бывает при go run в Termux), то берем os.Args[2]
	if filepath.IsAbs(os.Args[1]) && len(os.Args) > 2 {
		command = os.Args[2]
		// Чтобы дальше аргументы сдвинулись на один элемент:
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	} else {
		command = os.Args[1]
	}

	switch command {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Использование: go run DayList.go add \"Содержание заметки\"")
			os.Exit(1)
		}
		content := os.Args[2]
		addNote(content)
	case "list":
		listNotes()
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Использование: go run DayList.go delete <ID_заметки>")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("ID заметки должно быть числом.")
			os.Exit(1)
		}
		deleteNote(id)
	default:
		fmt.Println("Неизвестная команда:", command)
		fmt.Println("Доступные команды: add, list, delete")
		os.Exit(1)
	}

	// Сохраняем заметки в файл
	if err := saveNotes(); err != nil {
		fmt.Printf("Ошибка сохранения заметок: %v\n", err)
		os.Exit(1)
	}
}
