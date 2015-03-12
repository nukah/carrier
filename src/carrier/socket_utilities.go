package carrier

import (
	"errors"
	"fmt"
	"github.com/nukah/go-socket.io"
	"log"
)

func checkSocketAuthorization(ns *socketio.NameSpace) {
	if ns.Session.Values["uid"] == nil {
		log.Printf("(SocketAuth) Disconnecting socket after 10 seconds without auth")
		ns.CloseConnection()
	}
}

func setSocketAuthorization(s *socketio.NameSpace, uid string) {

}

func FindUserBySocket(ns *socketio.NameSpace) (*User, error) {
	user := new(User)
	if _, found := SocketsMap[ns]; !found {
		return user, errors.New(fmt.Sprintf("(UserBySocket) User not found for socket session (%s)", ns.Id()))
	}

	query := DB.Find(user, SocketsMap[ns])
	if query.Error != nil {
		return user, errors.New(fmt.Sprintf("(UserBySocket) User not found in database (%s)", query.Error))
	}
	return user, nil
}

func FindSocketByUserId(user_id int) (map[*socketio.NameSpace]bool, error) {
	sockets := map[*socketio.NameSpace]bool{}

	if sockets, found := UsersMap[user_id]; !found {
		return sockets, errors.New(fmt.Sprintf("(SocketByUser) Not found for uid (%d)", user_id))
	}
	return sockets, nil
}
