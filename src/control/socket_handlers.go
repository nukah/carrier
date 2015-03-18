package control

import (
	"github.com/googollee/go-socket.io"
	"log"
)

type APIRequest struct {
	EventAction, Event string
}

func ConnectHandler(ns *socketio.Socket) {
	// time.AfterFunc(time.Second*10, func() {
	// 	if ns.Session.Values["uid"] == nil {
	// 		ns.CloseConnection()
	// 	}
	// })
}

func CallInitHandler(ns *socketio.Socket, call_id string) {
	call := new(Call)
	call.Find(call_id)
	this.calls[call_id] = call

	pipeline := this.redis.Pipeline()

	pipeline.HSet(call.RedisKey(), "source_answer", "")
	pipeline.HSet(call.RedisKey(), "destination_answer", "")
	pipeline.HSet(call.RedisKey(), "source_accept", "")
	pipeline.HSet(call.RedisKey(), "destination_accept", "")
	pipeline.HSet(call.RedisKey(), "source_reveal", "")
	pipeline.HSet(call.RedisKey(), "destination_reveal", "")

	pipeline.Exec()

	err := call.Init()
	if err != nil {
		log.Printf("(Call: %d -> %d) Error on initiating call: %s", call.Source.ID, call.Destination.ID, err)
		return
	}
}
