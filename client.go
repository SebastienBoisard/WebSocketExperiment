// +build ignore

package main

import (
	"flag"
	"log"
	"net/url"
	"os"
   "fmt"
   "bufio"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {

	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime)


	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()


   reader := bufio.NewReader(os.Stdin)

	for {

      // Read a text from stdin
      text, _ := reader.ReadString('\n')

      err := c.WriteMessage(websocket.TextMessage, []byte(text))
      if err != nil {
         log.Println("write:", err)
         return
      }
	}
}
