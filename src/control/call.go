package control

import (
	"errors"
	_ "fmt"
	"time"
)

const (
	statusCalling      = 1
	statusActive       = 2
	statusDisconnected = 8
	statusFinished     = 9
	statusSkipped      = 10
	statusDeleted      = 11
	statusRejected     = 12
	statusFailed       = 13
	statusEscaped      = 14
)

func (c *Call) callLimitReached() error {
	var callCount int
	_db.Table("calls").Select("COUNT(id) as today_calls").Where("source_id = ?", c.Source.ID).Where("status = ?", statusFinished).Pluck("today_calls", &callCount)
	if callCount > 5 && !c.Source.Vip {
		return errors.New("Source call limit reached")
	}
	return nil
}

func (c *Call) participantsAvailable() bool {
	if userOnline(c.Destination) {
		return false
	}

	if userOnline(c.Source) {
		return false
	}

	if userInCall(c.Destination) {
		return false
	}
	return true
}

func (c *Call) finishCall(status int) {

}

func (c *Call) Initialize(call_id int) error {

	query := _db.Find(&c, call_id)

	if query.Error != nil {
		return query.Error
	}

	_db.Model(c).Related(&c.Destination, "destination_id")
	_db.Model(c).Related(&c.Source, "source_id")

	if err := c.callLimitReached(); err != nil {
		return err
	}

	if !c.participantsAvailable() {
		c.finishCall(13)
	}

	makeCallConnect(c.Source, *c)
	makeCallConnect(c.Destination, *c)

	time.AfterFunc(time.Second*40, func() {
	})

	return nil
}
