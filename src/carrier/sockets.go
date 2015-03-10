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

func APIHandler(ns *socketio.NameSpace, body string) {
}

func DisconnectionHandler(ns *socketio.NameSpace) {
	user, _ := FindUserBySocket(ns)
	user.SetOffline(ns)
}

func AuthorizationHandler(ns *socketio.NameSpace, token string) {
	// trimmed, _ := url.QueryUnescape(token)
	// data, err := base64.StdEncoding.DecodeString(strings.Join(strings.Split(trimmed, "\n"), ""))
	// if err != nil {
	// 	log.Fatal("Incoming cookie information is corrupted (decode step)")
	// }
	// buf := bytes.NewBuffer(data)
	// decoder := msgpack.NewDecoder(buf)
	// decoder.DecodeMapFunc = func(dec *msgpack.Decoder) (interface{}, error) {
	// 	n, err := dec.DecodeMapLen()
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	m := make(map[string]interface{}, n)
	// 	for i := 0; i < n; i++ {
	// 		mk, err := dec.DecodeString()
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		mv, err := dec.DecodeInterface()
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		m[mk] = mv
	// 	}
	// 	return m, nil
	// }
	// result, err := decoder.DecodeInterface()
	// if err != nil {
	// 	log.Fatal("Incoming cookie information is corrupted (unpack step)")
	// }
	user := &User{}

	//if q := DB.Find(&user, result["warden.user.api.key"][1:2]); q.Error != nil {
	if q := DB.Find(&user, token); q.Error != nil {
		log.Println("User with id(%s) from socket not found", token)
	}
	user.SetOnline(ns)
}
