package errors

import "fmt"

func Wrap(err error, op, msg string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s [%s]: [%w]", msg, op, err)
}
