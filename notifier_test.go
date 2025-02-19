package notifier_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/obalunenko/getenv"
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
					"send alert to 'multi[iowriter: one;iowriter: two]': format alert: message is empty",
					i)
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
						"should be one of '[INFO WARNING CRITICAL]': invalid severity", i)
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

func getEnv(tb testing.TB, key string) string {
	tb.Helper()

	v := getenv.EnvOrDefault(key, "")
	if v == "" {
		tb.Skipf("skip test because %s is empty", key)
	}

	return v
}

// TestNewTelegram tests NewTelegram function.
func TestNewTelegram(t *testing.T) {
	type args struct {
		token  string
		chatID string
	}

	type testcase struct {
		name    string
		args    args
		wantErr require.ErrorAssertionFunc
	}

	testcases := []testcase{
		{
			name: "valid parameters",
			args: args{
				token:  getEnv(t, testTGTokenEnv),
				chatID: getEnv(t, testTGChatIDEnv),
			},
			wantErr: require.NoError,
		},
		{
			name: "empty token",
			args: args{
				token:  "",
				chatID: getEnv(t, testTGChatIDEnv),
			},
			wantErr: require.Error,
		},
		{
			name: "empty chat id",
			args: args{
				token:  getEnv(t, testTGTokenEnv),
				chatID: "",
			},
			wantErr: require.Error,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := notifier.NewTelegram(tc.args.token, tc.args.chatID)

			tc.wantErr(t, err)
		})
	}
}
