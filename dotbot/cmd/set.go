package cmd

import (
	"github.com/jcwillox/dotbot/store"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:       "set <property> <value>",
	Short:     "Modify properties in the per-user dotbot state file",
	ValidArgs: []string{"directory"},
	Args:      cobra.MinimumNArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		store.SetSave(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
