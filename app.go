// A real-time chat application server
// Derived from this tutorial:
// https://scotch.io/bar-talk/build-a-realtime-chat-server-with-go-and-websockets

package main

import (
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)
