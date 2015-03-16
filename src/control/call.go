package control

import (
	"fmt"
	"log"
	"time"
)

const (
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

func (c *Call) FinishCall(status int) {
	defer delete(this.calls, string(c.ID))
	defer func() {
		recover()
		log.Printf("(Call) Finishing call %d with status (%d)", c.ID, status)
	}()
	c.CallTimer.Stop()
}

func (c *Call) Init() error {
	if c.callLimitReached() {
		makeCallStop(c.Source, *c, "call_limit")
	}

	if !userOnline(c.Destination) {
		makeCallStop(c.Source, *c, "destination_offline")
		c.FinishCall(StatusFailed)
		return nil
	}

	if !userOnline(c.Source) {
		makeCallStop(c.Source, *c, "source_offline")
		c.FinishCall(StatusFailed)
		return nil
	}

	if userInCall(c.Destination) {
		makeCallStop(c.Source, *c, "destination_busy")
		c.FinishCall(StatusFailed)
		return nil
	}

	c.Connect()

	return nil
}

func (c *Call) Connect() error {
	go func() {
		makeCallConnect(c.Source, *c)
		makeCallConnect(c.Destination, *c)

		c.CallTimer = *time.AfterFunc(time.Second*40, func() {
			callInfo := this.redis.HGetAllMap(c.RedisKey()).Val()
			if callInfo["source_answer"] == "" || callInfo["destination_answer"] == "" {
				makeCallStop(c.Source, *c, "destination_timeout")
				c.FinishCall(StatusSkipped)
				return
			}
		})
	}()
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
