package main

import (
	"github.com/nukah/go-socket.io"
	"log"
	_ "time"
)

func main() {
	client, err := socketio.Dial("http://localhost:3001/")
	if err != nil {
		log.Panic(err)
	}

	client.On("connect", func(ns *socketio.NameSpace) {
		// 	time.AfterFunc(time.Second*11, func() {
		// 		ns.Emit("authorize", "5")
		// 	})
		// 	//ns.Emit("call_start", "1")
		ns.Emit("call_init", "1")
		// })

		// client.On("call", func(ns *socketio.NameSpace, event_type string, body string) {
		// 	log.Println(event_type)
		// 	log.Println(body)
		// })
	})
	client.Run()
}
