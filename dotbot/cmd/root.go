package cmd

import (
	"github.com/jcwillox/dotbot/plugins"
	"github.com/jcwillox/emerald"
	"github.com/spf13/cobra"
	"log"
)

var (
	color string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "dotbot",
	Short:   "",
	Version: "0.0.1",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := plugins.ReadConfig("samples/config.yaml")
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
	rootCmd.RegisterFlagCompletionFunc("color", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
}
