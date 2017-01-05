// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"bufio"
	"os"
	"fmt"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

// Upgrader specifies parameters for upgrading an HTTP connection to a WebSocket connection.
// Without parameters, default options will be applied (ReadBufferSize WriteBufferSize are set to 4096
var upgrader = websocket.Upgrader{} // use default options

type WebSocketStore struct {
	token  string
	socket *websocket.Conn
}

var clients []WebSocketStore

func handleActionFunc(w http.ResponseWriter, r *http.Request) {

	t := r.URL.Query().Get("token")

	// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
	// c (i.e. Conn type) represents a WebSocket connection.
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	clients = append(clients, WebSocketStore{token: t, socket: c})

	// TODO: find when to close the WebSocket
	//defer c.Close()
}

func sendAction(c *websocket.Conn) {

	id := strconv.Itoa(rand.Intn(4))

	var action Action

	switch id {
	case "1":
		action = Action{
			Name:       "action" + id,
			Parameters: []ActionParameter{{Name: "param", Value: int(1)}},
		}

	case "2":
		action = Action{
			Name:       "action" + id,
			Parameters: []ActionParameter{{Name: "param", Value: "2"}},
		}

	case "3":
		action = Action{
			Name:       "action" + id,
			Parameters: []ActionParameter{{Name: "param1", Value: 3.3}, {Name: "param2", Value: true}},
		}

	default:
		action = Action{
			Name: "action" + id,
		}
	}

	jsonAction, err := json.Marshal(action)
	if err != nil {
		log.Println("marshal:", err)
		return
	}

	err = c.WriteMessage(websocket.TextMessage, []byte(jsonAction))
	if err != nil {
		log.Println("write:", err)
		return
	}

	log.Printf("sent: %s", jsonAction)

	// messageType int, message []byte, err error
	messageType, message, err := c.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		return
	}

	log.Printf("recv: %s (type=%d)", message, messageType)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	clients = make([]WebSocketStore, 0)
	go func() {
		reader := bufio.NewReader(os.Stdin)

		for {

			for {
				if len(clients) > 0 {
					break
				}
				fmt.Println("No client for now. Press [enter] to refresh")
				reader.ReadString('\n')
			}

			fmt.Println("List of the clients:")
			for i, v := range clients {
				fmt.Printf("%d- %s\n", i, v.token)
			}
			fmt.Print("Choose a client: ")
			var choice int
			fmt.Scanf("%d", &choice)

			sendAction(clients[choice].socket)
		}
	}()
	http.HandleFunc("/action", handleActionFunc)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
