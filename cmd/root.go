package cmd

import (
	"os"

	"github.com/sirupsen/logrus"

	// "github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/tf-plan-summary/tf-plan-summary/pkg/core/log"
	"github.com/tf-plan-summary/tf-plan-summary/pkg/core/result"

	"github.com/spf13/cobra"
)

var (
	verbose         bool
	plansDir        string
	envProjectRegex string

	rootCmd = &cobra.Command{
		Use:   "tf-plan-summary",
		Short: "tf-plan-summary is a tool to generate summaries from a terragrunt run-* command",
		Long: `
tf-plan-summary is a tool to generate summaries from a terragrunt run-* command`,
	}
)

// Execute executes the root command.
func Execute() {
	logrus.SetFormatter(log.NewTextFormat())

	if err := rootCmd.Execute(); err != nil {
		logrus.Errorf("%s %s", result.FAILURE, err)
		os.Exit(1)
	}
}

func init() {

	logrus.SetOutput(os.Stdout)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "debug", "", false, "Debug Output")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
	}

	rootCmd.AddCommand(
		summarizeCmd,
		versionCmd,
		manCmd)
}
