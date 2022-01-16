package cmd

import (
	"github.com/jcwillox/dotbot/plugins"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "updates dotbot and dotfiles repo if possible",
	Run: func(cmd *cobra.Command, args []string) {
		plugins.UpdaterUpdate()
		_, _ = plugins.UpdaterUpdateRepo()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
