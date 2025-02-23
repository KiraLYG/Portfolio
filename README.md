# 🚀 Portfolio on Go 
The repository contains projects that demonstrate backend development skills in Go.
Here you will find implementation of RESTful API, websockets, parsers and utilities for working with data.
*Target*: Demonstrate ability to solve a variety of problems, write clean code, and work with popular tools in the Go ecosystem.

## 🛠 Technologies

    Language: Go 1.21+

    Libraries:

        gorilla/websocket

        encoding/json, encoding/xml

    Tools: Go Modules, Git

## 📂 Projects

### **DayList.go**

add - Adding a note (go run DayList.go add "Your Note")

list - View all notes (go run DayList.go list)

delete - Deleting a note by id (go run DayList.go delete (ID))

### **RESTful_API.go**

**Description**: Microservice for managing resources (tasks, users) with support for CRUD operations.

First, the main RESTful_API.go file is started

Management commands:

Get a list of tasks:
curl http://localhost:8080/tasks

Create a new task:
curl -X POST http://localhost:8080/tasks \
     -H "Content-Type: application/json" \
     -d '{"title": "Первая задача", "description": "Описание задачи", "completed": false}'

Get a task by ID:
curl http://localhost:8080/tasks/1

Update a task by ID:
curl -X PUT http://localhost:8080/tasks/1 \
     -H "Content-Type: application/json" \
     -d '{"title": "Обновлённая задача", "description": "Новое описание", "completed": true}'

Delete an task by ID
curl -X DELETE http://localhost:8080/tasks/1

### **Tracker.go**

**Description**: CLI utility for tracking time and tasks.

Creating an issue:
go run Tracker.go start "Your task"
If a task with that name exists, a new session is added to it, otherwise a new task is created.

Check the issue status:
go run Tracker.go status

Stop the session:
go run Tracker.go stop

Displays a list of all tasks with a total time count for all sessions:
go run Tracker.go list

### **rssparser.go**

**Description**: Parsing RSS feeds (for example, news sites) into structured data.

Example:
go run rssparser.go https://habr.com/ru/rss/all/all/

### **fileutil.go**

This command will recursively traverse the specified directory and output groups of duplicates:
go run fileutil.go duplicates /path/to/directory

This command will rename all files in the specified directory, adding the specified prefix and sequence number to each name:
go run fileutil.go rename /path/to/directory newprefix

### **WebChat**

**Description**: Chat using websockets for instant messaging.

! Attention! This file uses a package github.com/gorilla/websocket . Install this package before launching.

First you need to run the main file WebChat.go
go run WebChat.go

After that, you need to go to http://localhost:8080
