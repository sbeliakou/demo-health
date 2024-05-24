package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

var serviceEnabled bool
var mu sync.RWMutex
var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]bool)

func changeServiceEnabled(values ...bool) {
	if len(values) > 0 {
		serviceEnabled = values[0]
	} else {
		serviceEnabled = !serviceEnabled
	}

	if serviceEnabled {
		log.Println("Health check set to enabled")
	} else {
		log.Println("Health check set to disabled")
	}
}

func setServiceEnabled(value bool) {
	changeServiceEnabled(value)
}

func main() {
	setServiceEnabled(true)

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/toggle", handleToggle)
	http.HandleFunc("/liveness-probe-demo-ws", handleConnections)
	http.HandleFunc("/", handleHome)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	if serviceEnabled {
		w.WriteHeader(http.StatusOK)
		log.Printf("Health endpoint reported %d (healthy)", http.StatusOK)
		fmt.Fprintf(w, "healthy")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Printf("Health endpoint reported %d (unhealthy)", http.StatusServiceUnavailable)
		fmt.Fprintf(w, "unhealthy")
	}
	mu.RUnlock()
}

func handleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		mu.Lock()
		changeServiceEnabled()
		mu.Unlock()
		notifyClients()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		log.Println("Error executing template:", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer ws.Close()

	clients[ws] = true

	err = ws.WriteJSON(serviceEnabled)
	if err != nil {
		log.Println("Error writing to client:", err)
	}

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			delete(clients, ws)
			break
		}
	}
}

func notifyClients() {
	for client := range clients {
		err := client.WriteJSON(serviceEnabled)
		if err != nil {
			log.Println("Error notifying client:", err)
		}
	}
}
