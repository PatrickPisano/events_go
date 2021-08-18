package events

type ErrNotFound struct{
	Err error
}

func (e *ErrNotFound) Error() string {
	return e.Err.Error()
}

type ErrConflict struct{
	Err error
}

func (e *ErrConflict) Error() string {
	return e.Err.Error()
}