package carrier

import (
	_ "encoding/base64"
	"encoding/json"
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
		checkSocketAuthorization(ns)
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
	if err == nil && user.ID != 0 && user.InCall() {
		var callId int
		this.db.Table("calls").Where("source_id = ? or destination_id = ? and status = ?", user.ID, user.ID, 2).Select("id").Row().Scan(&callId)
		if result := controlCallCancel(*user, callId); result != nil {
			log.Println("CallStop error: ", result)
		}
	}
	if result := removeSocketAuthorization(ns); result != nil {
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

func CallCancelHandler(ns socketio.Socket, call_id string) {
	user, _ := FindUserBySocket(ns)
	callId, _ := strconv.Atoi(call_id)
	if result := controlCallCancel(*user, callId); result != nil {
		log.Println("CallStop error: ", result)
	}
}

func MessageHandler(ns socketio.Socket, messageJson string) {
	//user, _ := FindUserBySocket(ns)
	log.Println("Message handler: ", messageJson)
	var message Message
	if err := json.Unmarshal([]byte(messageJson), &message); err != nil {
		log.Println("Error ", err)
	}
	log.Println("Message: ", message)
}
