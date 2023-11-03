package workflows

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/slack-go/slack"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func SlackWorkflow(ctx workflow.Context, name string, channel string) (string, error) {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger := workflow.GetLogger(ctx)
	logger.Info("SlackWorkflow workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, SlackMessageActivity, name, channel).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("SlackWorkflow workflow completed.", "result", result)

	return result, nil
}

func SlackMessageActivity(ctx context.Context, name string, channel string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SlackMessageActivity", "name", name)

	api := slack.New(os.Getenv("SLACK_TOKEN"))

	stackTrace := "Traceback (most recent call last):\n  File \"tb.py\", line 15, in <module>\n    a()\n  File \"tb.py\", line 3, in a\n    j = b(i)\n  File \"tb.py\", line 9, in b\n    c()\n  File \"tb.py\", line 13, in c\n    error()\nNameError: name 'error' is not defined\n"
	attachment := slack.Attachment{
		Pretext: "Does this look like an error?",
		Text:    stackTrace,
	}
	channelID, timestamp, err := api.PostMessage(channel, slack.MsgOptionText(name, false), slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Printf("%s\n", err)
		return "", nil
	}

	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
	return "", nil
}
