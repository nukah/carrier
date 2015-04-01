package control

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

const (
	RevealTimeout   = 30
	AnswerTimeout   = 30
	CallDuration    = 180
	CallInitTimeout = 40

	StatusCalling      = 1
	StatusActive       = 2
	StatusDisconnected = 8
	StatusFinished     = 9
	StatusSkipped      = 10
	StatusDeleted      = 11
	StatusRejected     = 12
	StatusFailed       = 13
	StatusEscaped      = 14
)

func (c *Call) callLimitReached() bool {
	var callCount int
	this.db.Table("calls").Where("source_id = ?", c.Source.ID).Where("status = ?", StatusFinished).Count(&callCount)
	if callCount > 5 && !c.Source.Vip {
		return true
	}
	return false
}

func (c *Call) Init() {
	if c.callLimitReached() {
		makeCallStop(c.Source, *c, "call_limit")
	}

	if !userOnline(c.Destination) {
		makeCallStop(c.Source, *c, "destination_offline")
		c.Finish(StatusFailed)
		return
	}

	if !userOnline(c.Source) {
		makeCallStop(c.Source, *c, "source_offline")
		c.Finish(StatusFailed)
		return
	}

	if userInCall(c.Destination) {
		makeCallStop(c.Source, *c, "destination_busy")
		c.Finish(StatusFailed)
		return
	}
	defer this.mutex.Unlock()

	this.mutex.Lock()

	this.calls[c.ID] = c

	pipeline := this.redis.Pipeline()

	pipeline.HSet(c.RedisKey(), "source_answer", "")
	pipeline.HSet(c.RedisKey(), "destination_answer", "")

	if c.Type != "Call::Deferred" {
		pipeline.HSet(c.RedisKey(), "source_accept", "true")
	} else {
		pipeline.HSet(c.RedisKey(), "source_accept", "")
	}

	pipeline.HSet(c.RedisKey(), "destination_accept", "")
	pipeline.HSet(c.RedisKey(), "source_reveal", "")
	pipeline.HSet(c.RedisKey(), "destination_reveal", "")
	pipeline.Expire(c.RedisKey(), time.Second*CallDuration*2)
	pipeline.Exec()

	c.Connect()
}

func (c *Call) Connect() error {
	makeCallConnect(c.Source, *c)
	makeCallConnect(c.Destination, *c)
	c.callTimer = *time.AfterFunc(time.Second*CallInitTimeout, func() {
		callInfo := this.redis.HGetAllMap(c.RedisKey()).Val()
		if callInfo["source_accept"] == "" || callInfo["destination_accept"] == "" {
			makeCallStop(c.Source, *c, "destination_timeout")
			c.Finish(StatusSkipped)
			return
		}
	})
	return nil
}

func (c *Call) Finish(status int) {
	defer func() {
		recover()
		this.mutex.Lock()
		delete(this.calls, c.ID)
		log.Printf("(Call) Finishing call %d with status (%d)", c.ID, status)
		this.mutex.Unlock()
	}()
	switch status {
	case StatusFinished:
		this.db.Model(&c).Updates(Call{FinishedAt: time.Now()})
	case StatusRejected:
		this.db.Model(&c).Updates(Call{RejectedAt: time.Now()})
	}
	c.callTimer.Stop()
	this.db.Model(&c).Updates(Call{Status: status})
}

func (c *Call) Accept(user_id int, decision bool) error {
	var role string
	switch user_id {
	case c.Source.ID:
		role = "source_accept"
		c.SourceAccept = decision
		this.db.Save(&c)
	case c.Destination.ID:
		role = "destination_accept"
		c.DestinationAccept = decision
		this.db.Save(&c)
	default:
		return errors.New("Unknown role for user!")
	}
	this.redis.HSet(c.RedisKey(), role, strconv.FormatBool(decision))

	sourceAccept := this.redis.HGet(c.RedisKey(), "source_accept").Val()
	destinationAccept := this.redis.HGet(c.RedisKey(), "destination_accept").Val()

	if sourceAccept == "true" && destinationAccept == "true" {
		c.Start()
		return nil
	}

	if sourceAccept == "false" || destinationAccept == "false" {
		makeCallStop(c.Source, *c, "call_rejected")
		makeCallStop(c.Destination, *c, "call_rejected")
		c.Finish(StatusRejected)
		return nil
	}
	return nil
}

func (c *Call) Start() error {
	makeCallStart(c.Source, *c)
	makeCallStart(c.Destination, *c)
	c.Status = StatusActive
	c.AcceptedAt = time.Now()
	this.db.Save(&c)

	c.callTimer = *time.AfterFunc(time.Second*CallDuration, func() {
		c.Finish(StatusFinished)
		makeCallFinish(c.Source, *c)
		makeCallFinish(c.Destination, *c)
		if c.Incognito {
			c.StartReveal()
		} else {
			c.StartAnswer()
		}
	})
	return nil
}

func (c *Call) Stop(user User) error {
	var reason string
	switch c.Status {
	case StatusCalling:
		defer c.Finish(StatusEscaped)
		switch user.ID {
		case c.SourceID:
			reason = "source_cancelled"
		case c.DestinationID:
			reason = "destination_cancelled"
		}
	case StatusActive:
		defer c.Finish(StatusDisconnected)
		switch user.ID {
		case c.SourceID:
			reason = "source_disconnected"
		case c.DestinationID:
			reason = "destination_disconnected"
		}
	}

	makeCallStop(c.Source, *c, reason)
	makeCallStop(c.Destination, *c, reason)
	return nil
}

func (c *Call) StartReveal() error {
	makeCallAftermath(c.Source, *c, "call_reveal", "start")
	makeCallAftermath(c.Destination, *c, "call_reveal", "start")
	c.callTimer = *time.AfterFunc(time.Second*RevealTimeout, func() {
		makeCallAftermath(c.Source, *c, "call_reveal", "finish")
		makeCallAftermath(c.Destination, *c, "call_reveal", "finish")
		pipeline := this.redis.Pipeline()

		sourceReveal := pipeline.HGet(c.RedisKey(), "source_reveal").Val()
		destinationReveal := pipeline.HGet(c.RedisKey(), "destination_reveal").Val()

		pipeline.Exec()

		if (sourceReveal == "" && destinationReveal == "") || (sourceReveal == "false" || destinationReveal == "false") {
			makeCallReveal(c.Source, *c, false)
			makeCallReveal(c.Destination, *c, false)
		}

		if sourceReveal == "true" && destinationReveal == "true" {
			makeCallReveal(c.Source, *c, true)
			makeCallReveal(c.Destination, *c, true)
			c.StartAnswer()
		}
	})
	return nil
}

func (c *Call) StartAnswer() error {
	makeCallAftermath(c.Source, *c, "call_answer", "start")
	makeCallAftermath(c.Destination, *c, "call_answer", "start")
	c.callTimer = *time.AfterFunc(time.Second*AnswerTimeout, func() {
		makeCallAftermath(c.Source, *c, "call_answer", "finish")
		makeCallAftermath(c.Destination, *c, "call_answer", "finish")

		pipeline := this.redis.Pipeline()

		sourceAnswer := pipeline.HGet(c.RedisKey(), "source_answer").Val()
		destinationAnswer := pipeline.HGet(c.RedisKey(), "destination_answer").Val()

		pipeline.Exec()

		if (sourceAnswer == "" && destinationAnswer == "") || (sourceAnswer == "false" || destinationAnswer == "false") {
			makeCallAnswer(c.Source, *c, false)
			makeCallAnswer(c.Destination, *c, false)
		}

		if sourceAnswer == "true" && destinationAnswer == "true" {
			makeCallAnswer(c.Source, *c, true)
			makeCallAnswer(c.Destination, *c, true)
		}
	})
	return nil
}

func (c *Call) Find(id string) error {
	query := this.db.Find(&c, id)

	if query.Error != nil {
		return query.Error
	}
	this.db.Model(c).Related(&c.Destination, "destination_id")
	this.db.Model(c).Related(&c.Source, "source_id")

	return nil
}

func (c *Call) RedisKey() string {
	return fmt.Sprintf("fleet:calls:%d", c.ID)
}
