package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// Структура сообщения
type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

// Настройка WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Список клиентов и канал для вещания
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

func main() {
	// Отдача статики
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// WebSocket роут
	http.HandleFunc("/ws", handleConnections)

	// Обработка сообщений
	go handleMessages()

	fmt.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Ошибка сервера:", err)
	}
}

// Обработка входящих соединений
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Апгрейд HTTP до WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Ошибка WebSocket:", err)
		return
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Ошибка чтения:", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

// Рассылка сообщений всем клиентам
func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println("Ошибка отправки:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
