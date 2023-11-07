package workflows

import (
	"fmt"
	"os"
	"time"

	"go.temporal.io/sdk/workflow"

	slackActivities "github.com/Julien4218/temporal-slack-activity/activities"
	"github.com/Julien4218/temporal-slack-activity/models"
	"github.com/sirupsen/logrus"

	"github.com/Julien4218/workflow-poc/instrumentation"
)

const WorkflowName = "ErWorkflow"
const QueueName = WorkflowName + "-Queue"

type ErWorkflowInput struct {
	Email string
	Tier  string
}

func ErWorkflow(ctx workflow.Context, input *ErWorkflowInput) (string, error) {
	logrus.Infof("%s-SlackWorkflow started:ErWorklow", instrumentation.Hostname)
	defer logrus.Infof("%s-SlackWorkflow completed:ErWorklow", instrumentation.Hostname)
	txn := instrumentation.NrApp.StartTransaction("ErWorkflow")
	defer txn.End()

	// Define the SlackMessageActivity Execution options
	// StartToCloseTimeout or ScheduleToCloseTimeout must be set
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	logrus.Infof("Got input:%s", input)
	// Execute the SlackMessageActivity synchronously (wait for the result before proceeding)
	var result string
	message := fmt.Sprintf("Hello %s, this is a tier %s", input.Email, input.Tier)

	// Start the SlackWorkflow
	stackTrace := "Traceback (most recent call last):\n  File \"tb.py\", line 15, in <module>\n    a()\n  File \"tb.py\", line 3, in a\n    j = b(i)\n  File \"tb.py\", line 9, in b\n    c()\n  File \"tb.py\", line 13, in c\n    error()\nNameError: name 'error' is not defined\n"
	firstResponseWarning := message + "It looks like there might be an error."
	attachment := models.MessageAttachment{
		Pretext: "Here's the stack trace.",
		Text:    stackTrace,
	}
	channel := os.Getenv("SLACK_CHANNEL")
	fmt.Printf("Channel id is: %s", channel)

	data := models.SlackActivityData{
		ChannelId:            channel,
		FirstResponseWarning: firstResponseWarning,
		Attachment:           attachment,
	}

	err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, data).Get(ctx, &result)

	if err != nil {
		return "", err
	}
	// Make the results of the SlackWorkflow available
	return result, nil
}
