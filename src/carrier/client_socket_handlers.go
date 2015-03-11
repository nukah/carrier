package carrier

import (
	_ "bytes"
	_ "encoding/base64"
	"github.com/Intelity/go-socket.io"
	_ "gopkg.in/vmihailenco/msgpack.v2"
	"log"
	_ "net/url"
	_ "strings"
)

func MessageHandler(ns *socketio.NameSpace, body string) {
}

func ContactHandler(ns *socketio.NameSpace, body string) {

}

func CallHandler(ns *socketio.NameSpace, body string) {

}

func NotificationHandler(ns *socketio.NameSpace, body string) {

}

func UserHandler(ns *socketio.NameSpace, body string) {

}

func SystemHandler(ns *socketio.NameSpace, body string) {

}

func BanHandler(ns *socketio.NameSpace, body string) {

}

func CallResultHandler(ns *socketio.NameSpace, body string) {

}

func CallStateHandler(ns *socketio.NameSpace, body string) {

}

func ClaimHandler(ns *socketio.NameSpace, body string) {

}

type APIRequest struct {
	EventAction, Event string
}

func AuthorizationHandler(ns *socketio.NameSpace, token string) {
	user := new(User)

	err := DB.Find(&user, token).Error
	if err != nil {
		log.Printf("(Authorization) DB Search error: %s", err)
	}

	SocketsMap[ns] = int(user.ID)

	if _, found := UsersMap[user.ID]; !found {
		UsersMap[user.ID] = make(map[*socketio.NameSpace]bool)
	}
	UsersMap[user.ID][ns] = true

	user.SetOnline()
}

func DisconnectionHandler(ns *socketio.NameSpace) {
	user, err := FindUserBySocket(ns)

	if err != nil {
		log.Println(err)
	}

	defer delete(SocketsMap, ns)
	defer delete(UsersMap[user.ID], ns)

	user.SetOffline()
}
