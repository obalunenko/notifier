package notifier

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mymmrac/telego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// fakeTGToken matches telego token format.
	fakeTGToken  = "123456:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghi"
	fakeTGChatID = "42"

	getMeOK         = `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test","username":"test_bot"}}`
	getMeNoUsername = `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test"}}`
	unauthorized    = `{"ok":false,"error_code":401,"description":"Unauthorized"}`
	notFound        = `{"ok":false,"error_code":404,"description":"Not Found"}`
	sendMessageOK   = `{"ok":true,"result":{"message_id":1,"date":0,"chat":{"id":42,"type":"private"}}}`
	chatNotFound    = `{"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`
)

// sentMessage is a sendMessage request received by the fake Telegram server.
type sentMessage struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// fakeTelegram emulates Telegram Bot API methods used by the notifier.
type fakeTelegram struct {
	srv  *httptest.Server
	sent []sentMessage
}

func newFakeTelegram(tb testing.TB, getMe, sendMessage apiResp) *fakeTelegram {
	tb.Helper()

	ft := &fakeTelegram{}

	mux := http.NewServeMux()
	mux.HandleFunc("/bot"+fakeTGToken+"/getMe", func(w http.ResponseWriter, _ *http.Request) {
		writeResponse(tb, w, getMe)
	})
	mux.HandleFunc("/bot"+fakeTGToken+"/sendMessage", func(w http.ResponseWriter, r *http.Request) {
		var msg sentMessage

		if err := json.NewDecoder(r.Body).Decode(&msg); err == nil {
			ft.sent = append(ft.sent, msg)
		}

		writeResponse(tb, w, sendMessage)
	})

	ft.srv = httptest.NewServer(mux)
	tb.Cleanup(ft.srv.Close)

	return ft
}

func (ft *fakeTelegram) newNotifier(tb testing.TB) Notifier {
	tb.Helper()

	n, err := newTelegram(fakeTGToken, fakeTGChatID, telego.WithAPIServer(ft.srv.URL))
	require.NoError(tb, err)

	return n
}

// apiResp is a response returned by the fake Telegram server.
type apiResp struct {
	status int
	body   string
}

func apiResponse(status int, body string) apiResp {
	return apiResp{status: status, body: body}
}

func writeResponse(tb testing.TB, w http.ResponseWriter, resp apiResp) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.status)

	if _, err := io.WriteString(w, resp.body); err != nil {
		tb.Errorf("write fake telegram response: %v", err)
	}
}

func Test_newTelegram(t *testing.T) {
	type args struct {
		token  string
		chatID string
	}

	tests := []struct {
		name     string
		getMe    apiResp
		args     args
		wantKind string
		wantErr  require.ErrorAssertionFunc
	}{
		{
			name:  "valid parameters",
			getMe: apiResponse(http.StatusOK, getMeOK),
			args: args{
				token:  fakeTGToken,
				chatID: fakeTGChatID,
			},
			wantKind: "telegram[test_bot]",
			wantErr:  require.NoError,
		},
		{
			name:  "bot without username",
			getMe: apiResponse(http.StatusOK, getMeNoUsername),
			args: args{
				token:  fakeTGToken,
				chatID: fakeTGChatID,
			},
			wantKind: "telegram",
			wantErr:  require.NoError,
		},
		{
			name:  "empty token",
			getMe: apiResponse(http.StatusOK, getMeOK),
			args: args{
				token:  "",
				chatID: fakeTGChatID,
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, ErrEmptyTelegramToken, i...)
			},
		},
		{
			name:  "empty chat id",
			getMe: apiResponse(http.StatusOK, getMeOK),
			args: args{
				token:  fakeTGToken,
				chatID: "",
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, ErrEmptyTelegramChatID, i...)
			},
		},
		{
			name:  "malformed token",
			getMe: apiResponse(http.StatusOK, getMeOK),
			args: args{
				token:  "invalid-token",
				chatID: fakeTGChatID,
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, ErrInvalidToken, i...)
			},
		},
		{
			name:  "unauthorized token",
			getMe: apiResponse(http.StatusUnauthorized, unauthorized),
			args: args{
				token:  fakeTGToken,
				chatID: fakeTGChatID,
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, ErrInvalidToken, i...)
			},
		},
		{
			name:  "not found token",
			getMe: apiResponse(http.StatusNotFound, notFound),
			args: args{
				token:  fakeTGToken,
				chatID: fakeTGChatID,
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, ErrInvalidToken, i...)
			},
		},
		{
			name:  "server error",
			getMe: apiResponse(http.StatusInternalServerError, ""),
			args: args{
				token:  fakeTGToken,
				chatID: fakeTGChatID,
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.Error(t, err, i...)
				require.NotErrorIs(t, err, ErrInvalidToken, i...)
			},
		},
		{
			name:  "invalid chat id",
			getMe: apiResponse(http.StatusOK, getMeOK),
			args: args{
				token:  fakeTGToken,
				chatID: "not-a-number",
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorContains(t, err, "parse telegram chatID", i...)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := newFakeTelegram(t, tt.getMe, apiResponse(http.StatusOK, sendMessageOK))

			got, err := newTelegram(tt.args.token, tt.args.chatID, telego.WithAPIServer(ft.srv.URL))
			tt.wantErr(t, err)

			if err != nil {
				return
			}

			assert.Equal(t, tt.wantKind, got.Kind())
		})
	}
}

func Test_telegramNotifier_Alert(t *testing.T) {
	type args struct {
		severity Severity
		message  string
	}

	tests := []struct {
		name        string
		sendMessage apiResp
		args        args
		wantSent    []sentMessage
		wantErr     require.ErrorAssertionFunc
	}{
		{
			name:        "message sent",
			sendMessage: apiResponse(http.StatusOK, sendMessageOK),
			args: args{
				severity: SeverityInfo,
				message:  "<b>test</b> & message",
			},
			wantSent: []sentMessage{
				{
					ChatID:    42,
					Text:      "<b>ℹ️ Severity:</b> INFO\n<b>Alert Message:</b> &lt;b&gt;test&lt;/b&gt; &amp; message",
					ParseMode: telego.ModeHTML,
				},
			},
			wantErr: require.NoError,
		},
		{
			name:        "api error",
			sendMessage: apiResponse(http.StatusBadRequest, chatNotFound),
			args: args{
				severity: SeverityWarning,
				message:  "warning message",
			},
			wantSent: []sentMessage{
				{
					ChatID:    42,
					Text:      "<b>⚠️ Severity:</b> WARNING\n<b>Alert Message:</b> warning message",
					ParseMode: telego.ModeHTML,
				},
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorContains(t, err, "send telegram message failed", i...)
				require.ErrorContains(t, err, "chat not found", i...)
			},
		},
		{
			name:        "empty message",
			sendMessage: apiResponse(http.StatusOK, sendMessageOK),
			args: args{
				severity: SeverityInfo,
				message:  "",
			},
			wantSent: nil,
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, ErrEmptyMessage, i...)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := newFakeTelegram(t, apiResponse(http.StatusOK, getMeOK), tt.sendMessage)

			n := ft.newNotifier(t)

			err := n.Alert(context.Background(), tt.args.severity, tt.args.message)
			tt.wantErr(t, err)

			assert.Equal(t, tt.wantSent, ft.sent)
		})
	}
}
