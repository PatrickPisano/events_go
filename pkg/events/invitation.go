package events

type Invitation struct {
	EventID int
	Email string
	HasResponded bool
	Token string
	Accepted bool
}
