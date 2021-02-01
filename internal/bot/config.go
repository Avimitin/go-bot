package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"sync"
)

type Configuration struct {
	BotToken     string                `json:"bot_token"`
	CertedGroups []int64               `json:"certed_groups"`
	Users        map[int]string        `json:"users"`
	certedGroups map[int64]interface{} `json:"-"`
	mu           sync.Mutex            `json:"-"`
}

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

func (cfg *Configuration) isCerted(target int64) bool {
	_, ok := cfg.certedGroups[target]
	return ok
}

func (cfg *Configuration) userPermission(target int) string {
	return cfg.Users[target]
}

func (cfg *Configuration) isBanned(target int) bool {
	return cfg.userPermission(target) == "banned"
}

func (cfg *Configuration) save() {
	for k := range cfg.certedGroups {
		cfg.CertedGroups = append(cfg.CertedGroups, k)
	}
	err := cfg.DumpConfig()
	if err != nil {
		log.Printf("fail to dump config:%v", err)
	}
	log.Println("making tmp file")
	byt, err := json.Marshal(cfg)
	if err != nil {
		log.Printf("marshal %+v failed:%v", cfg, err)
		return
	}
	err = ioutil.WriteFile("/home/config.json.tmp", byt, os.ModePerm)
	if err != nil {
		log.Println("saving tmp file failed, program exit:", err)
	}
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
	if cfg.CertedGroups == nil {
		log.Fatal("got nil groups")
	}
	if cfg.Users == nil {
		log.Fatal("got nil user data")
	}
	for _, g := range cfg.CertedGroups {
		cfg.certedGroups[g] = struct{}{}
	}
	cfg.CertedGroups = cfg.CertedGroups[0:]
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
