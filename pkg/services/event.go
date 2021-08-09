package services

import (
	"events/pkg/errors"
	"events/pkg/events"
	"events/pkg/storage/postgres"
	"fmt"
)

// NewEventService is a function that initializes a struct. Use "New".
func NewEventService(r *postgres.EventStorage) *EventService {
	return &EventService{r}
}

type EventService struct {
	r *postgres.EventStorage
}

func (s *EventService) Events(uid int) ([]events.Event, error) {
	ee, err := s.r.Events(uid)
	if err != nil {
		return nil, err
	}

	return ee, nil
}

func (s *EventService) Event(id int) (*events.Event, error) {
	const op = "eventService.Event"
	e, err := s.r.Event(id)
	if err != nil {
		return nil, fmt.Errorf("getting events from repo [%s]: %w", op, err)
	}

	return e, nil
}

func (s *EventService) CreateEvent(e *events.Event, emails []string, uid int) (int,  error) {
	const op = "eventService.CreateEvent"

	id, err := s.r.SaveEvent(e, uid) // todo:: use transaction
	if err != nil {
		return 0, errors.Wrap(err, op, "saving event via repo")
	}

	err = s.r.SaveEventInvitations(id, emails)
	if err != nil {
		return 0, errors.Wrap(err, op, "saving batch  emails")
	}

	return id, nil
}

func (s *EventService) UpdateEvent(e *events.Event) error {
	const op = "eventService.Event"

	clean, err := s.Event(e.ID)
	if err != nil {
		return errors.Wrap(err, op,"getting event")
	}

	clean.Title = e.Title
	clean.Description = e.Description
	clean.Link = e.Link
	clean.StartTime = e.StartTime
	clean.EndTime = e.EndTime

	err = s.r.UpdateEvent(e)
	return errors.Wrap(err, op, "updating event via repo")
}