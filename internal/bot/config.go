package bot

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

type Configuration struct {
	BotToken string `json:"bot_token"`
func (cfg *Configuration) DumpConfig() error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal %+v failed: %v", cfg, err)
	}
	path := WhereCFG("") + "/config.json"
	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write %s failed:%v", path, err)
	}
	return nil
}

func newConfigFromGivenPath(path string) *Configuration {
	cfgPath := WhereCFG(path) + "/config.json"
	if cfgPath == "" {
		log.Fatal("get config path failed")
	}
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		log.Fatal("read config failed:" + err.Error())
	}
	var cfg *Configuration
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal("parsed config failed:" + err.Error())
	}
	return cfg
}

func NewConfig() *Configuration {
	return newConfigFromGivenPath("")
}

// WhereCFG give the config loader specific config path.
// If p is given, it will return given path. Else this function will
// find config from environment variable "BOTCFGPATH" or user's home directory.
// If can't found config from this place, return a null string value.
func WhereCFG(p string) (path string) {
	// if p had given, use p as path
	if p != "" {
		return p
	}

	// if path is specific in environment variable, use env as path
	if path = os.Getenv("BOTCFGPATH"); path != "" {
		return path
	}

	// if config path exist in user's home directory, use it as path
	u, err := user.Current()
	if err != nil {
		log.Fatalf("read user error: %v", err)
	}
	files, err := ioutil.ReadDir(u.HomeDir + "/.config")
	if err != nil {
		log.Fatalf("read directory %s error: %v", u.HomeDir+"/.config", err)
	}
	for _, file := range files {
		if path = file.Name(); path == "go-bot" {
			if file.IsDir() {
				return u.HomeDir + "/.config/" + path
			} else {
				log.Fatal("~/.config/go-bot is a directory")
			}
		}
	}
	return ""
}
