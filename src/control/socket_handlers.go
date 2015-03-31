package control

import (
	"github.com/googollee/go-socket.io"
)

type APIRequest struct {
	EventAction, Event string
}

func ConnectHandler(ns socketio.Socket) {
	// time.AfterFunc(time.Second*10, func() {
	// 	if ns.Session.Values["uid"] == nil {
	// 		ns.CloseConnection()
	// 	}
	// })
}

func CallInitHandler(ns socketio.Socket, call_id string) {
	call := new(Call)
	call.Find(call_id)

	if call.Status != StatusCalling {
		return
	}
	call.Init()
}
