package main

import (
	"flag"
	"log"
)

var server = flag.Bool("server", false, "launch server")
var addr = flag.String("addr", "localhost:8080", "http service address")
var token = flag.String("token", "1234", "token to identify this client")


func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime)

	if *server == true {
		startServer()
	} else {
		startWorker()
	}
}

