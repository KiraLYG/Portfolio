package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// findDuplicates обходит рекурсивно указанную директорию, вычисляет хэш для каждого файла,
// и выводит группы файлов с одинаковыми хэшами (то есть дубликаты).
func findDuplicates(dir string) {
	// Создаем карту: хэш -> список путей к файлам с таким хэшом.
	duplicates := make(map[string][]string)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// При ошибке пропускаем данный файл.
			return nil
		}
		if info.IsDir() {
			return nil
		}

		// Открываем файл для чтения.
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		// Вычисляем SHA-256 хэш файла.
		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return nil
		}
		hashValue := hex.EncodeToString(hasher.Sum(nil))
		duplicates[hashValue] = append(duplicates[hashValue], path)
		return nil
	})

	if err != nil {
		log.Fatalf("Ошибка обхода директории: %v", err)
	}

	// Выводим группы дубликатов (если найдено больше одного файла с одинаковым хэшем).
	fmt.Println("Найденные дубликаты:")
	found := false
	for hash, paths := range duplicates {
		if len(paths) > 1 {
			found = true
			fmt.Printf("Hash: %s\n", hash)
			for _, p := range paths {
				fmt.Printf("  %s\n", p)
			}
			fmt.Println()
		}
	}
	if !found {
		fmt.Println("Дубликаты не найдены.")
	}
}

// renameFiles переименовывает все файлы (без рекурсии) в указанной директории.
// Новое имя формируется по схеме: <prefix>_<номер>.<расширение>
func renameFiles(dir string, prefix string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Ошибка чтения директории: %v", err)
	}

	counter := 1
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		oldPath := filepath.Join(dir, entry.Name())
		ext := filepath.Ext(entry.Name())
		newName := fmt.Sprintf("%s_%03d%s", prefix, counter, ext)
		newPath := filepath.Join(dir, newName)
		if err := os.Rename(oldPath, newPath); err != nil {
			log.Printf("Ошибка переименования файла %s: %v", oldPath, err)
		} else {
			fmt.Printf("Переименован: %s -> %s\n", oldPath, newPath)
			counter++
		}
	}
}

func printUsage() {
	fmt.Println("Использование:")
	fmt.Println("  fileutil duplicates <directory>         - поиск дубликатов файлов")
	fmt.Println("  fileutil rename <directory> <prefix>      - переименование файлов в директории с заданным префиксом")
}

func main() {
	// Дополнительная обработка аргументов для корректной работы через "go run" в Termux.
	// Если первым аргументом является имя Go-файла или абсолютный путь, сдвигаем аргументы.
	if len(os.Args) > 1 && (strings.HasSuffix(os.Args[1], ".go") || filepath.IsAbs(os.Args[1])) {
		if len(os.Args) > 2 {
			os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		}
	}

	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "duplicates":
		// Пример: fileutil duplicates /path/to/directory
		dir := os.Args[2]
		findDuplicates(dir)
	case "rename":
		// Пример: fileutil rename /path/to/directory newname
		if len(os.Args) < 4 {
			fmt.Println("Укажите директорию и префикс для переименования файлов.")
			printUsage()
			os.Exit(1)
		}
		dir := os.Args[2]
		prefix := os.Args[3]
		renameFiles(dir, prefix)
	default:
		fmt.Println("Неизвестная команда:", command)
		printUsage()
		os.Exit(1)
	}
}
