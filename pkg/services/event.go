package services

import (
	"database/sql"
	errors2 "events/pkg/errors"
	"events/pkg/events"
	rand2 "events/pkg/rand"
	"events/pkg/storage/postgres"
	"fmt"
	"strings"
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

func (s *EventService) CreateEvent(e *events.Event, emails []string, coverImage []byte, coverImageExt string, uid int) (int,  error) {
	const op = "eventService.CreateEvent"

	tx, err := s.r.Tx()
	if err != nil {
		return 0, errors2.Wrap(err, op, "obtaining tx")
	}

	// only saves meta data
	id, err := s.r.SaveEventTx(tx, e, uid)
	if err != nil {
		tx.Rollback()
		return 0, errors2.Wrap(err, op, "saving event via repo")
	}

	e.ID = id

	// updates the data
	err = s.updateEventTx(tx, e, emails, coverImage, coverImageExt)
	if err != nil {
		_ = tx.Rollback()
		return 0, errors2.Wrap(err, op, "updating event via repo")
	}

	tx.Commit()

	return id, nil
}

func (s *EventService) UpdateEvent(e *events.Event) error {
	const op = "eventService.Event"

	clean, err := s.Event(e.ID)
	if err != nil {
		return errors2.Wrap(err, op,"getting event")
	}

	clean.Title = e.Title
	clean.Description = e.Description
	clean.Link = e.Link
	clean.StartTime = e.StartTime
	clean.EndTime = e.EndTime

	err = s.r.UpdateEvent(e)
	return errors2.Wrap(err, op, "updating event via repo")
}

func (s *EventService) updateEventTx(tx *sql.Tx, e *events.Event, emails []string, coverImage []byte, coverImageExt string) error {
	const op = "eventService.updateEventTx"

	savedEvent, err := s.r.EventTx(tx, e.ID)
	if err != nil {
		return errors2.Wrap(err, op, "getting event")
	}

	savedEvent.Title = 			e.Title
	savedEvent.Description = 	e.Description
	savedEvent.Link = 			e.Link
	savedEvent.StartTime = 		e.StartTime
	savedEvent.EndTime = 		e.EndTime
	savedEvent.WelcomeMessage = e.WelcomeMessage
	savedEvent.IsPublished = 	e.IsPublished

	// add cover image path if cover image exist
	var key []string
	if coverImage != nil {
		randStrs := rand2.RandString(32)
		key = []string{randStrs[:2], randStrs[2:4], randStrs[4:]}
		savedEvent.CoverImagePath = strings.Join(key, "/") + "." + coverImageExt
	}

	err = s.r.UpdateEventTx(tx, savedEvent)
	if err != nil {
		return errors2.Wrap(err, op, "updating event via repo")
	}

	var ii []events.Invitation
	for _, e := range emails {
		i := events.Invitation{
			Email:        e,
			Token:        rand2.RandString(32),
		}
		ii = append(ii, i)
	}

	err = s.r.DeleteAllEventInvitationsTx(tx, e.ID)
	if err != nil {
		return errors2.Wrap(err, op, "deleting all invitations")
	}

	err = s.r.SaveEventInvitationsTx(tx, e.ID, ii)
	if err != nil {
		return errors2.Wrap(err, op, "saving invitations")
	}

	if coverImage != nil {
		err = s.r.SaveEventCoverImage(coverImage, key, coverImageExt)
		if err != nil {
			return errors2.Wrap(err, op, "updating event via repo")
		}
	}

	return nil
}
