package cmd

import (
	"github.com/jcwillox/dotbot/store"
	"github.com/spf13/cobra"
	"golang.org/x/sys/execabs"
	"log"
	"os"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Run git diff in dotfiles directory",
	Run: func(_ *cobra.Command, args []string) {
		cmd := execabs.Command("git", "-C", store.BaseDir(), "diff")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			log.Fatalln("failed running git diff command:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
