package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestCatch(t *testing.T) {
	err := Catch(f1, f2)
	fmt.Println(err)
}

func f1() error {
	return errors.New("hello Error 1")
}

func f2() error {
	return errors.New("hello Error 2")
}
