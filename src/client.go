package main

import (
	"github.com/Intelity/go-socket.io"
	"log"
)

func main() {
	client, err := socketio.Dial("http://localhost:8080/")
	if err != nil {
		log.Panic(err)
	}

	client.On("connect", func(ns *socketio.NameSpace) {
		log.Println("Connection")
		log.Println("ID", ns.Session.SessionId)
		ns.Emit("authorize", "5")
	})

	client.Run()
}
