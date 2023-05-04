package notifier

import "errors"

var (
	// ErrEmptyTelegramToken is returned when the telegram token is empty.
	ErrEmptyTelegramToken = errors.New("telegram token is empty")
	// ErrEmptyTelegramChatID is returned when the telegram chat id is empty.
	ErrEmptyTelegramChatID = errors.New("telegram chat id is empty")
	// ErrEmptyMessage is returned when the message is empty.
	ErrEmptyMessage = errors.New("message is empty")
	// ErrInvalidToken is returned when the telegram token is invalid.
	ErrInvalidToken = errors.New("invalid token")
	// ErrEmptyNotifiers is returned when the notifiers' list is empty.
	ErrEmptyNotifiers = errors.New("notifiers list is empty")
	// ErrInvalidSeverity is returned when the severity is invalid.
	ErrInvalidSeverity = errors.New("invalid severity")
)
