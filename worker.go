package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"reflect"
	"strconv"

	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

func execAction1(param float64) string {
	time.Sleep(1000 * time.Millisecond)
	return "result from action1 param=" + strconv.FormatFloat(param, 'f', -1, 64)
}

func execAction2(param string) string {
	time.Sleep(2000 * time.Millisecond)
	return "result from action2 param=" + param
}

func execAction3(param1 float64, param2 bool) string {
	time.Sleep(4000 * time.Millisecond)
	return "result from action3 param1=" + strconv.FormatFloat(param1, 'f', -1, 64) + " param2=" + strconv.FormatBool(param2)
}

func execAction(actionMap map[string]interface{},
	actionName string,
	actionParameters []ActionParameter) (string, error) {

	// Retrieve the function from the map
	actionFunc := actionMap[actionName]
	// Test if the function was retrieved
	if actionFunc == nil {
		return "", errors.New("Unknown action name")
	}

	// func ValueOf(i interface{}) Value
	// ValueOf returns a new Value initialized to the concrete value stored in the interface i.
	f := reflect.ValueOf(actionFunc)

	// NumIn returns a function type's input parameter count.
	// It panics if the type's Kind is not Func.
	// TODO: check if f is a function
	if len(actionParameters) != f.Type().NumIn() {
		return "", errors.New("Wrong number of parameters")
	}

	var buffer bytes.Buffer
	inputParameters := make([]reflect.Value, len(actionParameters))
	for k, param := range actionParameters {
		buffer.WriteString(fmt.Sprintf("  parameter[%d]=%s", k, reflect.TypeOf(param.Value)))
		inputParameters[k] = reflect.ValueOf(param.Value)
	}

	// func (v Value) Call(in []Value) []Value
	// Cf. https://golang.org/pkg/reflect/#Value.Call
	results := f.Call(inputParameters)

	// TODO: check if len(results) == 1, and if results[0] is really a string
	result := results[0].String()

	return result, nil
}

/**
 * createWebSocketConnection establishes a WebSocket to a server with a specific url.
 * But the server can be down, so we have to retry to create the WebSocket every x second.
 * TODO: find if it's the best solution to create a WebSocket with retries
 */
func createWebSocketConnection(u url.URL) *websocket.Conn {

	for {
		c, httpResponse, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err == nil {
			return c

		}
		log.Println("dial: err=", err, "httpResponse=", httpResponse)
		time.Sleep(2 * time.Second)
	}
}

type Worker struct {
	wsUrl      url.URL
	connection *websocket.Conn
	toSend     chan []byte
	actionMap  map[string]interface{}
}

func (w *Worker) sendResponse() {
	defer w.connection.Close()

	for {
		select {
		case msg, ok := <-w.toSend:
			// The boolean variable ok returned by a receive operator indicates whether the received value was sent
			// on the channel (true) or is a zero value returned because the channel is closed and empty (false).
			if ok == false {
				// TODO: what to do here ?
			}
			err := w.connection.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("write:", err)
			}
		}
	}
}

func (w *Worker) receiveRequest() {
	defer w.connection.Close()

	for {

		messageType, message, err := w.connection.ReadMessage()
		if err != nil {
			log.Println("read: err=", err)
			// TODO: manage the different possible errors
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("IsUnexpectedCloseError: %v\n", err)
				w.connection = createWebSocketConnection(w.wsUrl)
			}
			continue
		}

		if messageType != websocket.TextMessage {
			log.Println("received a non text message")
			continue
		}

		go func() {
			var action Action
			json.Unmarshal(message, &action)

			log.Printf("received action[%s] name=%s\n", action.ID, action.Name)

			result, err := execAction(w.actionMap, action.Name, action.Parameters)
			if err != nil {
				log.Println("Error:  [", err, "]")
				action.Result = "action not found"
			} else {
				action.Result = result
			}

			jsonAction, err := json.Marshal(action)
			if err != nil {
				log.Println("marshal:", err)
				return
			}

			w.toSend <- []byte(jsonAction)
			log.Printf("sent action[%s] name=%s  result=%s\n", action.ID, action.Name, action.Result)
		}()
	}
}

func startWorker() {

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/action", RawQuery: "token=" + *token}
	log.Printf("connecting to %s", u.String())

	// Create a WebSocket to a server with a specific url.
	c := createWebSocketConnection(u)

	worker := Worker{
		wsUrl:      u,
		connection: c,
		toSend:     make(chan []byte),
		actionMap: map[string]interface{}{
			"action1": execAction1,
			"action2": execAction2,
			"action3": execAction3,
		},
	}

	go worker.sendResponse()

	worker.receiveRequest()
}
