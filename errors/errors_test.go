package errors

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

func TestReadFileBase64JSON(t *testing.T) {
	if f1Err := f1(); f1Err != nil {
		println("")
	}
	if f2Err := f2(); f2Err != nil {
		println("")
	}

	m := Return("testdata/test.base64")
	m = Bind(m, ReadFile)
	m = Bind(m, BytesToStr)
	m = Bind(m, Base64DecodeString)
	m = Bind(m, JSONUnmarshal)
	jsonMap, err := m(nil)
	if err != nil {
		fmt.Println(err)
	}
	t.Log(jsonMap)
}

// Base64DecodeString reads v as string and returns Monad: error -> []byte, error
func Base64DecodeString(v Any) Monad {
	vString := v.(string)
	return func(s error) (Any, error) {
		return base64.StdEncoding.DecodeString(vString)
	}
}

// JSONUnmarshal reads v as []byte and returns Monad: error -> map[string]interface{}, error
func JSONUnmarshal(v Any) Monad {
	vBytes := v.([]byte)
	return func(s error) (Any, error) {
		resultMap := make(map[string]interface{})
		err := json.Unmarshal(vBytes, &resultMap)
		return resultMap, err
	}
}

// ReadFile reads v as string and returns Monad: error -> []byte, error
func ReadFile(filename Any) Monad {
	filenameString := filename.(string)
	return func(error) (Any, error) {
		return ioutil.ReadFile(filenameString)
	}
}

// BytesToStr reads v as []byte and returns Monad: error -> string, error
func BytesToStr(v Any) Monad {
	vBytes := v.([]byte)
	return func(error) (Any, error) {
		return string(vBytes), nil
	}
}
