package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/emerald"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

func UpdaterUpdate() {
	if os.Getenv("DOTBOT_NO_UPDATE") == "1" {
		return
	}

	latest, err := GetGithubVersion(store.RepoUrl)
	if err != nil {
		log.Fatalln("failed to get latest version of dotbot", err)
	}
	if store.Version == latest {
		return
	}

	// quick check assets have been published
	head, err := http.Head(store.RepoUrl + "/releases/download/" + latest + "/checksums.txt")
	if err != nil {
		log.Fatalln("failed checking if assets are available", err)
	}
	if head.StatusCode == 404 {
		// quietly ignore until assets have been published
		return
	}
	logInstall("dotbot", store.Version, latest)

	// grab current real path
	exe := utils.ExecutablePath()
	exeNew := filepath.Join(filepath.Dir(exe), "."+filepath.Base(exe)+".new")
	exeExtractNew := filepath.Join(filepath.Dir(exe), "#."+filepath.Base(exe)+".new")
	exeOld := filepath.Join(filepath.Dir(exe), "."+filepath.Base(exe)+".old")

	if !utils.IsWritable(exe) {
		log.Log(emerald.Yellow, "WARN", emerald.Yellow, "skipping update as user does not have sufficient permissions")
		return
	}

	err = os.Rename(exe, exeOld)
	if err != nil {
		log.Fatalln("failed to rename executable to '.dotbot.old'", err)
	}

	// construct asset name
	arch := runtime.GOARCH
	ext := ".tar.gz"
	archiveExt := ""
	if runtime.GOOS == "windows" {
		if arch == "amd64" {
			arch = "x64"
		} else if arch == "386" {
			arch = "x86"
		}
		ext = ".zip"
		archiveExt = ".exe"
	} else if arch == "386" {
		arch = "i386"
	}
	asset := "dotbot_" + latest + "_" + runtime.GOOS + "_" + arch + ext

	// download archive
	dl := DownloadConfig{
		Url:  store.RepoUrl + "/releases/download/" + latest + "/" + asset,
		Mode: 438,
		Extract: ExtractItems{
			{
				Source: "dotbot" + archiveExt,
				Path:   exeExtractNew,
			},
		},
	}
	err = dl.Run()
	if err != nil {
		log.Fatalln("failed to download or extract archive", err)
	}
	store.RemoveTempFiles()

	// rename new file
	err = os.Rename(exeNew, exe)
	if err != nil {
		log.Fatalln("failed to rename executable to '.dotbot.new'", err)
	}

	// delete old file
	_ = os.Remove(exeOld)

	// restart program
	if runtime.GOOS == "windows" {
		cmd := exec.Command(exe, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalln("failed to automatically restart dotbot", err)
		}
		os.Exit(0)
	} else {
		err := syscall.Exec(exe, os.Args, os.Environ())
		if err != nil {
			log.Fatalln("failed to automatically restart dotbot", err)
		}
	}
}

func UpdaterCleanup() {
	if runtime.GOOS != "windows" {
		return
	}
	exe := utils.ExecutablePath()
	exeOld := filepath.Join(filepath.Dir(exe), "."+filepath.Base(exe)+".old")
	err := os.Remove(exeOld)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalln("failed to remove '.dotbot.exe.old' file", err)
	} else if err == nil {
		fmt.Println("removed old file")
	}
}
