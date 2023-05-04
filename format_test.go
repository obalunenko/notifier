package notifier

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wantFromFile(tb testing.TB, path string) string {
	tb.Helper()

	f, err := os.ReadFile(path)
	require.NoError(tb, err)

	return string(f)
}

func Test_formatAlert(t *testing.T) {
	ctx := context.Background()

	setCtx := func(ctx context.Context, metadata *Metadata) context.Context {
		if metadata != nil {
			ctx = ContextWithMetadata(ctx, *metadata)
		}

		return ctx
	}

	type args struct {
		message  string
		severity Severity
	}

	tests := []struct {
		name     string
		metadata *Metadata
		ctx      context.Context
		args     args
		wantPath string
		wantErr  require.ErrorAssertionFunc
	}{
		{
			name:     "without metadata, with context",
			metadata: nil,
			ctx:      ctx,
			args: args{
				message:  "test message",
				severity: SeverityInfo,
			},
			wantPath: filepath.Join("testdata", "Test_formatAlert_without_metadata.golden"),
			wantErr:  require.NoError,
		},
		{
			name: "with metadata, with context",
			metadata: &Metadata{
				AppName:      "test_app",
				InstanceName: "test_instance",
				Commit:       "test_commit",
				BuildDate:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.DateTime),
			},
			ctx: ctx,
			args: args{
				message:  "test message",
				severity: SeverityCritical,
			},
			wantPath: filepath.Join("testdata", "Test_formatAlert_with_metadata.golden"),
			wantErr:  require.NoError,
		},
		{
			name: "metadata with missed app name, with context",
			metadata: &Metadata{
				AppName:      "",
				InstanceName: "test_instance",
				Commit:       "test_commit",
				BuildDate:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.DateTime),
			},
			ctx: ctx,
			args: args{
				message:  "test message",
				severity: SeverityWarning,
			},
			wantPath: filepath.Join("testdata", "Test_formatAlert_metadata_with_missed_app_name.golden"),
			wantErr:  require.NoError,
		},
		{
			name: "metadata with empty fields, with context",
			metadata: &Metadata{
				AppName:      "",
				InstanceName: "",
				Commit:       "",
				BuildDate:    "",
			},
			ctx: ctx,
			args: args{
				message:  "test message",
				severity: SeverityInfo,
			},
			wantPath: filepath.Join("testdata", "Test_formatAlert_metadata_with_empty_fields.golden"),
			wantErr:  require.NoError,
		},
		{
			name: "message is empty, with context",
			metadata: &Metadata{
				AppName:      "test_app",
				InstanceName: "test_instance",
			},
			ctx: ctx,
			args: args{
				message:  "",
				severity: SeverityInfo,
			},
			wantPath: filepath.Join("testdata", "Test_formatAlert_message_is_empty.golden"),
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.EqualError(t, err, ErrEmptyMessage.Error())
			},
		},
		{
			name: "metadata set, context is nil",
			metadata: &Metadata{
				AppName:      "test_app",
				InstanceName: "test_instance",
				Commit:       "test_commit",
				BuildDate:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.DateTime),
			},
			ctx: nil,
			args: args{
				message:  "test message",
				severity: SeverityCritical,
			},
			wantPath: filepath.Join("testdata", "Test_formatAlert_with_metadata.golden"),
			wantErr:  require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx = setCtx(tt.ctx, tt.metadata)

			got, err := formatAlert(ctx, tt.args.severity, tt.args.message)
			tt.wantErr(t, err)

			expected := wantFromFile(t, tt.wantPath)

			assert.Equal(t, expected, got)
		})
	}
}
