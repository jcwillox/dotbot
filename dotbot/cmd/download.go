package cmd

import (
	"bufio"
	"bytes"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/plugins"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/emerald"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var dwFlags struct {
	Executable bool
	Force      bool
}

var downloadCmd = &cobra.Command{
	Use:   "download [<url>...]",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var mode utils.WeakFileMode = 0666
		if dwFlags.Executable {
			mode = 0777
		}
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
		if len(args) > 0 {
			downloadFile(args[0], cwd, mode)
		} else {
			reader := bufio.NewReader(os.Stdin)
			urls := make([]string, 0, 5)
			for {
				line, err := reader.ReadBytes('\n')
				if err == io.EOF {
					break
				} else if err != nil {
					log.Fatalln(err)
				}
				if !bytes.HasPrefix(line, []byte("http")) {
					if emerald.ColorEnabled {
						emerald.CursorUp(1)
					}
					break
				}
				urls = append(urls, string(bytes.TrimSpace(line)))
			}
			for _, url := range urls {
				downloadFile(url, cwd, mode)
			}
		}
	},
}

func downloadFile(url string, path string, mode utils.WeakFileMode) {
	dl := plugins.DownloadConfig{
		Url:    url,
		Path:   path,
		Mode:   mode,
		Force:  dwFlags.Force,
		Mkdirs: true,
	}
	err := dl.Run()
	if err != nil {
		log.Fatalln("Failed downloading file", err)
	}
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().BoolVarP(&dwFlags.Executable, "executable", "x", false, "make downloaded file executable")
	downloadCmd.Flags().BoolVarP(&dwFlags.Force, "force", "f", false, "overwrite destination file if it exists")
}
