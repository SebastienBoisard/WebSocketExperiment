// +build ignore

package main

import (
	"encoding/json"
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

func handleCommandFunc(w http.ResponseWriter, r *http.Request) {

	// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
	// c (i.e. Conn type) represents a WebSocket connection.
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()

	for {

		id := strconv.Itoa(rand.Intn(4))

		var action Action

		switch id {
		case "1":
			action = Action{
				Name:       "command" + id,
				Parameters: []ActionParameter{{Name: "param", Value: int(1)}},
			}

		case "2":
			action = Action{
				Name:       "command" + id,
				Parameters: []ActionParameter{{Name: "param", Value: "2"}},
			}

		case "3":
			action = Action{
				Name:       "command" + id,
				Parameters: []ActionParameter{{Name: "param1", Value: 3.3}, {Name: "param2", Value: true}},
			}

		default:
			action = Action{
				Name: "command" + id,
			}
		}

		jsonCommand, err := json.Marshal(action)
		if err != nil {
			log.Println("marshal:", err)
			return
		}

		err = c.WriteMessage(websocket.TextMessage, []byte(jsonCommand))
		if err != nil {
			log.Println("write:", err)
			break
		}

		log.Printf("sent: %s", jsonCommand)

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
	http.HandleFunc("/command", handleCommandFunc)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
