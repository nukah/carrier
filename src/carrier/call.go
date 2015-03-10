package carrier

import (
	_ "database/sql"
	"errors"
	"log"
	"time"
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

func (c *Call) CallLimitReached() error {
	if call_count := c.Destination.GetInCallCount(); call_count >= 5 {
		return errors.New("Destination call limit reached")
	}
	return nil
}

func (c *Call) ParticipantsAvailable() error {
	if c.Destination.InCall() || c.Source.InCall() {
		return errors.New("Participants in call already")
	}
	return nil
}

func (c *Call) FinishCall(status int) {

}

func (c *Call) Initialize(call_id int) error {
	query := DB.Find(&c, call_id)
	if query.Error != nil {
		return query.Error
	}

	DB.Model(c).Related(&c.Destination, "destination_id")
	DB.Model(c).Related(&c.Source, "source_id")

	if err := c.CallLimitReached(); err != nil {
		log.Printf("Call limit reached for user_id: %d", c.Destination.ID)
	}

	if err := c.ParticipantsAvailable(); err != nil {
		log.Printf("Participants not ready for call")
		go c.FinishCall(13)
	}

	if err := c.Destination.SendCallConnect(c); err != nil {
		return err
	}
	if err := c.Source.SendCallConnect(c); err != nil {
		return err
	}
	return nil
}
