package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	github "github.com/google/go-github/github"
	cobra "github.com/spf13/cobra"
	oauth2 "golang.org/x/oauth2"
)

type initFlags struct {
	User  string
	Token string
	Url   string
}

func init() {
	var f initFlags

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize repository mirror",
		Run: func(cmd *cobra.Command, args []string) {
			if pathExists(configFile) {
				fmt.Println("ERROR: Configuration file already exists in current directory. "+
					"Please run 'update' if you want to update your settings. "+
					"Alternatively, remove", configFile, "if you want to initialize "+
					"the repository again.")
				os.Exit(255)
			}

			runInit(f.User, f.Token, f.Url)
		},
	}

	initCmd.Flags().StringVarP(&f.User, "user", "u", "", "GitHub username")
	initCmd.MarkFlagRequired("user")
	initCmd.Flags().StringVarP(&f.Token, "token", "t", "", "GitHub token")
	initCmd.MarkFlagRequired("token")
	initCmd.Flags().StringVarP(&f.Url, "url", "r", "", "GitHub Enterprise URL")

	rootCmd.AddCommand(initCmd)
}

func runInit(username, token, baseurl string) {
	ctx := context.Background()

	conf := Configuration{
		Username: username,
		Token:    token,
		BaseUrl:  baseurl,
	}

	// Validate data
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: conf.Token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Set base URL
	if baseurl != "" {
		endpoint, err := url.Parse(baseurl)
		fatalIfError(err)
		if !strings.HasSuffix(endpoint.Path, "/") {
			endpoint.Path += "/"
		}
		conf.BaseUrl = baseurl
		client.BaseURL = endpoint
		client.UploadURL = endpoint
	}

	user, response, err := client.Users.Get(ctx, conf.Username)
	if err != nil {
		if response.StatusCode == 401 {
			fmt.Println("ERROR: Invalid token.")
			os.Exit(255)
		} else if response.StatusCode == 404 {
			fmt.Println("ERROR: Invalid user.")
			os.Exit(255)
		} else {
			fmt.Println(err)
			os.Exit(255)
		}
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

	repos, _, err := client.Repositories.List(ctx, "", nil)
	fatalIfError(err)
	for _, repo := range repos {
		urlPrefix := conf.Username + ":" + conf.Token + "@"
		url := strings.ReplaceAll(*repo.CloneURL, "https://", "https://"+urlPrefix)
		url = strings.ReplaceAll(url, "http://", "http://"+urlPrefix)
		dir := strings.ReplaceAll(*repo.FullName, "/", "_")
		dir = strings.ReplaceAll(dir, conf.Username+"_", "")
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
