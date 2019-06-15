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

func init() {
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize repository mirror",
		Run: func(cmd *cobra.Command, args []string) {
			runInit(cFlags, false)
		},
	}

	initCmd.Flags().StringVarP(&cFlags.Username, "user", "u", "", "GitHub username")
	fatalIfError(initCmd.MarkFlagRequired("user"))
	initCmd.Flags().StringVarP(&cFlags.Token, "token", "t", "", "GitHub token")
	initCmd.Flags().StringVarP(&cFlags.BaseURL, "url", "r", "", "GitHub Enterprise URL")
	initCmd.Flags().StringVarP(&cFlags.BaseDir, "dir", "d", ".", "Directory in which repositories will be stored")
	initCmd.Flags().BoolVarP(&cFlags.SubDirs, "subdirs", "s", false, "Enable creation of separate subdirectories for each org/user")

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
	if conf.BaseURL != "" {
		var endpoint, err = url.Parse(conf.BaseURL)
		fatalIfError(err)
		if !strings.HasSuffix(endpoint.Path, "/") {
			endpoint.Path += "/"
		}
		client.BaseURL = endpoint
		client.UploadURL = endpoint
	}

	return client
}

func getRepos(ctx context.Context, conf Configuration, client *github.Client) (repositories []Repo) {
	var repos []*github.Repository
	var err error

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
		dir := *repo.FullName

		if conf.Token != "" {
			urlPrefix := conf.Username + ":" + conf.Token + "@"
			url = strings.Replace(url, "https://", "https://"+urlPrefix, -1)
			url = strings.Replace(url, "http://", "http://"+urlPrefix, -1)
		}

		if !conf.SubDirs {
			dir = strings.Replace(dir, "/", "_", -1)
			dir = strings.Replace(dir, conf.Username+"_", "", -1)
		}

		dir = conf.BaseDir + "/" + dir
		branch := *repo.DefaultBranch

		repositories = append(repositories, Repo{
			URL:    url,
			Dir:    dir,
			Branch: branch,
		})
	}

	return repositories
}

func runInit(conf Configuration, update bool) {
	var ctx = context.Background()

	if pathExists(configFile) && !update {
		fatalError(errConfExists)
		return
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

	conf.Repos = getRepos(ctx, conf, client)

	// Write config
	conf.save()
}
