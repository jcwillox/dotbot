package cmd

import (
	"github.com/jcwillox/dotbot/plugins"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/emerald"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	color  string
	dryRun bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "dotbot",
	Short:   "",
	Version: "0.0.1",
	Run: func(cmd *cobra.Command, args []string) {
		if store.BaseDirectory != "" {
			err := os.Chdir(store.BaseDirectory)
			if err != nil {
				log.Fatalln("Unable to access dotfiles directory", err)
			}
		}
		path := "dotbot.yaml"
		if v, present := os.LookupEnv("DOTBOT_CONFIG"); present {
			path = v
		}
		config, err := plugins.ReadConfig(path)
		if err != nil {
			log.Panicln("Failed reading config file", err)
		}
		config.RunAll()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&color, "color", "auto", "when to use colors (always, auto, never)")
	rootCmd.PersistentFlags().BoolP("help", "h", false, "help for dotbot")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "enable dry run mode")

	_ = rootCmd.RegisterFlagCompletionFunc("color", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "always", "never"}, cobra.ShellCompDirectiveNoFileComp
	})
}

func initConfig() {
	// handle global flags
	switch color {
	case "auto":
		emerald.AutoSetColorState()
	case "always":
		emerald.SetColorState(true)
	case "never":
		emerald.SetColorState(false)
	}

	if rootCmd.PersistentFlags().Changed("dry-run") {
		store.DryRun = dryRun
	}
}
