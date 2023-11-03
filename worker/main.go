package main

import (
	"log"

	"github.com/Julien4218/temporal-slack-activity/activities"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	workflowPOC "github.com/Julien4218/workflow-poc"
	"github.com/Julien4218/workflow-poc/workflows"
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

	// This worker hosts both SlackWorkflow and SlackMessageActivity functions
	w := worker.New(c, workflowPOC.PocTaskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.SlackWorkflow)
	w.RegisterActivity(activities.PostMessageActivity)

	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
