package email

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	snsClient "github.com/ucl-arc-tre/aws-cost-alerts/internal/client/sns"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/config"
	"github.com/ucl-arc-tre/aws-cost-alerts/internal/types"
)

const (
	minimumAlertDuration = 12 * time.Hour
)

type Controller struct {
	sns snsClient.Interface
}

func New() *Controller {
	return NewWithClient(snsClient.New())
}

func NewWithClient(client snsClient.Interface) *Controller {
	return &Controller{sns: client}
}

func (c *Controller) Send(state *types.StateV1alpha1, errors []error) {
	if state == nil {
		log.Error().Msg("State was unset. Cannot send emails")
		return
	}
	content := ""
	for group, cost := range state.GroupsUsage() {
		threshold := config.GroupThreshold(group)
		if threshold < 1e-15 {
			log.Warn().Any("group", group).Msg("Unset or zero threshold")
			continue
		}
		usagePercentage := float64(cost.Total().Dollars/threshold) * 100.0
		if usagePercentage > 100 {
			log.Info().Any("group", group).Float64("%", usagePercentage).Msg(">100% used")
			content += fmt.Sprintf("%v  %.2f%% \n ", group, usagePercentage)
		} else {
			log.Debug().Any("group", group).Float64("%", usagePercentage).Msg("Calculated usage")
		}
	}
	if len(errors) > 0 {
		content += "⚠️ Errors\n"
	}
	for _, err := range errors {
		content += fmt.Sprintf("%v\n", err)
	}
	if content != "" && shouldEmailAdmins(state) {
		if err := c.sns.Send(header + content); err == nil {
			setEmailSentNowForAdmins(state)
		} else {
			log.Err(err).Msg("Failed to send to SNS")
		}
	}
}

func shouldEmailAdmins(state *types.StateV1alpha1) bool {
	adminEmails := config.AdminEmails()
	if len(adminEmails) == 0 {
		log.Warn().Msg("No admin emails defined")
		return false
	}
	return shouldEmail(state, adminEmails[0])
}

func shouldEmail(state *types.StateV1alpha1, email types.EmailAddress) bool {
	sendAt, exists := state.EmailsSentAt[email]
	if !exists {
		log.Debug().Any("email", email).Msg("Have not yet emailed")
		return true
	} else {
		timeSinceAlert := time.Since(sendAt)
		log.Debug().
			Float64("hoursSinceAlert", timeSinceAlert.Hours()).
			Float64("minimumAlertDelayHours", minimumAlertDuration.Hours()).
			Msg("Calculated if an email should be sent")
		return timeSinceAlert > minimumAlertDuration
	}
}

func setEmailSentNowForAdmins(state *types.StateV1alpha1) {
	sentAt := time.Now()
	for _, email := range config.AdminEmails() {
		state.EmailsSentAt[email] = sentAt
	}
}

const (
	header = `
👋 there are aws-usage-alert notifications:

Group         Usage
-----------------------------
`
)
