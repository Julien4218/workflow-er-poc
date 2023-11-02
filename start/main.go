package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

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
	name := "World"
	channel := os.Getenv("SLACK_CHANNEL")
	fmt.Printf("Channel id is: %s", channel)
	we, err := c.ExecuteWorkflow(context.Background(), options, workflows.SlackWorkflow, name, channel)
	if err != nil {
		log.Fatalln("unable to complete SlackWorkflow", err)
	}

	// Get the results
	var greeting string
	err = we.Get(context.Background(), &greeting)
	if err != nil {
		log.Fatalln("unable to get SlackWorkflow result", err)
	}

	printResults(greeting, we.GetID(), we.GetRunID())
}

func printResults(greeting string, workflowID, runID string) {
	fmt.Printf("\nWorkflowID: %s RunID: %s\n", workflowID, runID)
	fmt.Printf("\n%s\n\n", greeting)
}
