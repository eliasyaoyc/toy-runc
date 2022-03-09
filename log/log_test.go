package log

import (
	"os"
	"testing"
)

func TestNewTee(t *testing.T) {
	file1, err := os.OpenFile("./access.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	file2, err := os.OpenFile("./error.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	var tops = []TeeOption{
		{
			W: file1,
			Lef: func(lvl Level) bool {
				return lvl <= InfoLevel
			},
		},
		{
			W: file2,
			Lef: func(lvl Level) bool {
				return lvl > InfoLevel
			},
		},
	}

	logger := NewTee(tops)
	ResetDefault(logger)

	Info("demo3:", String("app", "start ok"),
		Int("major version", 3))
	Error("demo3:", String("app", "crash"),
		Int("reason", -1))
}

func TestRotate(t *testing.T) {
	var tops = []TeeOption{
		{
			Filename: "access.log",
			Ropt: RotateOptions{
				MaxSize:    1,
				MaxAge:     1,
				MaxBackups: 3,
				Compress:   true,
			},
			Lef: func(lvl Level) bool {
				return lvl <= InfoLevel
			},
		},
		{
			Filename: "error.log",
			Ropt: RotateOptions{
				MaxSize:    1,
				MaxAge:     1,
				MaxBackups: 3,
				Compress:   true,
			},
			Lef: func(lvl Level) bool {
				return lvl > InfoLevel
			},
		},
	}

	logger := NewTeeWithRotate(tops)
	ResetDefault(logger)

	// 为了演示自动rotate效果，这里多次调用log输出
	for i := 0; i < 20000; i++ {
		Info("demo3:", String("app", "start ok"),
			Int("major version", 3))
		Error("demo3:", String("app", "crash"),
			Int("reason", -1))
	}
}
