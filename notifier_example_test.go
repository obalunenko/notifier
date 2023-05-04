package notifier_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/obalunenko/notifier"
)

const (
	// Test telegram credentials.
	testTGTokenEnv  = "TEST_TELEGRAM_TOKEN"
	testTGChatIDEnv = "TEST_TELEGRAM_CHAT_ID"
)

// ExampleNotifier shows how to create a new notifier with a list of notifiers and send a message.
func ExampleNotifier() {
	ctx := context.Background()

	var buf bytes.Buffer

	// Declare list of notifiers.
	var notifiers []notifier.Notifier

	// Create a new io.Writer notifier to visualize alerts.
	wn, err := notifier.NewIOWriterNotifier(&buf)
	if err != nil {
		// Handle error in your way.
		panic(err)
	}

	notifiers = append(notifiers, wn)

	// Create a new telegram notifier if token and chatID env set.
	if token, chatID := os.Getenv(testTGTokenEnv), os.Getenv(testTGChatIDEnv); token != "" && chatID != "" {
		tgn, err := notifier.NewTelegram(token, chatID)
		if err != nil {
			// Handle error in your way.
			panic(err)
		}

		notifiers = append(notifiers, tgn)
	}

	// Add the notifier to the list of notifiers.
	n, err := notifier.NewMultiNotifier(notifiers...)
	if err != nil {
		// Handle error in your way.
		panic(err)
	}

	// Add notifier metadata to context.
	ctx = notifier.ContextWithMetadata(ctx, notifier.Metadata{
		AppName:      "test_app",
		InstanceName: "test_instance",
		Commit:       "test_commit",
		BuildDate:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.DateTime),
	})

	// Send alert.
	err = n.Alert(ctx, notifier.SeverityWarning, "[NOTIFIER_TEST]: example message")
	if err != nil {
		// Handle error in your way.
		panic(err)
	}

	fmt.Println(buf.String())
	// Output:
	// <b>⚠️ Severity:</b> WARNING
	// <b>Alert Message:</b> [NOTIFIER_TEST]: example message
	// <b>Meta:</b>
	//	• app_name: test_app
	//	• build_date: 2020-01-01 00:00:00
	//	• commit: test_commit
	//	• instance_name: test_instance
}
