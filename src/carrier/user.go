package carrier

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Intelity/go-socket.io"
	"strconv"
	"time"
)

func FindUserBySocket(ns *socketio.NameSpace) (*User, error) {
	var user = User{}

	if _, found := SocketsMap[ns]; !found {
		return nil, errors.New(fmt.Sprintf("(UserBySocket) User not found for socket session (%s)", ns.Session.SessionId))
	}
	uid := SocketsMap[ns]

	query := DB.Find(&user, uid)

	if query.Error != nil {
		return &user, errors.New(fmt.Sprintf("(UserBySocket) User not found in database (%s)", query.Error))
	}
	return &user, nil
}

func FindSocketByUserId(user_id int) (*map[*socketio.NameSpace]bool, error) {
	var sockets = map[*socketio.NameSpace]bool{}

	if sockets, found := UsersMap[user_id]; !found {
		return &sockets, errors.New(fmt.Sprintf("(SocketByUser) Not found for uid (%d)", user_id))
	}
	return &sockets, nil
}

type User struct {
	ID          int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	VerifiedAt  time.Time
	DestroyAt   time.Time
	Filter      string
	Role        string
	Banned      bool
	Vip         bool
	InFilter    string
	Name        string
	Gender      int
	DateOfBirth time.Time
	BannedTo    time.Time
	Real        bool

	CurrentProfile   Profile
	CurrentProfileId sql.NullInt64
	Profiles         []Profile
}

func (u *User) LastProfile() *Profile {
	profile := &Profile{}
	DB.Model(&u).Related(&Profile{}).Order("created_at DESC").First(&profile)
	return profile
}

func (u *User) GetInCallCount() int {
	return 0
}

func (u *User) SendCallConnect(call *Call) error {
	return nil
}

func (u *User) InCall() bool {
	return false
}

func (u *User) SetOnline(ns *socketio.NameSpace) error {
	if _, found := SocketsMap[ns]; found {
		return errors.New("User on that socket is already connected.")
	}

	SocketsMap[ns] = u.ID
	if _, found := UsersMap[u.ID]; !found {
		UsersMap[u.ID] = make(map[*socketio.NameSpace]bool)
	}
	UsersMap[u.ID][ns] = true

	defer func() {
		count := Redis.SCard("users:online").Val()
		Redis.SAdd("users:online:peaks", strconv.FormatInt(count, 10))
	}()

	pipeline := Redis.Pipeline()
	pipeline.SAdd("users:online", strconv.Itoa(u.ID))
	pipeline.SAdd("users:online:today", strconv.Itoa(u.ID))
	pipeline.SAdd("users:reports:cleanup", strconv.Itoa(u.ID))
	pipeline.HIncrBy("users:sessions", strconv.Itoa(u.ID), 1)
	pipeline.HSetNX("users:online:from", strconv.Itoa(u.ID), strconv.FormatInt(time.Now().Unix(), 10))

	pipeline.Exec()

	return nil
}

func (u *User) SetOffline(ns *socketio.NameSpace) error {
	pipeline := Redis.Pipeline()

	pipeline.SRem("users:online", strconv.Itoa(u.ID))
	pipeline.SAdd("users:reports:cleanup", strconv.Itoa(u.ID))
	online_since := pipeline.HGet("users:online:from", strconv.Itoa(u.ID))
	pipeline.HDel("users:online:from", strconv.Itoa(u.ID))
	pipeline.Exec()

	start, err := strconv.ParseInt(online_since.Val(), 0, 64)
	if err != nil {
		return err
	}

	finish := time.Now().Unix()

	Redis.HIncrBy("users:online:time", strconv.Itoa(u.ID), finish-start)

	defer delete(SocketsMap, ns)
	defer delete(UsersMap[u.ID], ns)

	return nil
}
