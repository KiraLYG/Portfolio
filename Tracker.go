package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Session представляет период выполнения задачи.
type Session struct {
	Start time.Time `json:"start"` // Время начала сессии
	End   time.Time `json:"end"`   // Время окончания (если не установлен, сессия активна)
}

// Task представляет задачу с набором сессий.
type Task struct {
	ID       int       `json:"id"`       // Уникальный идентификатор задачи
	Name     string    `json:"name"`     // Название задачи
	Sessions []Session `json:"sessions"` // Список сессий выполнения
}

var tasks []Task

// Имя файла для хранения данных
const dataFile = "tracker.json"

// loadTasks загружает задачи из файла.
// Если файла нет — создаёт пустой список.
func loadTasks() error {
	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			tasks = []Task{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &tasks)
}

// saveTasks сохраняет задачи в файл в формате JSON.
func saveTasks() error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dataFile, data, 0644)
}

// findActiveTask ищет задачу с активной (незавершённой) сессией.
func findActiveTask() *Task {
	for i := range tasks {
		if len(tasks[i].Sessions) > 0 {
			// Последняя сессия считается активной, если время окончания не установлено
			lastSession := tasks[i].Sessions[len(tasks[i].Sessions)-1]
			if lastSession.End.IsZero() {
				return &tasks[i]
			}
		}
	}
	return nil
}

// startTask запускает сессию для задачи с указанным именем.
func startTask(taskName string) {
	if active := findActiveTask(); active != nil {
		fmt.Printf("Нельзя начать новую задачу, пока не остановлена активная: %s\n", active.Name)
		return
	}

	// Ищем задачу по имени (без учёта регистра)
	var task *Task
	for i := range tasks {
		if strings.EqualFold(tasks[i].Name, taskName) {
			task = &tasks[i]
			break
		}
	}

	// Если задача не найдена, создаём новую с уникальным ID.
	if task == nil {
		newID := 1
		for _, t := range tasks {
			if t.ID >= newID {
				newID = t.ID + 1
			}
		}
		newTask := Task{
			ID:       newID,
			Name:     taskName,
			Sessions: []Session{},
		}
		tasks = append(tasks, newTask)
		task = &tasks[len(tasks)-1]
	}

	// Добавляем новую сессию с текущим временем начала.
	session := Session{
		Start: time.Now(),
	}
	task.Sessions = append(task.Sessions, session)

	if err := saveTasks(); err != nil {
		fmt.Println("Ошибка сохранения данных:", err)
		return
	}
	fmt.Printf("Начата задача: %s\n", task.Name)
}

// stopTask завершает активную сессию, если таковая существует.
func stopTask() {
	active := findActiveTask()
	if active == nil {
		fmt.Println("Нет активной задачи для остановки.")
		return
	}
	// Завершаем последнюю сессию
	idx := len(active.Sessions) - 1
	active.Sessions[idx].End = time.Now()
	duration := active.Sessions[idx].End.Sub(active.Sessions[idx].Start)

	if err := saveTasks(); err != nil {
		fmt.Println("Ошибка сохранения данных:", err)
		return
	}
	fmt.Printf("Задача '%s' остановлена. Время выполнения: %s\n", active.Name, duration)
}

// status выводит информацию об активной задаче и времени, которое прошло с её начала.
func status() {
	active := findActiveTask()
	if active == nil {
		fmt.Println("Нет активной задачи.")
		return
	}
	lastSession := active.Sessions[len(active.Sessions)-1]
	duration := time.Since(lastSession.Start)
	fmt.Printf("Активная задача: %s, время с начала: %s\n", active.Name, duration)
}

// listTasks выводит список всех задач с суммарным временем выполнения.
func listTasks() {
	if len(tasks) == 0 {
		fmt.Println("Нет записанных задач.")
		return
	}
	for _, task := range tasks {
		total := time.Duration(0)
		for _, s := range task.Sessions {
			// Если сессия активна — учитываем время от начала до текущего момента
			if s.End.IsZero() {
				total += time.Since(s.Start)
			} else {
				total += s.End.Sub(s.Start)
			}
		}
		fmt.Printf("Задача: %s, Общее время: %s\n", task.Name, total)
	}
}

// printUsage выводит справку по использованию приложения.
func printUsage() {
	fmt.Println("Использование:")
	fmt.Println("  tracker start \"Название задачи\"  - начать отслеживание задачи")
	fmt.Println("  tracker stop                      - остановить активную задачу")
	fmt.Println("  tracker status                    - показать активную задачу")
	fmt.Println("  tracker list                      - показать список задач с общим временем")
}

func main() {
	// Дополнительная обработка аргументов для корректной работы через "go run" в Termux.
	// Если первым аргументом является имя файла или абсолютный путь — сдвигаем аргументы.
	if len(os.Args) > 1 && (strings.HasSuffix(os.Args[1], ".go") || filepath.IsAbs(os.Args[1])) {
		if len(os.Args) > 2 {
			os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		}
	}

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// Загружаем данные из файла
	if err := loadTasks(); err != nil {
		fmt.Println("Ошибка загрузки данных:", err)
		return
	}

	command := os.Args[1]
	switch command {
	case "start":
		if len(os.Args) < 3 {
			fmt.Println("Укажите название задачи для запуска.")
			printUsage()
			return
		}
		taskName := os.Args[2]
		startTask(taskName)
	case "stop":
		stopTask()
	case "status":
		status()
	case "list":
		listTasks()
	default:
		printUsage()
	}
}
