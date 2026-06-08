package feedbook

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FirebasePushSender struct {
	client *messaging.Client
}

func NewFirebasePushSender(ctx context.Context, credentialsFile string, projectID string) (*FirebasePushSender, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}

	var config *firebase.Config
	if projectID != "" {
		config = &firebase.Config{ProjectID: projectID}
	}

	app, err := firebase.NewApp(ctx, config, opts...)
	if err != nil {
		return nil, err
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}
	return &FirebasePushSender{client: client}, nil
}

func (s *FirebasePushSender) Send(token string, title string, body string, data map[string]string) (string, error) {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				ChannelID: "feedbook_push",
			},
		},
	}
	return s.client.Send(context.Background(), message)
}
