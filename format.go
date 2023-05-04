package notifier

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"text/template"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// alertData holds the alert information that will be sent in the Telegram message.
type alertData struct {
	Message  string
	Severity Severity
	Metadata map[string]string
}

var (
	//go:embed format.gohtml
	alertFormat string

	tplAlert = template.Must(
		template.New("alert").
			Funcs(template.FuncMap{"severityEmoji": severityEmoji}).
			Parse(alertFormat),
	)
)

const (
	emojiInfo     = "ℹ️"
	emojiWarning  = "⚠️"
	emojiCritical = "🚨"
)

var severityToEmoji = map[Severity]string{
	SeverityInfo:     emojiInfo,
	SeverityWarning:  emojiWarning,
	SeverityCritical: emojiCritical,
}

// severityEmoji returns the emoji for the given severity.
func severityEmoji(severity Severity) string {
	return severityToEmoji[severity]
}

// formatAlert formats the alert message using the Golang template.
func formatAlert(ctx context.Context, severity Severity, message string) (string, error) {
	if message == "" {
		return "", ErrEmptyMessage
	}

	if !severity.Valid() {
		return "", fmt.Errorf("'%s', should be one of '%v': %w", severity, allowedSeverities, ErrInvalidSeverity)
	}

	var buf bytes.Buffer

	ad := alertData{
		Message:  tgbotapi.EscapeText(tgbotapi.ModeHTML, message),
		Severity: severity,
		Metadata: nil,
	}

	m, ok := MetadataFromContext(ctx)
	if ok {
		ad.Metadata = m.toMap()
	}

	if err := tplAlert.Execute(&buf, ad); err != nil {
		return "", err
	}

	return buf.String(), nil
}
