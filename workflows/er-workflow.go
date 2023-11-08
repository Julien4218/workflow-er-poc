package workflows

import (
	"fmt"
	"os"
	"time"

	"github.com/Julien4218/workflow-poc/models/signals"

	slackModels "github.com/Julien4218/temporal-slack-activity/models"
	"go.temporal.io/sdk/workflow"

	slackActivities "github.com/Julien4218/temporal-slack-activity/activities"
	"github.com/sirupsen/logrus"

	"github.com/Julien4218/workflow-poc/instrumentation"
)

const WorkflowName = "ErWorkflow"

type ErWorkflowInput struct {
	Email string
	Tier  string
}

func ErWorkflow(ctx workflow.Context, input *ErWorkflowInput) (string, error) {
	logrus.Infof("%s-SlackWorkflow started:ErWorklow", instrumentation.Hostname)
	defer logrus.Infof("%s-SlackWorkflow completed:ErWorklow", instrumentation.Hostname)
	txn := instrumentation.NrApp.StartTransaction("ErWorkflow")
	defer txn.End()

	ctx = updateWorkflowContextOptions(ctx)
	logrus.Infof("Got input:%s", input)
	// Execute the SlackMessageActivity synchronously (wait for the result before proceeding)
	message := fmt.Sprintf("Hello %s, this is a tier %s", input.Email, input.Tier)
	requiredSlackData := lookupSlackData(message)
	var result string
	if err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, requiredSlackData).Get(ctx, &result); err != nil {
		logrus.Errorf("Activity failed. Error: %s", err)
		return "", err
	}

	var slackIsIncidentSignal signals.SlackIsIncidentSignal
	signalChannel := workflow.GetSignalChannel(ctx, signals.SlackIsIncidentSignalName)
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChannel, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &slackIsIncidentSignal)
	})
	selector.Select(ctx)
	if slackIsIncidentSignal.IsIncident {
		logrus.Infof("is incident")
	} else {
		logrus.Infof("is not incident")
	}

	logrus.Infof("SlackWorkflow workflow completed. Result: %s", result)

	// Make the results of the SlackWorkflow available
	return result, nil
}

func updateWorkflowContextOptions(ctx workflow.Context) workflow.Context {
	// Define the SlackMessageActivity Execution options
	// StartToCloseTimeout or ScheduleToCloseTimeout must be set
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	return ctx
}

func lookupSlackData(message string) slackModels.SlackActivityData {
	//todo in the future do the actual calls to look this up
	slackActivityData := slackModels.SlackActivityData{
		ChannelId:            os.Getenv("SLACK_CHANNEL"),
		FirstResponseWarning: message + "It looks like there might be an error.",
		Attachment: slackModels.MessageAttachment{
			Pretext: "Here's the stack trace.",
			Text:    "Traceback (most recent call last):\n  File \"tb.py\", line 15, in <module>\n    a()\n  File \"tb.py\", line 3, in a\n    j = b(i)\n  File \"tb.py\", line 9, in b\n    c()\n  File \"tb.py\", line 13, in c\n    error()\nNameError: name 'error' is not defined\n",
		},
	}
	return slackActivityData
}
