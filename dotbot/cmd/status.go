package cmd

import (
	"github.com/jcwillox/dotbot/store"
	"github.com/spf13/cobra"
	"golang.org/x/sys/execabs"
	"log"
	"os"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Run git status in dotfiles directory",
	Run: func(_ *cobra.Command, args []string) {
		cmd := execabs.Command("git", "-C", store.BaseDir(), "status", "-s")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			log.Fatalln("failed running git status command:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
