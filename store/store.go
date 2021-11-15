package store

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var store map[string]string
var location string

var BaseDirectory string
var HomeDirectory string
var DryRun bool = true

const storePerm os.FileMode = 0600

func Set(key, value string) {
	store[key] = value
}

func Get(key string) string {
	return store[key]
}

func Save() error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(location, data, storePerm)
}

func SetSave(key, value string) error {
	Set(key, value)
	return Save()
}

func Load() {
	_, err := os.Stat(location)
	if os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(location), 755)
		if err != nil {
			log.Fatalln("Failed to create folder for state file", err)
		}
		err = os.WriteFile(location, []byte("{}"), storePerm)
		if err != nil {
			log.Fatalln("Failed to write default state file", err)
		}
	} else if err != nil {
		log.Fatalln("Failed to stat state file", err)
	}
	data, err := os.ReadFile(location)
	if err != nil {
		log.Fatalln("Failed to read state file", err)
	}
	err = json.Unmarshal(data, &store)
	if err != nil {
		log.Fatalln("Failed to unmarshal state file", err)
	}
}

func getHome() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Failed to find users home directory", err)
	}
	return dir
}

func getLocation() string {
	if location = os.Getenv("XDG_STATE_HOME"); location != "" {
		return location
	}
	if runtime.GOOS == "windows" {
		appdata, present := os.LookupEnv("LOCALAPPDATA")
		if !present {
			log.Fatalln("Failed to find local appdata directory")
		}
		return filepath.Join(appdata, "dotbot/state.json")
	}
	return filepath.Join(HomeDirectory, ".local/state/dotbot/state.json")
}

func init() {
	HomeDirectory = getHome()
	location = getLocation()
	Load()
	BaseDirectory = Get("directory")
}
