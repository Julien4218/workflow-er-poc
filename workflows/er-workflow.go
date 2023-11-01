package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	slackActivities "github.com/Julien4218/temporal-slack-activity/activities"
)

const WorkflowName = "ErWorkflow"
const QueueName = WorkflowName + "-Queue"

type ErWorkflowInput struct {
	Email string
	Tier  string
}

func ErWorkflow(ctx workflow.Context, input *ErWorkflowInput) (string, error) {
	// Define the Activity Execution options
	// StartToCloseTimeout or ScheduleToCloseTimeout must be set
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	// Execute the Activity synchronously (wait for the result before proceeding)
	var result string
	message := fmt.Sprintf("Hello %s, this is a tier %s", input.Email, input.Tier)
	err := workflow.ExecuteActivity(ctx, slackActivities.PostMessageActivity, message).Get(ctx, &result)
	if err != nil {
		return "", err
	}
	// Make the results of the Workflow available
	return result, nil
}
