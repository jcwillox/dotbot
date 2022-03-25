package store

import (
	"encoding/json"
	"github.com/jcwillox/dotbot/log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	DryRun           = false
	Groups           []string
	RegisteredGroups []string
	HomeDirectory    string
	Version          = "devel"
	RepoUrl          = "https://github.com/jcwillox/dotbot"
)

var (
	store    map[string]string
	location string
)

const storePerm os.FileMode = 0600

func Set(key, value string) {
	store[key] = value
}

func Get(key string) string {
	return store[key]
}

func HasGet(key string) (string, bool) {
	val, present := store[key]
	return val, present
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
		err := os.MkdirAll(filepath.Dir(location), os.ModePerm)
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

func BaseDir() string {
	return Get("directory")
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
	if dryRun := os.Getenv("DRY_RUN"); dryRun == "true" {
		DryRun = true
	}
}

var tempFiles = make([]string, 0, 5)

func TrackTempFile(path string) {
	tempFiles = append(tempFiles, path)
}

func RemoveTempFiles() {
	for _, path := range tempFiles {
		err := os.Remove(path)
		if err != nil {
			log.Fatalln("Failed removing temporary file", err)
		}
	}
	tempFiles = tempFiles[:0]
}

var tmplVars = make(map[string]interface{})

func TmplVar(key string, val interface{}) {
	tmplVars[key] = val
}

func TmplVars(vars map[string]interface{}) {
	for key, newVal := range vars {
		tmplVars[key] = newVal
	}
}

func GetVar(key string) (value interface{}, present bool) {
	value, present = tmplVars[key]
	return value, present
}

func GetVars() map[string]interface{} {
	return tmplVars
}

func VarsClosure(vars map[string]interface{}) func() {
	prev := make(map[string]interface{})
	for key, newVal := range vars {
		if val, present := tmplVars[key]; present {
			prev[key] = val
		}
		tmplVars[key] = newVal
	}
	return func() {
		// iterate over changed keys and restore old value
		for key := range vars {
			if val, present := prev[key]; present {
				tmplVars[key] = val
			} else {
				// remove key if no old value
				delete(tmplVars, key)
			}
		}
	}
}
