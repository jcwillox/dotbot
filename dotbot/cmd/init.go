package cmd

import (
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/plugins"
	"github.com/jcwillox/dotbot/store"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var apply = false

var initCmd = &cobra.Command{
	Use:   "init <owner>[/<repo>]",
	Short: "Clone and setup a dotbot dotfiles repo",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repo := args[0]
		// assume repo is named dotfiles
		if !strings.Contains(repo, "/") {
			repo += "/dotfiles"
		}
		parts := strings.SplitN(repo, "/", 2)
		name := parts[1]

		err := plugins.GitConfig{
			Url:     "https://github.com/" + repo,
			Path:    name,
			Name:    repo,
			Method:  "clone_pull",
			Shallow: false,
		}.Run()
		if err != nil {
			log.Fatalln("failed to clone repo", err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			log.Panicln("failed to get current directory", err)
		}

		store.SetSave("directory", filepath.Join(cwd, name))

		if apply {
			_ = os.Setenv("DOTBOT_NO_UPDATE_REPO", "1")
			rootCmd.Run(rootCmd, nil)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&apply, "apply", false, "run dotbot immediately after cloning")
}
