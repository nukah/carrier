package carrier

import (
	_ "database/sql"
	"errors"
	"fmt"
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

type Call struct {
	ID                  int
	SourceAnswer        bool
	DestinationAnswer   bool
	Incognito           bool
	SourceReveal        bool
	DestinationReveal   bool
	Status              int
	CreatedAt           time.Time
	UpdatedAt           time.Time
	FinishedAt          time.Time
	SourceAnswerAt      time.Time
	DestinationAnswerAt time.Time
	SourceRevealAt      time.Time
	DestinationRevealAt time.Time
	AcceptedAt          time.Time
	RejectedAt          time.Time
	Type                string
	Destination         User
	DestinationID       int
	SourceID            int
	Source              User
}

type CallEvent struct {
	Type           string
	CallId         int
	CallType       string
	CallStopReason string
	Source         int
	Destination    int
}

func (c *Call) CallLimitReached() error {
	if call_count := c.Destination.GetInCallCount(); call_count >= 5 {
		return errors.New("Destination call limit reached")
	}
	return nil
}

func (c *Call) ParticipantsAvailable() error {
	if c.Destination.Offline() {
		return errors.New("Destination user is offline!")
	}
	if c.Source.Offline() {
		return errors.New("Source user is offline!")
	}
	if c.Destination.InCall() || c.Source.InCall() {
		return errors.New("Participants in call already")
	}
	return nil
}

func (c *Call) FinishCall(status int) {

}

func (c *Call) Initiated() bool {
	return false
}

func (c *Call) Initialize(call_id int) error {

	query := DB.Find(&c, call_id)

	if query.Error != nil {
		return query.Error
	}

	if c.Initiated() {
		return errors.New("Call is already in process")
	}

	DB.Model(c).Related(&c.Destination, "destination_id")
	DB.Model(c).Related(&c.Source, "source_id")

	if err := c.CallLimitReached(); err != nil {
		return err
	}

	if err := c.ParticipantsAvailable(); err != nil {
		c.FinishCall(13)
		return err
	}

	if err := c.Destination.SendCallConnect(*c); err != nil {
		return err
	}
	time.AfterFunc(time.Second*40, func() {
	})

	if err := c.Source.SendCallConnect(*c); err != nil {
		return err
	}
	return nil
}
