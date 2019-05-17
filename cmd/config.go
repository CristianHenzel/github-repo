package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
)

const configFile = "gr.conf"

// Repo holds a repository URL and its local directory equivalent
type Repo struct {
	Url    string `json:"Url"`
	Dir    string `json:"Dir"`
	Branch string `json:"Branch"`
}

// Configuration holds git configuration data
type Configuration struct {
	Fullname string `json:"Fullname"`
	Username string `json:"Username"`
	BaseUrl  string `json:"BaseUrl"`
	Token    string `json:"Token"`
	Email    string `json:"Email"`
	Repos    []Repo `json:"Repos"`
}

func loadConfig() Configuration {
	var conf Configuration

	bytes, err := ioutil.ReadFile(configFile)
	e, ok := err.(*os.PathError)
	if ok && e.Err == syscall.ENOENT {
		fmt.Println("Couldn't find configuration file. Make sure that you are in the base " +
			"directory and that init has been run successfully.")
		os.Exit(255)
	}
	fatalIfError(err)

	err = json.Unmarshal(bytes, &conf)
	fatalIfError(err)

	return conf
}

func (conf *Configuration) save() {
	bytes, err := json.MarshalIndent(conf, "", "\t")
	fatalIfError(err)
	err = ioutil.WriteFile(configFile, bytes, 0644)
	fatalIfError(err)

	fmt.Println("Configuration saved. You can now run pull to download/update your repositories.")
}
