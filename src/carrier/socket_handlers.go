package carrier

import (
	_ "bytes"
	_ "encoding/base64"
	"github.com/googollee/go-socket.io"
	_ "gopkg.in/vmihailenco/msgpack.v2"
	"log"
	_ "net/url"
	"time"
)

type APIRequest struct {
	EventAction, Event string
}

func ConnectHandler(ns socketio.Socket) {
	time.AfterFunc(time.Second*10, func() {
		checkSocketAuthorization(ns)
	})
	ns.Emit("test")
	log.Printf("(Connect) New client(%s) connected", ns.Id())
}

func AuthorizationHandler(ns socketio.Socket, token string) {
	user := new(User)
	log.Printf("Test")
	err := this.db.Find(&user, token).Error
	if err != nil {
		log.Printf("(Authorization) DB Search error: %s", err)
	}
	if _, found := UsersMap[user.ID]; !found {
		UsersMap[user.ID] = make(map[socketio.Socket]bool)
	}

	SocketsMap[ns] = int(user.ID)
	UsersMap[user.ID][ns] = true
	this.redis.HSet("formation:users", string(user.ID), this.id)
	user.SetOnline()
}

func DisconnectionHandler(ns socketio.Socket) {
	user, _ := FindUserBySocket(ns)

	defer delete(SocketsMap, ns)
	if user != nil {
		delete(UsersMap[user.ID], ns)
		this.redis.HDel("formation:users", string(user.ID))
		user.SetOffline()
	}
}

func CallAcceptHandler(ns socketio.Socket, call_id string, decision bool) {
	user, _ := FindUserBySocket(ns)
	controlCallAccept(*user, call_id, decision)
}
