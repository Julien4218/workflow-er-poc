package workflows

import (
	"time"

	"github.com/slack-go/slack"

	"github.com/Julien4218/temporal-slack-activity/activities"

	"go.temporal.io/sdk/workflow"
)

func SlackWorkflow(ctx workflow.Context, firstResponseWarning string, channel string, attachment slack.Attachment) (string, error) {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger := workflow.GetLogger(ctx)
	logger.Info("SlackWorkflow workflow started", "firstResponseWarning", firstResponseWarning)

	var result string
	err := workflow.ExecuteActivity(ctx, activities.PostMessageActivity, firstResponseWarning, channel, attachment).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("SlackWorkflow workflow completed.", "result", result)

	return result, nil
}
