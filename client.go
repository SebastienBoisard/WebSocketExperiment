// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type Command struct {
	Name   string
	Result string
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {

	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/command"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			continue
		}

		if messageType != websocket.TextMessage {
			log.Println("received a non text message")
			continue
		}

		var command Command
		json.Unmarshal(message, &command)

		log.Println("received ", command.Name)

		err = c.WriteMessage(websocket.TextMessage, []byte(command.Result))
		if err != nil {
			log.Println("write:", err)
			continue
		}

	}
}
