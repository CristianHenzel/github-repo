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

// Repo holds a repository URL and its local directory equivalent.
type Repo struct {
	URL    string `json:"url"`
	Dir    string `json:"dir"`
	Branch string `json:"branch"`
	Parent string `json:"parent"`
}

// Configuration holds git configuration data.
type Configuration struct {
	Fullname    string `json:"fullName"`
	Username    string `json:"username"`
	BaseDir     string `json:"baseDir"`
	BaseURL     string `json:"baseURL"`
	Token       string `json:"token"`
	Email       string `json:"email"`
	Concurrency uint   `json:"concurrency"`
	SubDirs     bool   `json:"subDirs"`
	Repos       []Repo `json:"repos"`
}

func loadConfig() *Configuration {
	conf := &Configuration{}

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
	err = ioutil.WriteFile(configFile, bytes, 0o644)
	fatalIfError(err)

	fmt.Println("Configuration saved. You can now run pull to download/update your repositories.")
}
