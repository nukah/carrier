package carrier

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func (u *User) LastProfile() *Profile {
	profile := &Profile{}
	this.db.Model(&u).Related(&Profile{}).Order("created_at DESC").First(&profile)
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
		CallId:         call.ID,
		CallType:       call.Type,
		CallStopReason: "",
		Source:         call.Source.ID,
		Destination:    call.Destination.ID,
	}
	sessions, err := FindSocketByUserId(u.ID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return errors.New(fmt.Sprintf("User %d not found in any session", u.ID))
	}
	for session := range sessions {
		session.Emit("call", formatted_call)
	}
	return nil
}

func (u *User) SendCallStop(call Call, reason string) error {
	formatted_call := &callEvent{
		Type:           "stop",
		CallId:         call.ID,
		CallType:       call.Type,
		CallStopReason: reason,
		Source:         call.Source.ID,
		Destination:    call.Destination.ID,
	}
	sessions, err := FindSocketByUserId(u.ID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return nil
	}
	for session := range sessions {
		session.Emit("call", formatted_call)
	}
	return nil
}

func (u *User) SendCallStart(call Call) error {
	formatted_call := &callEvent{
		Type:           "start",
		CallId:         call.ID,
		CallType:       call.Type,
		CallStopReason: "",
		Source:         call.Source.ID,
		Destination:    call.Destination.ID,
	}
	sessions, err := FindSocketByUserId(u.ID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return nil
	}
	for session := range sessions {
		session.Emit("call", formatted_call)
	}
	return nil
}

func (u *User) SendCallFinish(call Call) error {
	formatted_call := &callEvent{
		Type:           "finish",
		CallId:         call.ID,
		CallType:       call.Type,
		CallStopReason: "",
		Source:         call.Source.ID,
		Destination:    call.Destination.ID,
	}

	sessions, err := FindSocketByUserId(u.ID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return errors.New(fmt.Sprintf("User %d not found in any session", u.ID))
	}
	for session := range sessions {
		session.Emit("call", formatted_call)
	}
	return nil
}

func (u *User) SendCallAnswer(call Call, decision bool) error {
	formatted_call := &callResultEvent{
		Type:     "answer",
		CallId:   call.ID,
		Decision: decision,
	}

	sessions, err := FindSocketByUserId(u.ID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return errors.New(fmt.Sprintf("User %d not found in any session", u.ID))
	}
	for session := range sessions {
		session.Emit("call", formatted_call)
	}
	return nil
}

func (u *User) SendCallAftermath(call Call, event_type string, action_type string) error {
	formatted_call := &callAftermathEvent{
		Type:   event_type,
		Action: action_type,
		CallId: call.ID,
	}

	sessions, err := FindSocketByUserId(u.ID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return errors.New(fmt.Sprintf("User %d not found in any session", u.ID))
	}
	for session := range sessions {
		session.Emit("call_aftermath", formatted_call)
	}
	return nil
}

func (u *User) SendCallReveal(call Call, decision bool) error {
	formatted_call := &callResultEvent{
		Type:     "reveal",
		CallId:   call.ID,
		Decision: decision,
	}

	sessions, err := FindSocketByUserId(u.ID)
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return errors.New(fmt.Sprintf("User %d not found in any session", u.ID))
	}
	for session := range sessions {
		session.Emit("call", formatted_call)
	}
	return nil
}

func (u *User) InCall() bool {
	var currentCalls int
	this.db.Table("calls").Where("source_id = ? OR destination_id = ? AND status = ?", u.ID, u.ID, 2).Count(&currentCalls)
	return currentCalls >= 1
}

func (u *User) SetOnline() error {
	defer func() {
		count := this.redis.SCard("users:online").Val()
		this.redis.SAdd("users:online:peaks", strconv.FormatInt(count, 10))
	}()

	this.db.Exec("DELETE FROM online_users WHERE ID = ?", u.ID)
	this.db.Exec("INSERT INTO online_users(id, active) VALUES(?, ?)", u.ID, true)

	pipeline := this.redis.Pipeline()
	pipeline.SAdd("users:online", string(u.ID))
	pipeline.SAdd("users:online:today", string(u.ID))
	pipeline.SAdd("users:reports:cleanup", string(u.ID))
	pipeline.HIncrBy("users:sessions", string(u.ID), 1)
	pipeline.HSetNX("users:online:from", string(u.ID), strconv.FormatInt(time.Now().Unix(), 10))

	pipeline.Exec()

	return nil
}

func (u *User) SetOffline() error {
	pipeline := this.redis.Pipeline()

	this.db.Exec("DELETE FROM online_users WHERE ID = ?", u.ID)

	pipeline.SRem("users:online", string(u.ID))
	pipeline.SAdd("users:reports:cleanup", string(u.ID))
	online_since := pipeline.HGet("users:online:from", string(u.ID))
	pipeline.HDel("users:online:from", string(u.ID))
	pipeline.Exec()

	start, err := strconv.ParseInt(online_since.Val(), 0, 64)
	if err != nil {
		return err
	}

	this.redis.HIncrBy("users:online:time", string(u.ID), time.Now().Unix()-start)

	return nil
}
