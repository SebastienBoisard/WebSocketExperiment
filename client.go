package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/url"
	"reflect"
	"strconv"

	"github.com/gorilla/websocket"
)

func execAction1(param float64) string {
	return "result from action1 param=" + strconv.FormatFloat(param, 'f', -1, 64)
}

func execAction2(param string) string {
	return "result from action2 param=" + param
}

func execAction3(param1 float64, param2 bool) string {
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

	inputParameters := make([]reflect.Value, len(actionParameters))
	for k, param := range actionParameters {
		log.Printf("action.name=%s parameter[%d]=%s\n", actionName, k, reflect.TypeOf(param.Value))
		inputParameters[k] = reflect.ValueOf(param.Value)
	}

	// func (v Value) Call(in []Value) []Value
	// Cf. https://golang.org/pkg/reflect/#Value.Call
	results := f.Call(inputParameters)

	// TODO: check if len(results) == 1, and if results[0] is really a string
	result := results[0].String()

	return result, nil
}

var addr = flag.String("addr", "localhost:8080", "http service address")
var token = flag.String("token", "1234", "token to identify this client")

func main() {

	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime)

	actionMap := map[string]interface{}{
		"action1": execAction1,
		"action2": execAction2,
		"action3": execAction3,
	}

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/action", RawQuery: "token="+*token}
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

		var action Action
		json.Unmarshal(message, &action)

		log.Println("received ", action.Name)

		result, err := execAction(actionMap, action.Name, action.Parameters)
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

		err = c.WriteMessage(websocket.TextMessage, []byte(jsonAction))
		if err != nil {
			log.Println("write:", err)
			continue
		}

	}
}
