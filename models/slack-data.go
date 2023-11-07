package models

import "github.com/Julien4218/temporal-slack-activity/models"

type SlackData struct {
	ChannelId            string
	FirstResponseWarning string
	Attachment           models.MessageAttachment
}
