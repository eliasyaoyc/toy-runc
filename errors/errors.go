package errors

import (
	oerrors "errors"
	"fmt"
)

func New(message string) error {
	return oerrors.New(message)
}

func Wrap(message string, err error) error {
	message += " :%w"
	return fmt.Errorf(message, err)
}

func Is(err, target error) bool {
	return oerrors.Is(err, target)
}

func Unwrap(err error) error {
	return oerrors.Unwrap(err)
}

func As(err error, target interface{}) bool {
	return oerrors.As(err, target)
}

func Catch(fs ...func() error) error {
	var err error
	for i := 0; i < len(fs); i++ {
		if e := fs[i](); e != nil {
			err = Wrap("", e)
		}
	}
	return err
}
