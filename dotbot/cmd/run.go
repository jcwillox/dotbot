package cmd

import (
	"github.com/jcwillox/dotbot/plugins"
	"github.com/jcwillox/dotbot/store"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

var (
	fromStdin bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "execute individual dotbot configs/directives",
	Run: func(cmd *cobra.Command, args []string) {
		if store.BaseDirectory != "" {
			err := os.Chdir(store.BaseDirectory)
			if err != nil {
				log.Fatalln("Unable to access dotfiles directory", err)
			}
		}
		if fromStdin {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Panicln("Failed reading from std-input", err)
			}
			config, err := plugins.FromBytes(data)
			if err != nil {
				log.Fatalln("Failed parsing config from std-input", err)
			}
			config.RunAll()
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&fromStdin, "stdin", false, "read config from std-input")
}
