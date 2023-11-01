package main

import (
	"github.com/Julien4218/workflow-poc"
	"github.com/Julien4218/workflow-poc/workflows"
	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
)

func main() {
	// Create the client object just once per process
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, workflowPOC.PocTaskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.Workflow)
	w.RegisterActivity(workflows.Activity)

	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
