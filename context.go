package notifier

import (
	"context"
)

type ctxKey struct{}

// Metadata contains information about the application.
type Metadata struct {
	AppName      string
	InstanceName string
	Commit       string
	BuildDate    string
	Extra        map[string]string
}

func (m Metadata) toMap() map[string]string {
	metadataMap := make(map[string]string, 4)

	if m.AppName != "" {
		metadataMap["app_name"] = m.AppName
	}

	if m.InstanceName != "" {
		metadataMap["instance_name"] = m.InstanceName
	}

	if m.Commit != "" {
		metadataMap["commit"] = m.Commit
	}

	if m.BuildDate != "" {
		metadataMap["build_date"] = m.BuildDate
	}

	if m.Extra != nil {
		for k, v := range m.Extra {
			metadataMap[k] = v
		}
	}

	return metadataMap
}

// ContextWithMetadata returns a new context with the given metadata.
// If ctx is nil, context.Background() is used.
func ContextWithMetadata(ctx context.Context, metadata Metadata) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, ctxKey{}, &metadata)
}

// MetadataFromContext returns the metadata stored in ctx, if any.
func MetadataFromContext(ctx context.Context) (*Metadata, bool) {
	if ctx == nil {
		return nil, false
	}

	metadata, ok := ctx.Value(ctxKey{}).(*Metadata)
	if !ok || metadata == nil {
		return nil, false
	}
	return metadata, true
}
