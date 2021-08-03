package events

import "time"

type Event struct {
	ID             int
	Title          string
	Description    string
	IsVirtual      bool
	Address        string
	Link           string
	NumberOfSeats  int
	StartTime      time.Time
	EndTime        time.Time
	WelcomeMessage string
	HostID		   int
	IsPublished    bool
}
