package carrier

import (
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
