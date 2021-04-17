package errors

import (
	"fmt"
	"strings"
)

type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (e Error) Error() (out string) {
	return e.Code + ": " + e.Message
}

func WithMessage(code, message string) Error {
	return Error{
		Code:    code,
		Message: message,
	}
}

func GetMessageOnly(in error) error {
	if f, ok := in.(Error); ok {
		return fmt.Errorf(f.Message)
	}
	return in
}

func GetCodeFromError(in error) string {
	if f, ok := in.(Error); ok {
		return f.Code
	}
	return in.Error()
}

func (e Error) AppendMessage(in ...string) Error {
	for i, s := range in {
		if i == 0 {
			e.Message = s
			continue
		}
		e.Message = strings.TrimSpace(e.Message + ". " + s)
	}
	return e
}

func (e Error) SetData(data interface{}) Error {
	e.Data = data
	return e
}

func IsEqual(a error, b error) bool {
	if a.Error() == b.Error() {
		return true
	}
	if GetCodeFromError(a) == GetCodeFromError(b) {
		return true
	}
	return false
}
