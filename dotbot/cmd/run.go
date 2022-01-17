package cmd

import (
	"fmt"
	"github.com/jcwillox/dotbot/plugins"
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/emerald"
	"github.com/k0kubun/pp/v3"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

var (
	fromStdin bool
)

var runCmd = &cobra.Command{
	Use:       "run [<directive>] [<key=value...>]",
	Short:     "Execute individual dotbot configs or directives",
	ValidArgs: []string{"template"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			// special case to allow easily testing templates
			if args[0] == "template" {
				if len(args) < 2 {
					fmt.Println("No template provided!")
					os.Exit(1)
				}
				result, err := template.Parse(args[1]).Render()
				if !emerald.ColorEnabled {
					if err != nil {
						log.Fatalln(err)
					}
					fmt.Println(result)
				} else {
					if err != nil {
						pp.Println(err)
						os.Exit(1)
					}
					pp.Println(result)
				}
			}
		} else {
			_ = utils.ChBaseDir()
			if fromStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					log.Panicln("Failed reading from std-input", err)
				}
				config, err := plugins.FromBytes(data)
				if err != nil {
					log.Fatalln("Failed parsing config from std-input", err)
				}
				config.RunAll(true)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&fromStdin, "stdin", false, "read config from std-input")
}
