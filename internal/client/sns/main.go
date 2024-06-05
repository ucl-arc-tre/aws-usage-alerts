package sns

import (
	"context"
	"errors"

	awsSNS "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/rs/zerolog/log"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
)

type Client struct {
	aws *awsSNS.Client
}

func New() *Client {
	return &Client{awsSNS.NewFromConfig(config.AWS())}
}

// Send an email with some content
func (c *Client) Send(content string) error {
	subject := "aws-usage-alerts: Notification"
	topicArn := config.TopicARN()
	if topicArn == "" {
		return errors.New("cannot send SNS message. TopicARN is unset")
	}
	_, err := c.aws.Publish(
		context.Background(),
		&awsSNS.PublishInput{
			Message:  &content,
			Subject:  &subject,
			TopicArn: &topicArn,
		},
	)
	if err != nil {
		log.Err(err).Str("topicARN", topicArn).Msg("Failed to send message")
		return err
	} else {
		log.Debug().Msg("Sent successfully")
		return nil
	}
}
