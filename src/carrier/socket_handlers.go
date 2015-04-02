package carrier

import (
	_ "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/googollee/go-socket.io"
	ms "github.com/mitchellh/mapstructure"
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

func MessageHandler(ns socketio.Socket, messageJson map[string]interface{}) {
	var message Message
	user, _ := FindUserBySocket(ns)
	ms.Decode(messageJson, &message)

	t := time.Now().UTC()

	message.Source = *user
	message.SourceId = user.ID
	message.Action = "send"
	message.CallId = user.GetActiveCallId()
	message.CreatedAt = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())

	rMessage, _ := json.Marshal(message)

	defer this.redis.LPush("messages", string(rMessage))

	if user.InCall() && message.Type == "call" {
		controlCallMessage(message)
		return
	}

	if message.Type == "contact" && message.DestinationId != 0 {
		carrierContactMessage(message)
		return
	}
}
