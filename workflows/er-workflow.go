package workflows

import (
	"fmt"
	"github.com/slack-go/slack"
	"os"
	"time"

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
	var result slackModels.MessageDetails
	if err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, requiredSlackData).Get(ctx, &result); err != nil {
		logrus.Errorf("Activity failed. Error: %s", err)
		return "", err
	}

	var incidentIsPending = true
	for incidentIsPending {
		var reactions []slack.ItemReaction
		err := workflow.ExecuteActivity(ctx, slackActivities.GetMessageReactions, result.ChannelID, result.Timestamp).Get(ctx, &reactions)
		if err != nil {
			return "", err
		}
		logrus.Infof("logging reactions%t", reactions)

		if hasIsIncidentReactionOnMessage(reactions) {
			incidentIsPending = false
			logrus.Infof("is incident")
			requiredSlackData := slackModels.SlackActivityData{
				ChannelId: os.Getenv("SLACK_CHANNEL"),
				//todo change the datastructure of the SlackActivityData object
				FirstResponseWarning: "Thanks for confirming the incident. Let's get this party started! :tada:",
				Attachment: slackModels.MessageAttachment{
					Pretext: "",
					Text:    "",
				},
			}
			if err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, requiredSlackData).Get(ctx, &result); err != nil {
				return "", err
			}

		} else if hasNotIncidentReactionOnMessage(reactions) {
			incidentIsPending = false
			logrus.Infof("is not incident")
			requiredSlackData := slackModels.SlackActivityData{
				ChannelId: os.Getenv("SLACK_CHANNEL"),
				//todo change the datastructure of the SlackActivityData object
				FirstResponseWarning: "No errors in sight!",
				Attachment: slackModels.MessageAttachment{
					Pretext: "",
					Text:    "",
				},
			}
			if err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, requiredSlackData).Get(ctx, &result); err != nil {
				return "", err
			}
			return "", nil
		} else {
			err := workflow.Sleep(ctx, 1*time.Second)
			if err != nil {
				return "", err
			}

			//	todo we should maybe have a max lifetime for this workflow
		}
	}

	logrus.Infof("SlackWorkflow workflow completed. Result: %s", result)
	return "", nil
}

func hasNotIncidentReactionOnMessage(reactions []slack.ItemReaction) bool {
	if len(reactions) > 0 {
		fmt.Print(reactions[0].Name)
		if reactions[0].Name == "two" {
			return true
		}
	}
	return false
}

func hasIsIncidentReactionOnMessage(reactions []slack.ItemReaction) bool {
	if len(reactions) > 0 {
		fmt.Print(reactions[0].Name)
		if reactions[0].Name == "one" {
			return true
		}
	}
	return false
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
