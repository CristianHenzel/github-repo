package cmd

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	github "github.com/google/go-github/github"
	cobra "github.com/spf13/cobra"
	oauth2 "golang.org/x/oauth2"
)

type initFlags struct {
	User  string
	Token string
	Url   string
	Dir   string
}

func init() {
	var f initFlags

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize repository mirror",
		Run: func(cmd *cobra.Command, args []string) {
			runInit(f, false)
		},
	}

	initCmd.Flags().StringVarP(&f.User, "user", "u", "", "GitHub username")
	err := initCmd.MarkFlagRequired("user")
	fatalIfError(err)
	initCmd.Flags().StringVarP(&f.Token, "token", "t", "", "GitHub token")
	initCmd.Flags().StringVarP(&f.Url, "url", "r", "", "GitHub Enterprise URL")
	initCmd.Flags().StringVarP(&f.Dir, "dir", "d", "./", "Directory in which repositories will be stored")

	rootCmd.AddCommand(initCmd)
}

func newGithubClient(conf Configuration) *github.Client {
	var ctx = context.Background()
	var httpClient *http.Client

	// Create github client
	if conf.Token == "" {
		httpClient = http.DefaultClient
	} else {
		var tokenSource = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Token})
		httpClient = oauth2.NewClient(ctx, tokenSource)
	}
	var client = github.NewClient(httpClient)

	// Set base URL
	if conf.BaseUrl != "" {
		var endpoint, err = url.Parse(conf.BaseUrl)
		fatalIfError(err)
		if !strings.HasSuffix(endpoint.Path, "/") {
			endpoint.Path += "/"
		}
		client.BaseURL = endpoint
		client.UploadURL = endpoint
	}

	return client
}

func runInit(f initFlags, update bool) {
	var ctx = context.Background()
	var repos []*github.Repository

	if pathExists(configFile) && !update {
		fatalError(errConfExists)
		return
	}

	conf := Configuration{
		Username: f.User,
		Token:    f.Token,
		BaseDir:  f.Dir,
		BaseUrl:  f.Url,
	}

	// GetUint returns 0 if the flag was not set or if there is any error
	var con, _ = rootCmd.PersistentFlags().GetUint("concurrency")
	conf.Concurrency = con

	var client = newGithubClient(conf)

	user, _, err := client.Users.Get(ctx, conf.Username)
	if err != nil {
		fatalIfError(err)
		return
	}

	conf.Username = user.GetLogin()

	if user.GetName() != "" {
		conf.Fullname = user.GetName()
	} else {
		conf.Fullname = user.GetLogin()
	}

	if user.GetEmail() != "" {
		conf.Email = user.GetEmail()
	} else {
		conf.Email = conf.Username + "@users.noreply.github.com"
	}

	if conf.Token == "" {
		// Get public repositories for specified username
		repos, _, err = client.Repositories.List(ctx, conf.Username, nil)
	} else {
		// Get all repositories for authenticated user
		repos, _, err = client.Repositories.List(ctx, "", nil)
	}
	fatalIfError(err)
	for _, repo := range repos {
		url := *repo.CloneURL
		if conf.Token != "" {
			urlPrefix := conf.Username + ":" + conf.Token + "@"
			url = strings.Replace(url, "https://", "https://"+urlPrefix, -1)
			url = strings.Replace(url, "http://", "http://"+urlPrefix, -1)
		}
		dir := strings.Replace(*repo.FullName, "/", "_", -1)
		dir = strings.Replace(dir, conf.Username+"_", "", -1)
		dir = conf.BaseDir + "/" + dir
		branch := *repo.DefaultBranch

		conf.Repos = append(conf.Repos, Repo{
			Url:    url,
			Dir:    dir,
			Branch: branch,
		})
	}

	// Write config
	conf.save()
}
