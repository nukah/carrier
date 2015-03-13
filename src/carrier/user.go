package carrier

import (
	"strconv"
	"time"
)

func (u *User) LastProfile() *Profile {
	profile := &Profile{}
	_db.Model(&u).Related(&Profile{}).Order("created_at DESC").First(&profile)
	return profile
}

func (u *User) Online() bool {
	if _, session := UsersMap[u.ID]; session {
		return true
	}
	return false
}

func (u *User) SendCallConnect(call Call) error {
	formatted_call := &callEvent{
		Type:           "connect",
		CallId:         u.ID,
		CallType:       call.Type,
		CallStopReason: "",
		Source:         call.Source.ID,
		Destination:    call.Destination.ID,
	}

	if sessions, err := FindSocketByUserId(u.ID); err == nil {
		for session := range sessions {
			go session.Emit("call", formatted_call.to_JSON())
		}
	}
	return nil
}

func (u *User) InCall() bool {
	return false
}

func (u *User) SetOnline() error {
	defer func() {
		count := _redis.SCard("users:online").Val()
		_redis.SAdd("users:online:peaks", strconv.FormatInt(count, 10))
	}()

	pipeline := _redis.Pipeline()
	pipeline.SAdd("users:online", string(u.ID))
	pipeline.SAdd("users:online:today", string(u.ID))
	pipeline.SAdd("users:reports:cleanup", string(u.ID))
	pipeline.HIncrBy("users:sessions", string(u.ID), 1)
	pipeline.HSetNX("users:online:from", string(u.ID), strconv.FormatInt(time.Now().Unix(), 10))

	pipeline.Exec()

	return nil
}

func (u *User) SetOffline() error {
	pipeline := _redis.Pipeline()

	pipeline.SRem("users:online", string(u.ID))
	pipeline.SAdd("users:reports:cleanup", string(u.ID))
	online_since := pipeline.HGet("users:online:from", string(u.ID))
	pipeline.HDel("users:online:from", string(u.ID))
	pipeline.Exec()

	start, err := strconv.ParseInt(online_since.Val(), 0, 64)
	if err != nil {
		return err
	}

	_redis.HIncrBy("users:online:time", string(u.ID), time.Now().Unix()-start)

	return nil
}
