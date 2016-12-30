// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

// Upgrader specifies parameters for upgrading an HTTP connection to a WebSocket connection.
// Without parameters, default options will be applied (ReadBufferSize WriteBufferSize are set to 4096
var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {

	// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
	// c (i.e. Conn type) represents a WebSocket connection.
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()

	for {

		command := "command" + strconv.Itoa(rand.Intn(5))

		err = c.WriteMessage(1, []byte(command))
		if err != nil {
			log.Println("write:", err)
			break
		}

		log.Printf("sent: %s", command)

		// messageType int, message []byte, err error
		messageType, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("recv: %s (type=%d)", message, messageType)

		time.Sleep(time.Second * 2)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/command", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
