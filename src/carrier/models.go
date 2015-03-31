package carrier

import (
	"database/sql"
	"time"
)

type Profile struct {
	Name        string
	Gender      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DateOfBirth time.Time
	State       string
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

type Call struct {
	ID                  int
	SourceAnswer        bool
	DestinationAnswer   bool
	Incognito           bool
	SourceReveal        bool
	DestinationReveal   bool
	SourceAccept        bool
	DestinationAccept   bool
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
	callTimer           time.Timer
}
