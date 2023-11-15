package workflows

import (
	"fmt"
	"os"
	"time"

	"github.com/slack-go/slack"

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
	var result slackModels.MessageDetails
	if err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, generateIncidentAlertMessage(generateInitialMessage(input))).Get(ctx, &result); err != nil {
		logrus.Errorf("Activity failed. Error: %s", err)
		return "", err
	}

	futureWait := "foo"
	err := workflow.ExecuteActivity(ctx, slackActivities.AddReaction, "one", result.ChannelID, result.Timestamp).Get(ctx, &futureWait)
	if err != nil {
		return "", err
	}
	workflow.ExecuteActivity(ctx, slackActivities.AddReaction, "two", result.ChannelID, result.Timestamp)

	var incidentIsPending = true
	for incidentIsPending {
		reactionKeysMap := make(map[string]bool)
		reactionCountsMap := make(map[string]int)

		s, err2 := getAndMapMessageReactions(ctx, result, reactionKeysMap, reactionCountsMap)
		if err2 != nil {
			return s, err2
		}

		if hasIsIncidentReactionOnMessage(reactionKeysMap, reactionCountsMap) {
			incidentIsPending = false
			s, err2 := doIncidentPath(ctx, result)
			if err2 != nil {
				return s, err2
			}
			continue
		} else if hasNotAnIncidentReactionOnMessage(reactionKeysMap, reactionCountsMap) {
			s, err2 := doNotIncidentPath(ctx, result)
			if err2 != nil {
				return s, err2
			}
			return "", nil
		} else {
			//	todo we should maybe have a max lifetime for this workflow
			s, err2 := waitForInput(ctx)
			if err2 != nil {
				return s, err2
			}
		}
	}

	logrus.Infof("SlackWorkflow workflow completed. Result: %s", result)
	return "", nil
}

func getAndMapMessageReactions(ctx workflow.Context, result slackModels.MessageDetails, reactionKeysMap map[string]bool, reactionCountsMap map[string]int) (string, error) {
	var reactions []slack.ItemReaction
	err := workflow.ExecuteActivity(ctx, slackActivities.GetMessageReactions, result.ChannelID, result.Timestamp).Get(ctx, &reactions)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(reactions); i++ {
		reactionKeysMap[reactions[i].Name] = true
		reactionCountsMap[reactions[i].Name] = reactions[i].Count
	}
	return "", nil
}

func waitForInput(ctx workflow.Context) (string, error) {
	err := workflow.Sleep(ctx, 1*time.Second)
	if err != nil {
		return "", err
	}
	return "", nil
}

func doNotIncidentPath(ctx workflow.Context, result slackModels.MessageDetails) (string, error) {
	if err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, generateNotAnIncidentMessage()).Get(ctx, &result); err != nil {
		return "", err
	}
	return "", nil
}

func doIncidentPath(ctx workflow.Context, result slackModels.MessageDetails) (string, error) {
	if err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, generateIncidentConfirmationMessage()).Get(ctx, &result); err != nil {
		return "", err
	}

	var childResponse string
	err := workflow.ExecuteChildWorkflow(ctx, IncidentWorkflow, IncidentWorkflowInput{}).Get(ctx, &childResponse)
	if err != nil {
		return "", err
	}
	logrus.Infof("child incident workflow completed")
	err = workflow.ExecuteChildWorkflow(ctx, RetrospectiveWorkflow, RetrospectiveWorkflowInput{}).Get(ctx, &childResponse)
	if err != nil {
		return "", err
	}

	if err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, generateIncidentResolvedMessage()).Get(ctx, &result); err != nil {
		return "", err
	}
	return "", nil
}

func generateIncidentConfirmationMessage() slackModels.SlackActivityData {
	logrus.Infof("Incident has been confirmed")
	requiredSlackData := slackModels.SlackActivityData{
		ChannelId:            os.Getenv("SLACK_CHANNEL"),
		FirstResponseWarning: "Thanks for confirming the incident. Let's get this party started! :tada:",
		Attachment: slackModels.MessageAttachment{
			Pretext: "",
			Text:    "",
		},
	}
	return requiredSlackData
}

func generateIncidentResolvedMessage() slackModels.SlackActivityData {
	logrus.Infof("child incident workflow completed")
	requiredSlackData := slackModels.SlackActivityData{
		ChannelId: os.Getenv("SLACK_CHANNEL"),
		//todo change the datastructure of the SlackActivityData object
		FirstResponseWarning: "Incident resolved! Thanks for all the hard work everyone :tada:",
		Attachment: slackModels.MessageAttachment{
			Pretext: "",
			Text:    "",
		},
	}
	return requiredSlackData
}

func generateNotAnIncidentMessage() slackModels.SlackActivityData {
	logrus.Infof("is not incident")
	requiredSlackData := slackModels.SlackActivityData{
		ChannelId:            os.Getenv("SLACK_CHANNEL"),
		FirstResponseWarning: "No errors in sight!",
		Attachment: slackModels.MessageAttachment{
			Pretext: "",
			Text:    "",
		},
	}
	return requiredSlackData
}

func generateInitialMessage(input *ErWorkflowInput) string {
	message := ""
	if input != nil {
		message = fmt.Sprintf("Hello %s, this is a tier %s", input.Email, input.Tier)
	}
	return message
}

func generateIncidentAlertMessage(message string) slackModels.SlackActivityData {
	slackActivityData := slackModels.SlackActivityData{
		ChannelId: os.Getenv("SLACK_CHANNEL"),
		FirstResponseWarning: message + "It looks like there might be an error. \n" +
			":one: To confirm the incident and start debugging \n" +
			":two: To dismiss",
		Attachment: slackModels.MessageAttachment{
			Pretext: "Here's the stack trace.",
			Text:    "Traceback (most recent call last):\n  File \"tb.py\", line 15, in <module>\n    a()\n  File \"tb.py\", line 3, in a\n    j = b(i)\n  File \"tb.py\", line 9, in b\n    c()\n  File \"tb.py\", line 13, in c\n    error()\nNameError: name 'error' is not defined\n",
		},
	}
	return slackActivityData
}

func hasNotAnIncidentReactionOnMessage(reactionKeysMap map[string]bool, reactionCountsMap map[string]int) bool {
	//todo these methods won't work for non english installs
	if reactionKeysMap["two"] && reactionCountsMap["two"] > 1 {
		return true
	}
	return false
}

func hasIsIncidentReactionOnMessage(reactionKeysMap map[string]bool, reactionCountsMap map[string]int) bool {
	if reactionKeysMap["one"] && reactionCountsMap["one"] > 1 {
		return true
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
