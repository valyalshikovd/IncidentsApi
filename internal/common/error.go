package common

type ErrCode string

const (
	CodeNotFound   ErrCode = "NOT_FOUND"
	CodeNotValid   ErrCode = "NOT_VALID"
	CodeIternalErr ErrCode = "INTERNAL_ERROR"
)

type Error struct {
	Code ErrCode
	Text string
}

func NewError(code ErrCode, text string) *Error {
	return &Error{Code: code, Text: text}
}

func (e *Error) Error() string {
	return e.Text
}
