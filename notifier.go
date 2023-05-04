// Package notifier provides functionality to send notifications.
package notifier

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Notifier declares notifier contract.
type Notifier interface {
	// Alert sends a message to the notifier.
	Alert(ctx context.Context, severity Severity, message string) error
	// Kind returns the notifier kind.
	Kind() string
}

func isInvalidToken(err error) bool {
	return strings.Contains(err.Error(), "Not Found")
}

// telegramNotifier sends messages to a telegram chat.
type telegramNotifier struct {
	// Telegram chat id.
	chatID int64
	// Telegram client.
	client *tgbotapi.BotAPI
}

// NewTelegram returns a new telegram notifier.
func NewTelegram(token, chatID string) (Notifier, error) {
	if token == "" {
		return nil, ErrEmptyTelegramToken
	}

	if chatID == "" {
		return nil, ErrEmptyTelegramChatID
	}

	client, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		if isInvalidToken(err) {
			err = ErrInvalidToken
		}

		return nil, fmt.Errorf("create telegram client: %w", err)
	}

	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse telegram chatID: %w", err)
	}

	return &telegramNotifier{
		chatID: id,
		client: client,
	}, nil
}

// Kind returns the notifier kind.
func (t *telegramNotifier) Kind() string {
	kind := "telegram"

	uname := t.client.Self.UserName
	if uname == "" {
		return kind
	}

	return fmt.Sprintf("%s[%s]", kind, uname)
}

// Alert sends a message to the telegram chat.
func (t *telegramNotifier) Alert(ctx context.Context, severity Severity, message string) error {
	alert, err := formatAlert(ctx, severity, message)
	if err != nil {
		return fmt.Errorf("format alert: %w", err)
	}

	msg := tgbotapi.NewMessage(t.chatID, alert)
	// Telegram messages should be sent in HTML format.
	// https://confluence.softswiss.com/display/ADT/Alerting+notes
	msg.ParseMode = tgbotapi.ModeHTML

	_, err = t.client.Send(msg)
	if err != nil {
		return fmt.Errorf("send telegram message failed: %w", err)
	}

	return nil
}

// multiNotifier is a notifier that sends messages to multiple notifiers.
type multiNotifier []Notifier

// NewMultiNotifier returns a new multiNotifier notifier.
// Useful when it's needed to send messages to multiple telegram chats or other notifiers.
func NewMultiNotifier(notifiers ...Notifier) (Notifier, error) {
	if len(notifiers) == 0 {
		return nil, ErrEmptyNotifiers
	}

	return multiNotifier(notifiers), nil
}

func (m multiNotifier) Kind() string {
	kinds := make([]string, 0, len(m))

	for _, n := range m {
		kinds = append(kinds, n.Kind())
	}

	return fmt.Sprintf("multi[%s]", strings.Join(kinds, ";"))
}

// Alert sends a message to all notifiers.
func (m multiNotifier) Alert(ctx context.Context, severity Severity, message string) error {
	var errs error

	for _, notifier := range m {
		err := notifier.Alert(ctx, severity, message)
		if err != nil {
			if errors.Is(err, ErrEmptyMessage) || errors.Is(err, ErrInvalidSeverity) {
				// If the message is empty, there is no need to send it to other notifiers.
				return fmt.Errorf("send alert to '%s': %w", m.Kind(), err)
			}

			errs = errors.Join(errs, fmt.Errorf("send alert to '%s': %w", notifier.Kind(), err))
		}
	}

	return errs
}

// iowriterNotifier is a notifier that writes messages to io.Writer.
type iowriterNotifier struct {
	w    io.Writer
	kind string
}

// NewIOWriterNotifier creates a new notifier that writes messages to io.Writer.
// Useful for testing.
// If kind is not provided, "iowriter" is used.
func NewIOWriterNotifier(w io.Writer, kind ...string) (Notifier, error) {
	if w == nil {
		return nil, fmt.Errorf("io.Writer is nil")
	}

	n := &iowriterNotifier{
		w:    w,
		kind: "iowriter",
	}

	if len(kind) > 0 {
		n.kind += ": " + strings.Join(kind, " ")
	}

	return n, nil
}

// Kind returns the notifier kind.
func (n *iowriterNotifier) Kind() string {
	return n.kind
}

// Alert sends a message to the io.Writer.
func (n *iowriterNotifier) Alert(ctx context.Context, severity Severity, msg string) error {
	alert, err := formatAlert(ctx, severity, msg)
	if err != nil {
		return fmt.Errorf("format alert: %w", err)
	}

	_, err = fmt.Fprintln(n.w, alert)
	if err != nil {
		return fmt.Errorf("write alert: %w", err)
	}

	return nil
}
