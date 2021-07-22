package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/user"
	"strings"

	gitconfig "github.com/go-git/go-git/v5/config"
	github "github.com/google/go-github/github"
	cobra "github.com/spf13/cobra"
	oauth2 "golang.org/x/oauth2"
)

var (
	gitAliasesRepo = "gr-git-aliases"
	gitAliasesFile = "aliases.json"
)

type gitAlias struct {
	Alias   string `json:"alias"`
	Command string `json:"command"`
}

func init() {
	if cFlags == nil {
		cFlags = &Configuration{}
	}

	initCmd := &cobra.Command{
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

func newGithubClient(conf *Configuration) *github.Client {
	var httpClient *http.Client

	ctx := context.Background()

	// Create github client
	if conf.Token == "" {
		httpClient = http.DefaultClient
	} else {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Token})
		httpClient = oauth2.NewClient(ctx, tokenSource)
	}

	client := github.NewClient(httpClient)

	// Set base URL
	if conf.BaseURL != "" {
		endpoint, err := url.Parse(conf.BaseURL)
		fatalIfError(err)

		if !strings.HasSuffix(endpoint.Path, "/") {
			endpoint.Path += "/"
		}

		client.BaseURL = endpoint
		client.UploadURL = endpoint
	}

	return client
}

func addGitAliases(ctx context.Context, conf *Configuration, client *github.Client) {
	var ga []gitAlias

	aliasesContent, _, _, err := client.Repositories.GetContents(ctx, conf.Username, gitAliasesRepo, gitAliasesFile, nil)
	fatalIfError(err)
	aliasesBytes, err := base64.StdEncoding.DecodeString(*aliasesContent.Content)
	fatalIfError(err)
	fatalIfError(json.Unmarshal(aliasesBytes, &ga))

	cfg := gitconfig.NewConfig()
	usr, err := user.Current()
	fatalIfError(err)

	gitconfigPath := usr.HomeDir + "/.gitconfig"
	b, err := ioutil.ReadFile(gitconfigPath)
	fatalIfError(err)

	fatalIfError(cfg.Unmarshal(b))

	section := cfg.Raw.Section("alias")
	for _, alias := range ga {
		section.SetOption(alias.Alias, alias.Command)
	}

	fatalIfError(cfg.Validate())
	bytes, err := cfg.Marshal()
	fatalIfError(err)

	fatalIfError(ioutil.WriteFile(gitconfigPath, bytes, 0o644))
}

func getRepos(ctx context.Context, conf *Configuration, client *github.Client) (repositories []Repo) {
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
		cloneURL := *repo.CloneURL
		dir := *repo.FullName
		parent := ""

		if *repo.Name == gitAliasesRepo {
			addGitAliases(ctx, conf, client)
		}

		if *repo.Fork {
			repo, _, err = client.Repositories.GetByID(ctx, *repo.ID)
			fatalIfError(err)

			parent = *repo.Parent.CloneURL
		}

		if conf.Token != "" {
			urlPrefix := conf.Username + ":" + conf.Token + "@"
			cloneURL = strings.ReplaceAll(cloneURL, "https://", "https://"+urlPrefix)
			cloneURL = strings.ReplaceAll(cloneURL, "http://", "http://"+urlPrefix)
		}

		if !conf.SubDirs {
			dir = strings.ReplaceAll(dir, "/", "_")
			dir = strings.ReplaceAll(dir, conf.Username+"_", "")
		}

		dir = conf.BaseDir + "/" + dir
		branch := *repo.DefaultBranch

		repositories = append(repositories, Repo{
			URL:    cloneURL,
			Dir:    dir,
			Branch: branch,
			Parent: parent,
		})
	}

	return repositories
}

func runInit(conf *Configuration, update bool) {
	ctx := context.Background()

	if pathExists(configFile) && !update {
		fatalError(errConfExists)

		return
	}

	// GetUint returns 0 if the flag was not set or if there is any error
	con, _ := rootCmd.PersistentFlags().GetUint("concurrency")
	conf.Concurrency = con

	client := newGithubClient(conf)

	usr, _, err := client.Users.Get(ctx, conf.Username)
	if err != nil {
		fatalIfError(err)

		return
	}

	conf.Username = usr.GetLogin()

	if usr.GetName() != "" {
		conf.Fullname = usr.GetName()
	} else {
		conf.Fullname = usr.GetLogin()
	}

	if usr.GetEmail() != "" {
		conf.Email = usr.GetEmail()
	} else {
		conf.Email = conf.Username + "@users.noreply.github.com"
	}

	conf.Repos = getRepos(ctx, conf, client)

	// Write config
	conf.save()
}
