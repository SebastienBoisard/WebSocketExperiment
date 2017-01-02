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

func execCommand1() string {
	return "result from command1"
}

func execCommand2() string {
	return "result from command2"
}

func execCommand3() string {
	return "result from command3"
}

func execCommand(funcName func() string) string {
	return funcName()
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {

	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime)

	commandMap := map[string]func() string{
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

		funcCommand := commandMap[command.Name]
		if funcCommand == nil {
			log.Println("Error: command name unknown [", command.Name, "]")
			command.Result = "command not found"
		} else {
			command.Result = execCommand(funcCommand)
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
