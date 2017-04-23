package gus

import "strings"

var (
	ErrNotAuth        = &NotAuthenticatedError{}
	ErrCantDeleteSelf = ErrInvalid("You can't delete yourself.")
)

type NotAuthenticatedError struct {
}

func (n *NotAuthenticatedError) Error() string {
	return "Not Authenticated"
}

func ErrNotFound() error {
	return &NotFoundError{}
}

type NotFoundError struct {
}

func (n *NotFoundError) Error() string {
	return "Not found"
}

func ErrInvalid(messages ...string) error {
	return &ValidationError{Messages: messages}
}

type ValidationError struct {
	Messages []string `json:"messages"`
}

func (v *ValidationError) Error() string {
	return "- " + strings.Join(v.Messages, "\n- ")
}
