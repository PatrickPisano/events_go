package services

import (
	"events/pkg/events"
	"events/pkg/storage/postgres"
)

// NewEventService is a function that initializes a struct. Use "New".
func NewEventService(r *postgres.EventStorage) *EventService {
	return &EventService{r}
}

type EventService struct {
	r *postgres.EventStorage
}

func (s *EventService) Events() ([]events.Event, error) {
	ee, err := s.r.Events()
	if err != nil {
		return nil, err
	}

	return ee, nil
}

func (s *EventService) Event(id int) (*events.Event, error) {
	e, err := s.r.Event(id)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (s *EventService) CreateEvent(e *events.Event) (int,  error) {
	id, err := s.r.SaveEvent(e)
	if err != nil {
		return 0, err
	}
	// todo:: implement logic

	return id, nil
}