package cmd

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/plugins"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/emerald"
	"github.com/spf13/cobra"
	"os"
)

var (
	color  string
	dryRun bool
	debug  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "dotbot",
	Short:   "A powerful bootstrapping utility for your dotfiles and system",
	Version: store.Version,
	Run: func(cmd *cobra.Command, args []string) {
		err := utils.ChBaseDir()
		if err != nil {
			log.Fatalln(err)
		}
		path := utils.GetConfigPath()
		if loadRunConfig(path) {
			fmt.Println("reloading configuration...")
			loadRunConfig(path)
		}
	},
}

func loadRunConfig(path string) bool {
	config, err := plugins.ReadConfig(path)
	if err != nil {
		log.Fatalln("config file not found:", err)
	}
	return config.RunAll()
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringSliceVarP(&store.Groups, "group", "g", nil, "run a specific group of directives")
	rootCmd.Flags().BoolP("version", "v", false, "version for dotbot")
	rootCmd.PersistentFlags().StringVar(&color, "color", "auto", "when to use colors (always, auto, never)")
	rootCmd.PersistentFlags().BoolP("help", "h", false, "help for dotbot")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "enable dry run mode")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debugging output")

	_ = rootCmd.RegisterFlagCompletionFunc("color", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "always", "never"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = rootCmd.RegisterFlagCompletionFunc("group", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if base, present := store.HasGet("directory"); present {
			err := os.Chdir(base)
			if err == nil {
				path := utils.GetConfigPath()
				_, _ = plugins.ReadConfig(path)
			}
		}
		return store.RegisteredGroups, cobra.ShellCompDirectiveNoFileComp
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
	if rootCmd.PersistentFlags().Changed("debug") {
		log.EnableDebug = debug
	}
}
