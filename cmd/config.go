package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"syscall"
)

const configFile = "gr.conf"

// Repo holds a repository URL and its local directory equivalent
type Repo struct {
	URL    string `json:"URL"`
	Dir    string `json:"Dir"`
	Branch string `json:"Branch"`
	Parent string `json:"Parent"`
}

// Configuration holds git configuration data
type Configuration struct {
	Fullname    string `json:"Fullname"`
	Username    string `json:"Username"`
	BaseDir     string `json:"BaseDir"`
	BaseURL     string `json:"BaseURL"`
	Token       string `json:"Token"`
	Email       string `json:"Email"`
	Concurrency uint   `json:"Concurrency"`
	SubDirs     bool   `json:"SubDirs"`
	Repos       []Repo `json:"Repos"`
}

func loadConfig() *Configuration {
	var conf *Configuration = &Configuration{}

	cwd, err := os.Getwd()
	fatalIfError(err)

	for {
		filePath := path.Join(cwd, configFile)

		bytes, err := ioutil.ReadFile(filePath)
		e, ok := err.(*os.PathError)
		if ok && e.Err == syscall.ENOENT {
			if cwd == "/" {
				fatalError(errConfNotExists)
				return conf
			}

			cwd = path.Dir(cwd)
			continue
		}
		fatalIfError(err)

		err = json.Unmarshal(bytes, conf)
		fatalIfError(err)

		err = os.Chdir(cwd)
		fatalIfError(err)

		return conf
	}
}

func (conf *Configuration) save() {
	bytes, err := json.MarshalIndent(conf, "", "\t")
	fatalIfError(err)
	err = ioutil.WriteFile(configFile, bytes, 0644)
	fatalIfError(err)

	fmt.Println("Configuration saved. You can now run pull to download/update your repositories.")
}
