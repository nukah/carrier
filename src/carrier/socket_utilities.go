package carrier

import (
	"errors"
	"fmt"
	"github.com/googollee/go-socket.io"
)

func checkSocketAuthorization(ns socketio.Socket) {

}

func setSocketAuthorization(s socketio.Socket, uid string) {

}

func FindUserBySocket(ns socketio.Socket) (*User, error) {
	user := new(User)
	if _, found := SocketsMap[ns]; !found {
		return user, errors.New(fmt.Sprintf("(UserBySocket) User not found for socket session (%s)", ns.Id()))
	}

	query := this.db.Find(user, SocketsMap[ns])
	if query.Error != nil {
		return user, errors.New(fmt.Sprintf("(UserBySocket) User not found in database (%s)", query.Error))
	}
	return user, nil
}

func FindSocketByUserId(user_id int) (map[socketio.Socket]bool, error) {
	sockets := map[socketio.Socket]bool{}

	if sockets, found := UsersMap[user_id]; !found {
		return sockets, errors.New(fmt.Sprintf("(SocketByUser) Not found for uid (%d)", user_id))
	}
	return sockets, nil
}
