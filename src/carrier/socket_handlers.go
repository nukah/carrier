package carrier

import (
	_ "bytes"
	_ "encoding/base64"
	"github.com/googollee/go-socket.io"
	_ "gopkg.in/vmihailenco/msgpack.v2"
	"log"
	_ "net/url"
	"strconv"
	"time"
)

type APIRequest struct {
	EventAction, Event string
}

func ConnectHandler(ns socketio.Socket) {
	time.AfterFunc(time.Second*10, func() {
		if result := checkSocketAuthorization(ns); result != "" {
			log.Println(result)
		}
	})
}

func AuthorizationHandler(ns socketio.Socket, token string) {
	user := new(User)

	err := this.db.Find(&user, token).Error
	if err != nil {
		log.Printf("(Authorization) DB Search error: %s", err)
		return
	}
	setSocketAuthorization(ns, user)
}

func DisconnectionHandler(ns socketio.Socket) {
	user, err := FindUserBySocket(ns)
	if err == nil && user != nil && user.InCall() {
		var callId int
		//this.db.Table("calls").Where("source_id = ? or destination_id = ? and status = ?", user.ID, user.ID, 2)
		controlCallStop(*user)
	}
	if result := removeSocketAuthorization(ns); result != "" {
		log.Println(result)
	}
}

func CallAcceptHandler(ns socketio.Socket, call_id string, accept string) {
	user, _ := FindUserBySocket(ns)
	var decision = false
	switch accept {
	case "true":
		decision = true
	}
	callId, _ := strconv.Atoi(call_id)

	if result := controlCallAccept(*user, callId, decision); result != nil {
		log.Println("CallAccept error: ", result)
	}
}

func CallStopHandler(ns socketio.Socket, call_id string) {
	user, _ = FindUserBySocket(ns)
	callId, _ = strconv.Atoi(call_id)
	if result := controlCallStop(*user, callId); result != nil {
		log.Println("CallStop error: ", result)
	}
}
