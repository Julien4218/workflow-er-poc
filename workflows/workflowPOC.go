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

// Workflow is a Hello World workflow definition.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("HelloWorld workflow completed.", "result", result)

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	//fmt.Printf("token is %s", os.Getenv("slacktoken"))

	api := slack.New(os.Getenv("SLACK_TOKEN"))
	//api := slack.New(os.Getenv("slacktoken"))

	attachment := slack.Attachment{
		Pretext: "pre-hello",
		Text:    "text-world",
	}
	channelID, timestamp, err := api.PostMessage("C063JK1RHN1", slack.MsgOptionText("hello world", false), slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Printf("%s\n", err)
		return "", nil
	}

	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
	return "Hello " + name + "!", nil
}
