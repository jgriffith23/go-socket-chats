// A real-time chat application server
// Derived from this tutorial:
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

// A queue for messages sent by clients.
var broadcast = make(chan Message)

// An object that turns a normal HTTP connection into a WebSocket.
// FIXME: Write a CheckOrigin function for the Upgrader.
var upgrader = websocket.Upgrader{}

// An object to contain user messages.
type Message struct {
    Email string `json:"email"`
    Username string `json:"username"`
    Message string `json:"message"`
}