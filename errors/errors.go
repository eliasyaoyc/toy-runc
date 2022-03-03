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
	for _, f := range fs {
		if e := f(); e != nil {
			err = Wrap("", e)
		}
	}
	return err
}

type Any interface{}

type Monad func(error) (Any, error)

func Return(v Any) Monad {
	return func(s error) (Any, error) {
		return v, s
	}
}

func Bind(m Monad, f func(Any) Monad) Monad {
	return func(s error) (Any, error) {
		newV, newS := m(s)
		if newS != nil {
			return nil, newS
		}
		return f(newV)(newS)
	}
}
