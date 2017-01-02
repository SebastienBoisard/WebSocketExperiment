// +build ignore

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/url"
	"reflect"

	"github.com/gorilla/websocket"
)

type Command struct {
	Name   string
	Result string
}

func execCommand1() string {
	return "result from command1"
}

func execCommand2() string {
	return "result from command2"
}

func execCommand3() string {
	return "result from command3"
}

func execCommand(commandMap map[string]interface{},
	commandName string,
	commandParameters ...interface{}) (string, error) {

   // Retreive the function from the map
	commandFunc := commandMap[commandName]
   // Test if the function was retreived 
	if commandFunc == nil {
		return "", errors.New("Unknown command name")
	}

   // func ValueOf(i interface{}) Value
   // ValueOf returns a new Value initialized to the concrete value stored in the interface i. 
	f := reflect.ValueOf(commandFunc)
   // NumIn returns a function type's input parameter count.
   // It panics if the type's Kind is not Func.
   // TODO: check if f is a function
	if len(commandParameters) != f.Type().NumIn() {
		return "", errors.New("Wrong number of parameters")
	}

	inputParameters := make([]reflect.Value, len(commandParameters))
	for k, param := range commandParameters {
		inputParameters[k] = reflect.ValueOf(param)
	}

   // func (v Value) Call(in []Value) []Value
   // Cf. https://golang.org/pkg/reflect/#Value.Call
	results := f.Call(inputParameters)

   // TODO: check if len(results) == 1, and if results[0] is really a string
	result := results[0].String()

	return result, nil
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {

	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime)

	commandMap := map[string]interface{}{
		"command1": execCommand1,
		"command2": execCommand2,
		"command3": execCommand3,
	}

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

		result, err := execCommand(commandMap, command.Name)
		if err != nil {
			log.Println("Error:  [", err, "]")
			command.Result = "command not found"
		} else {
			command.Result = result
		}

		jsonCommand, err := json.Marshal(command)
		if err != nil {
			log.Println("marshal:", err)
			return
		}

		err = c.WriteMessage(websocket.TextMessage, []byte(jsonCommand))
		if err != nil {
			log.Println("write:", err)
			continue
		}

	}
}
