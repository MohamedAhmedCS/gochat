package main

import (
    "flag"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")

// upgrader upgrades HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    // Allow connections from any origin for demo purposes.
    CheckOrigin: func(r *http.Request) bool { return true },
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
    // Upgrade the HTTP connection to a WebSocket connection.
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
    client.hub.register <- client

    // Launch goroutines for reading and writing.
    go client.writePump()
    go client.readPump()
}

func serveHome(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }
    if r.Method != "GET" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    http.ServeFile(w, r, "home.html")
}

func main() {
    flag.Parse()
    hub := newHub()
    go hub.run()

    http.HandleFunc("/", serveHome)
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        serveWs(hub, w, r)
    })

    log.Println("Server started on", *addr)
    if err := http.ListenAndServe(*addr, nil); err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
