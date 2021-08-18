package events

import "time"

type Event struct {
	ID             int
	Title          string
	Description    string
	Link           string
	StartTime      *time.Time
	EndTime        *time.Time
	CoverImagePath string
	WelcomeMessage string
	HostID		   int
	IsPublished    bool
}
