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
	user := new(User)
	if _, found := SocketsMap[ns]; !found {
		return user, errors.New(fmt.Sprintf("(UserBySocket) User not found for socket session (%s)", ns.Session.SessionId))
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
	formatted_call := &CallEvent{
		Type:           "connect",
		CallId:         u.ID,
		CallType:       call.Type,
		CallStopReason: "",
		Source:         call.Source.ID,
		Destination:    call.Destination.ID,
	}

	if sessions, err := FindSocketByUserId(u.ID); err == nil {
		for session := range sessions {
			go session.Emit("call", formatted_call)
		}
	} else {
		// RPC call for another carrier in formation
	}

	return nil
}

func (u *User) InCall() bool {
	return false
}

func (u *User) SetOnline() error {
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

func (u *User) SetOffline() error {
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

	Redis.HIncrBy("users:online:time", strconv.Itoa(u.ID), time.Now().Unix()-start)

	return nil
}