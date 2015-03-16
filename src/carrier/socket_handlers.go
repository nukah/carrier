package carrier

import (
	_ "bytes"
	_ "encoding/base64"
	"github.com/nukah/go-socket.io"
	_ "gopkg.in/vmihailenco/msgpack.v2"
	"log"
	_ "net/url"
	_ "strings"
	"time"
)

type APIRequest struct {
	EventAction, Event string
}

func ConnectHandler(ns *socketio.NameSpace) {
	time.AfterFunc(time.Second*10, func() {
		checkSocketAuthorization(ns)
	})
	log.Printf("(Connect) New client(%s) connected", ns.Id())
}

func AuthorizationHandler(ns *socketio.NameSpace, token string) {
	user := new(User)

	err := this.db.Find(&user, token).Error
	if err != nil {
		log.Printf("(Authorization) DB Search error: %s", err)
	}
	if _, found := UsersMap[user.ID]; !found {
		UsersMap[user.ID] = make(map[*socketio.NameSpace]bool)
	}

	SocketsMap[ns] = int(user.ID)
	UsersMap[user.ID][ns] = true
	ns.Session.Values["uid"] = user.ID
	this.redis.HSet("formation:users", string(user.ID), this.id)
	user.SetOnline()
}

func DisconnectionHandler(ns *socketio.NameSpace) {
	user, _ := FindUserBySocket(ns)

	defer delete(SocketsMap, ns)
	if user != nil {
		delete(UsersMap[user.ID], ns)
		this.redis.HDel("formation:users", string(user.ID))
		user.SetOffline()
	}
}
