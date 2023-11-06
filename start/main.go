package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/Julien4218/temporal-slack-activity/models"

	workflowPOC "github.com/Julien4218/workflow-poc"
	"github.com/Julien4218/workflow-poc/workflows"

	"go.temporal.io/sdk/client"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Create the client object just once per process
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	options := client.StartWorkflowOptions{
		ID:        "poc_task_queue",
		TaskQueue: workflowPOC.PocTaskQueue,
	}

	// Start the SlackWorkflow
	stackTrace := "Traceback (most recent call last):\n  File \"tb.py\", line 15, in <module>\n    a()\n  File \"tb.py\", line 3, in a\n    j = b(i)\n  File \"tb.py\", line 9, in b\n    c()\n  File \"tb.py\", line 13, in c\n    error()\nNameError: name 'error' is not defined\n"
	firstResponseWarning := "It looks like there might be an error."
	attachment := models.MessageAttachment{
		Pretext: "Here's the stack trace.",
		Text:    stackTrace,
	}
	channel := os.Getenv("SLACK_CHANNEL")
	fmt.Printf("Channel id is: %s", channel)
	workflowExecution, err := c.ExecuteWorkflow(context.Background(), options, workflows.SlackWorkflow, firstResponseWarning, channel, attachment)
	if err != nil {
		log.Fatalln("unable to complete SlackWorkflow", err)
	}

	// Get the results
	var greeting string
	err = workflowExecution.Get(context.Background(), &greeting)
	if err != nil {
		log.Fatalln("unable to get SlackWorkflow result", err)
	}

	printResults(greeting, workflowExecution.GetID(), workflowExecution.GetRunID())
}

func printResults(greeting string, workflowID, runID string) {
	fmt.Printf("\nWorkflowID: %s RunID: %s\n", workflowID, runID)
	fmt.Printf("\n%s\n\n", greeting)
}
