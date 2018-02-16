// A real-time chat application server
// Following this tutorial:
// https://scotch.io/bar-talk/build-a-realtime-chat-server-with-go-and-websockets

package main

import (
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)

// FIXME: Rework this code so that the variables don't have to be global.

// Connected clients.
// Golang note: calling make() actually initializes the map in memory.
var clients = make(map[*websocket.Conn]bool)

// A queue for messages sent by clients. Creates a Go channel, which
// lets you send/receive values. Channel operator is <-.
var broadcast = make(chan Message)

// An object that turns a normal HTTP connection into a WebSocket.
// FIXME: Write a CheckOrigin function for the Upgrader.
var upgrader = websocket.Upgrader{}

// An object to contain user messages.
// FIXME: Move type definitions out of server file?
type Message struct {
    Email string `json:"email"`
    Username string `json:"username"`
    Message string `json:"message"`
}

// FIXME: Add custom functionality for logging, error handling, and template 
// rendering.

func main() {
    // Simple file server. Serves HTML, CSS, JS.
    fileServer := http.FileServer(http.Dir("../public"))

    // Homepage uses the file server.
    http.Handle("/", fileServer)

    // Websocket connections will use a different server.
    http.HandleFunc("/websocket", handleConnections)

    // A goroutine. Concurrent process. Passes messages from broadcast to
    // clients.
    go handleMessages()

    log.Println("Started serving on port 8080.")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatalln("ListenAndServe: ", err)
    }
}

// Convert GET request into a web socket, register client,. 
func handleConnections(res http.ResponseWriter, req *http.Request) {
    // Create connection.
    conn, err := upgrader.Upgrade(res, req, nil)
    if err != nil {
        log.Fatalln(err)
    }

    // We should close the socket connection when this function is done.
    defer conn.Close()

    // Add new client
    clients[conn] = true

    // Golang note: Example of infinite loop syntax.
    for {
        // Extract message JSON into a struct
        var msg Message
        err := conn.ReadJSON(&msg)
        if err != nil {
            // An error in connection doesn't mean the server should crash.
            // Assume client disconnected and remove from registry.
            log.Println("error: ", err)
            delete(clients, conn)
            break
        }

        // Sent new message to the global channel.
        broadcast <- msg
    }
}

// Fetch messages from channel and send to all registered clients as JSON.
func handleMessages() {
    // FIXME: Could this function take a channel as a parameter so that we don't
    // need a global?
    for {
        msg := <-broadcast

        // Golang note: range is a bit like range() in Python. Gets indices
        // for slices; gets key for maps.
        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
                // Again, assume client disconnected if there's an error.
                log.Println("error: ", err)
                client.Close()
                delete(clients, client)
            }
        }
    }
}