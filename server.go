package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"time"
)

// Upgrader specifies parameters for upgrading an HTTP connection to a WebSocket connection.
// Without parameters, default options will be applied (ReadBufferSize WriteBufferSize are set to 4096
var upgrader = websocket.Upgrader{} // use default options

type Hub struct {
	token      string
	connection *websocket.Conn
	toSend     chan []byte
    actionMap  map[string] chan *Action
}

var hubs []*Hub

func handleActionFunc(w http.ResponseWriter, r *http.Request) {

	t := r.URL.Query().Get("token")

	// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
	// c (i.e. Conn type) represents a WebSocket connection.
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()

	h := &Hub{token: t, connection: c, toSend: make(chan []byte), actionMap: make(map[string] chan *Action)}
	hubs = append(hubs, h)

	go h.sendRequest()

	go h.receiveResponse()

	h.run()
}

func (h *Hub) receiveResponse() {
	defer h.connection.Close()

	for {
		messageType, message, err := h.connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			continue
		}

		if messageType != websocket.TextMessage {
			log.Println("received a non text message")
			continue
		}

		var action Action
		json.Unmarshal(message, &action)

		h.actionMap[action.ID] <- &action
	}
}


func (h *Hub) createRequest() {
	id := strconv.Itoa(rand.Intn(4))

	var parameters []ActionParameter

	switch id {
	case "1":
		parameters = []ActionParameter{{Name: "param", Value: int(1)}}

	case "2":
		parameters = []ActionParameter{{Name: "param", Value: "2"}}

	case "3":
		parameters = []ActionParameter{{Name: "param1", Value: 3.3}, {Name: "param2", Value: true}}
	}

	actionID := uuid.NewV4().String()

	actionChan := make(chan *Action)
	h.actionMap[actionID] = actionChan

	action := Action{
		ID:         actionID,
		Name:       "action" + id,
		Parameters: parameters,
	}

	jsonAction, err := json.Marshal(action)
	if err != nil {
		log.Println("marshal:", err)
		return
	}

	h.toSend <- []byte(jsonAction)

	log.Printf("sent: %s", jsonAction)

	responseAction := <- actionChan


	log.Printf("recv: action[%s] name=%s  result=%s\n", responseAction.ID, responseAction.Name, responseAction.Result)

}

func (h *Hub) run() {

	for {

		go h.createRequest()

		time.Sleep(time.Duration(rand.Int63n(50))*time.Millisecond)
	}
}

func (h *Hub) sendRequest() {
	defer h.connection.Close()

	for {
		select {
		case msg, ok := <-h.toSend:
			// The boolean variable ok returned by a receive operator indicates whether the received value was sent
			// on the channel (true) or is a zero value returned because the channel is closed and empty (false).
			if ok == false {
				// TODO: what to do here ?
			}
			err := h.connection.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("write:", err)
			}
		}
	}
}

func startServer() {
	/*
		clients = make([]WebSocketStore, 0)
		actionMap = make(map[string]chan Action)
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

				go sendAction(clients[choice].socket)
			}
		}()
	*/
	http.HandleFunc("/action", handleActionFunc)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
