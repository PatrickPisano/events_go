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

func (e *EventService) Events() ([]events.Event, error) {
	var ee []events.Event

	/*start := time.Now()
	end := start.Add(time.Hour * 24 * 7)*/

	/*for i := 0; i < 10; i++ {
		ee = append(ee, events.Event{
			ID:             i,
			Title:          fmt.Sprintf("Title %d", i),
			Description:    "Some description",
			IsVirtual:      false,
			Address:        "An address",
			Link:           "",
			NumberOfSeats:  100,
			StartTime:      &start,
			EndTime:        &end,
			WelcomeMessage: "",
			IsPublished:    false,
		})
	}*/

	ee, err := e.r.Events()
	if err != nil {
		return nil, err
	}

	return ee, nil
}

func (s *EventService) CreateEvent(e *events.Event) (int,  error) {
	id, err := s.r.SaveEvent(e)
	if err != nil {
		return 0, err
	}
	// todo:: implement logic

	return id, nil
}