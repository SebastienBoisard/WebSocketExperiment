package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

// Upgrader specifies parameters for upgrading an HTTP connection to a WebSocket connection.
// Without parameters, default options will be applied (ReadBufferSize WriteBufferSize are set to 4096
var upgrader = websocket.Upgrader{} // use default options

// Hub is the link between a worker and the server
type Hub struct {
	token      string
	connection *websocket.Conn
	toSend     chan []byte
	actions    actionMap
}

type actionMap struct {
	sync.RWMutex
	m map[string]chan *Action
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

	h := &Hub{token: t, connection: c, toSend: make(chan []byte), actions: actionMap{m: make(map[string]chan *Action)}}
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
			log.Println("Error while reading:", err)
			continue
		}

		if messageType != websocket.TextMessage {
			log.Println("received a non text message")
			continue
		}

		var action Action
		json.Unmarshal(message, &action)

		h.actions.Lock()
		h.actions.m[action.ID] <- &action
		h.actions.Unlock()
	}
}

func (h *Hub) createRequest() {

	startCreateRequest := time.Now()

	id := strconv.Itoa(rand.Intn(4))

	var parameters []ActionParameter
	var wantedResult string

	switch id {
	case "1":
		parameters = []ActionParameter{{Name: "param", Value: int(1)}}
		wantedResult = "result from action1 param=1"

	case "2":
		parameters = []ActionParameter{{Name: "param", Value: "2"}}
		wantedResult = "result from action2 param=2"

	case "3":
		parameters = []ActionParameter{{Name: "param1", Value: 3.3}, {Name: "param2", Value: true}}
		wantedResult = "result from action3 param1=3.3 param2=true"

	default:
		wantedResult = "action not found"
	}

	actionID := uuid.NewV4().String()

	actionChan := make(chan *Action)
	h.actions.Lock()
	h.actions.m[actionID] = actionChan
	h.actions.Unlock()

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

	//log.Printf("sent: %s", jsonAction)

	responseAction := <-actionChan

	if responseAction.ID != action.ID {
		log.Printf("error on action.ID\n")
		return
	}

	if responseAction.Result != wantedResult {
		log.Printf("error on result expected=%s but got=%s\n", wantedResult, responseAction.Result)
		return
	}

	h.actions.Lock()
	delete(h.actions.m, actionID)
	h.actions.Unlock()

	counter++

	elapsedTime := time.Now().Sub(startCreateRequest).Seconds()
	if elapsedTime > 1 {
		fmt.Printf("request action%s completed in %f\n", id, elapsedTime)
	}

	if counter%1000 == 0 {
		fmt.Printf("%d requests in %1.0f seconds\n", counter, time.Now().Sub(startTime).Seconds())
	}

	//log.Printf("recv: action[%s] name=%s  result=%s\n", responseAction.ID, responseAction.Name, responseAction.Result)
}

func (h *Hub) run() {

	for i := 0; i < 1000; i++ {
		go func() {
			for {
				go h.createRequest()

				time.Sleep(time.Duration(rand.Int63n(50)) * time.Millisecond)
			}
		}()
	}

	for {
		go h.createRequest()

		time.Sleep(time.Duration(rand.Int63n(50)) * time.Millisecond)
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

var counter int
var startTime = time.Now()

func startServer() {
	http.HandleFunc("/action", handleActionFunc)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
