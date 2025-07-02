package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tf-plan-summary/tf-plan-summary/pkg/core/summarize"
)

var (
	summarizeCmd = &cobra.Command{
		Use:   "summarize [plan_detail]",
		Short: "Generates summaries from one or more terragrunt plan",
		Args:  cobra.MatchAll(cobra.MaximumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			var planDetail string
			if len(args) > 0 {
				planDetail = args[0] // Get the first positional argument
			}
			return summarize.Summarize(plansDir, planDetail, envProjectRegex)
		},
	}
)

func init() {
	summarizeCmd.Flags().StringVar(&plansDir, "plans-dir", "plans", "Directories where the plans are stored.")
	summarizeCmd.Flags().StringVar(&envProjectRegex, "env-project-regex", "^/?([^/]+)/?(.*)", "Regex to parse the environment and project name from the project path")
}
