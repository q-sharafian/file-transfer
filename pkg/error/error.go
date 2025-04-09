package error

import "fmt"

const SeparatorMsg = ":"

type ErrorCode any

type Error struct {
	message string
	code    ErrorCode
}

// Return all error message
func (e Error) Error() string {
	return fmt.Sprintf(e.message)
}

func (e *Error) AppendEnd(msg string) *Error {
	e.message = fmt.Sprintf("%s%s %s", e.message, SeparatorMsg, msg)
	return e
}

func (e *Error) AppendBegin(msg string) *Error {
	e.message = fmt.Sprintf("%s%s %s", msg, SeparatorMsg, e.message)
	return e
}

func (e *Error) SetCode(code ErrorCode) *Error {
	e.code = code
	return e
}

func (e *Error) GetCode() ErrorCode {
	return e.code
}

func NewError(msg string, code ErrorCode, args ...any) Error {
	return Error{
		message: fmt.Sprintf(msg, args...),
		code:    code,
	}
}

// Create an error and return pointer to that
func NewErrorP(msg string, code ErrorCode, args ...any) *Error {
	return &Error{
		message: fmt.Sprintf(msg, args...),
		code:    code,
	}
}

func NewErrorFmt(msg string, code ErrorCode, args ...any) Error {
	return Error{
		message: fmt.Sprintf(msg, args),
		code:    code,
	}
}

// A simple error implementation
type SError struct {
	message string
}

// Error returns the error message.
func (e *SError) Error() string {
	return e.message
}

// NewError creates a new simple error instance.
func NewSError(message string) *SError {
	return &SError{
		message: message,
	}
}
