package notifier

//go:generate stringer -type Severity -output severity_string.go -trimprefix Severity -linecomment

// Severity represents the severity of an alert.
type Severity int

const (
	severityUnknown Severity = iota
	// SeverityInfo represents an info alert.
	SeverityInfo // INFO
	// SeverityWarning represents a warning alert.
	SeverityWarning // WARNING
	// SeverityCritical represents a critical alert.
	SeverityCritical // CRITICAL

	severitySentinel
)

// Valid checks if the severity is valid.
func (s Severity) Valid() bool {
	return s > severityUnknown && s < severitySentinel
}

var allowedSeverities = func() []string {
	allowed := make([]string, 0, severitySentinel)

	for i := SeverityInfo; i < severitySentinel; i++ {
		allowed = append(allowed, i.String())
	}

	return allowed
}()
