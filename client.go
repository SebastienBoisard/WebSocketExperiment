// +build ignore

package main

import (
	"flag"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

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

		m := string(message[:])

		switch m {
		case "command1":
			log.Println("received command1")

			err := c.WriteMessage(websocket.TextMessage, []byte("response1"))
			if err != nil {
				log.Println("write:", err)
				continue
			}

		case "command2":
			log.Println("received command2")

			err := c.WriteMessage(websocket.TextMessage, []byte("response2"))
			if err != nil {
				log.Println("write:", err)
				continue
			}

		case "command4":
			log.Println("received command4")

			err := c.WriteMessage(websocket.TextMessage, []byte("response4"))
			if err != nil {
				log.Println("write:", err)
				continue
			}

		default:
			log.Println("received unknown command")

			err := c.WriteMessage(websocket.TextMessage, []byte("unknown response"))
			if err != nil {
				log.Println("write:", err)
				continue
			}
		}
	}
}
   