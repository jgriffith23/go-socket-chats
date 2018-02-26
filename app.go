// A real-time chat application server

// Originally derived from this tutorial:
// https://scotch.io/bar-talk/build-a-realtime-chat-server-with-go-and-websockets

package main

import (
    "html/template"
    "log"
    "net/http"
    "github.com/getsentry/raven-go"
    // "github.com/gorilla/sessions"
    "github.com/gorilla/websocket"
)

// FIXME: Rework this code so that the variables don't have to be global?

// Golang note: calling make() actually initializes the map in memory.
// Connected clients.
var clients = make(map[*websocket.Conn]bool)

// A queue for messages sent by clients. Creates a Go channel, which
// lets you send/receive values. Channel operator is <-.
var broadcast = make(chan Message)

// An object that turns a normal HTTP connection into a WebSocket.
// FIXME: Write a CheckOrigin function for the Upgrader.
var upgrader = websocket.Upgrader{}

// FIXME: Move type definitions out of server file?
// An object to contain user messages.
type Message struct {
    Username string `json:"username"`
    Message string `json:"message"`
}

var templates *template.Template

func init() { 

    // Gather templates.
    templates = template.Must(template.ParseGlob("templates/*.gohtml"))
}

type augError struct {
    Error error
    Message string
    Code int
}

//////////////////////////////////////////////////////////////
// Custom server. Log errors with Sentry and render templates.
//////////////////////////////////////////////////////////////

// Wrap handler functions to add custom actions for any request that
// should return a view.

type viewHandler func(http.ResponseWriter, *http.Request) (ae *augError, tpl string)
func (fn viewHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
    augErr, template := fn(res, req)
    if augErr != nil {
        http.Error(
            res,
            augErr.Message,
            augErr.Code,
        )
        raven.CaptureErrorAndWait(augErr.Error, nil)
        return
    }

    err := templates.ExecuteTemplate(res, template, nil)
    if err != nil {
        http.Error(
            res,
            "Page could not be displayed.",
            http.StatusInternalServerError,
        )
        raven.CaptureErrorAndWait(err, nil)
        return
    }
}

// FIXME: Add custom functionality for logging, error handling, and template 
// rendering.

func main() {

    http.Handle("/", viewHandler(index))
    http.HandleFunc("/websocket", handleConnections)

    staticServer := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", staticServer))

    // A goroutine. Concurrent process. Passes messages from broadcast to
    // clients.
    go handleMessages()

    log.Println("Started serving on port 8080.")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatalln("ListenAndServe: ", err)
    }
}

// index serves the homepage.
func index(res http.ResponseWriter, req *http.Request) (ae *augError, tpl string) {
    return nil, "index.gohtml"
}

// handleConnections converts a GET request into a web socket and 
// registers a new client. 
func handleConnections(res http.ResponseWriter, req *http.Request) {

    // Create connection.
    conn, err := upgrader.Upgrade(res, req, nil)
    if err != nil {
        raven.CaptureErrorAndWait(err, nil)
        log.Fatalln(err)
    }

    // We should close the socket connection when this function is done.
    defer conn.Close()

    // Add new client
    clients[conn] = true
    log.Println("user connected")

    // Golang note: Example of infinite loop syntax.
    for {

        // Extract message JSON into a struct
        var msg Message
        err := conn.ReadJSON(&msg)
        if err != nil {

            // An error in connection doesn't mean the server should crash.
            // Assume client disconnected and remove from registry.
            raven.CaptureErrorAndWait(err, nil)
            delete(clients, conn)
            break
        }

        // Sent new message to the global channel.
        broadcast <- msg
    }
}

// handleMessages fetches messages from the channel and sends them
// to all registered clients as JSON.
func handleMessages() {

    // FIXME: Could this function take a channel as a parameter so that we don't
    // need a global?
    for {
        msg := <- broadcast
        log.Println("handleMessages: ", msg)

        // Golang note: range is a bit like range() in Python. Gets indices
        // for slices; gets key for maps.

        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {

                // Again, assume client disconnected if there's an error.
                raven.CaptureErrorAndWait(err, nil)
                log.Println("handleMessages error: ", err)
                client.Close()
                delete(clients, client)
            }
        }
    }
}