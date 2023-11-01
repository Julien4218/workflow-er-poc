package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/Julien4218/workflow-poc/workflows"

	slackActivities "github.com/Julien4218/temporal-slack-activity/activities"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run worker",
	Run: func(cmd *cobra.Command, args []string) {

		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		c, err := client.Dial(client.Options{
			HostPort:  os.Getenv("TEMPORAL_HOSTPORT"),
			Namespace: "default",
		})
		if err != nil {
			log.Fatalf("client error: %v", err)
		}
		defer c.Close()

		w := worker.New(c, workflows.QueueName, worker.Options{})

		w.RegisterWorkflow(workflows.ErWorkflow)
		w.RegisterActivity(slackActivities.PostMessageActivity)

		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalf("worker exited: %v", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() error {
	// workerCmd.Use = appName

	// Silence Cobra's internal handling of command usage help text.
	// Note, the help text is still displayed if any command arg or
	// flag validation fails.
	workerCmd.SilenceUsage = true

	// Silence Cobra's internal handling of error messaging
	// since we have a custom error handler in main.go
	workerCmd.SilenceErrors = true

	err := workerCmd.Execute()
	return err
}
