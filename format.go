package notifier

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
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

// htmlEscaper escapes characters reserved by the Telegram HTML parse mode.
var htmlEscaper = strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")

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
		Message:  htmlEscaper.Replace(message),
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
