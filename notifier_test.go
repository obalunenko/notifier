package notifier_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/obalunenko/notifier"
)

func newTestNotifier(t testing.TB, w io.Writer, name string) notifier.Notifier {
	t.Helper()

	n, err := notifier.NewIOWriterNotifier(w, name)
	require.NoError(t, err)

	return n
}

func TestMultiNotifier_Alert(t *testing.T) {
	ctx := context.Background()

	ctx = notifier.ContextWithMetadata(ctx, notifier.Metadata{
		AppName:      "test_app",
		InstanceName: "test_instance",
		Commit:       "test_commit",
		BuildDate:    time.Now().Format(time.DateTime),
	})

	bufOne := bytes.NewBuffer(nil)
	bufTwo := bytes.NewBuffer(nil)

	// Create a new multi notifier.
	ntfr, err := notifier.NewMultiNotifier(
		newTestNotifier(t, bufOne, "one"),
		newTestNotifier(t, bufTwo, "two"),
	)
	require.NoError(t, err)

	type args struct {
		severity notifier.Severity
		message  string
	}

	type testcase struct {
		name        string
		args        args
		wantMessage string
		wantErr     require.ErrorAssertionFunc
	}

	testcases := []testcase{
		{
			name: "valid parameters",
			args: args{
				severity: notifier.SeverityInfo,
				message:  "valid parameters",
			},
			wantMessage: "valid parameters",
			wantErr:     require.NoError,
		},
		{
			name: "empty message",
			args: args{
				severity: notifier.SeverityInfo,
				message:  "",
			},
			wantMessage: "",
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.EqualError(t, err,
					"send alert to 'multi[iowriter: one;iowriter: two]': format alert: message is empty")
			},
		},
		{
			name: "invalid severity",
			args: args{
				severity: notifier.Severity(100),
				message:  "invalid severity",
			},
			wantMessage: "", // No message should be sent.
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.EqualError(t, err,
					"send alert to 'multi[iowriter: one;iowriter: two]': format alert: 'Severity(100)', "+
						"should be one of '[INFO WARNING CRITICAL]': invalid severity")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				bufOne.Reset()
				bufTwo.Reset()
			})

			err = ntfr.Alert(ctx, tc.args.severity, tc.args.message)
			tc.wantErr(t, err)

			if tc.wantMessage != "" {
				require.Contains(t, bufOne.String(), tc.wantMessage, "first notifier")
				require.Contains(t, bufTwo.String(), tc.wantMessage, "second notifier")
			}
		})
	}
}
