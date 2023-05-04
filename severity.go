package notifier

import "fmt"

// Severity represents the severity of an alert.
type Severity int

const (
	severityUnknown Severity = iota
	// SeverityInfo represents an info alert.
	SeverityInfo
	// SeverityWarning represents a warning alert.
	SeverityWarning
	// SeverityCritical represents a critical alert.
	SeverityCritical

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

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", s)
	}
}
