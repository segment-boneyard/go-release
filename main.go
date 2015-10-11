package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/oauth2"

	"github.com/blang/semver"
	"github.com/docopt/docopt.go"
	"github.com/google/go-github/github"
)

const (
	usage = `release.

Usage:
  release <org> <repo> <assets>...
          [--tag <tag>]
          [--token <token>]
          [--name <name>]
          [--body <body>]
          [--draft]
          [--prerelease]
  release -h | --help
  release --version

Options:
  -h --help        Show this screen.
  --version        Show version.
  --token <token>  Github Token. Checks $GITHUB_TOKEN if not provided. [default: ]
  --tag <tag>      Git tag. Uses latest published tag if not provided. [default: ]
  --name <name>    Release name. Uses tag if not provided. [default: ]
  --body <body>    Release body. Empty by default. [default: ]
  --draft          Identify the release as a draft.
  --prerelease     Identify the release as a prerelease.
`
	version = "0.1.1"
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	// Grab inputs.
	org := args["<org>"].(string)
	repo := args["<repo>"].(string)
	assets := args["<assets>"].([]string)
	token := args["--token"].(string)
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			log.Fatal("github token is required")
		}
	}
	client := newGithub(token)
	tag := args["--tag"].(string)
	if tag == "" {
		tag = latestTag(client, org, repo)
	}
	name := args["--name"].(string)
	if name == "" {
		name = tag
	}
	body := args["--body"].(string)
	draft := args["--draft"].(bool)
	prerelease := args["--prerelease"].(bool)

	log.Printf("Creating release %q for %s/%s with tag %s.\n", name, org, repo, tag)

	// Create the release.
	release, _, err := client.Repositories.CreateRelease(org, repo, &github.RepositoryRelease{
		Name:       &name,
		Draft:      &draft,
		Prerelease: &prerelease,
		TagName:    &tag,
		Body:       &body,
	})
	if err != nil {
		panic(err)
	}

	log.Println("Created release", *release.ID)

	var wg sync.WaitGroup
	for _, asset := range assets {
		wg.Add(1)
		go func(asset string) {
			defer wg.Done()
			file, err := os.Open(asset)
			if err != nil {
				log.Println("[error] could not open", asset, err)
				return
			}

			log.Println("Uploading asset", asset)

			_, name := filepath.Split(asset)

			_, _, err = client.Repositories.UploadReleaseAsset(org, repo, *release.ID, &github.UploadOptions{name}, file)
			if err != nil {
				log.Println("[error] could not upload", asset, err)
				return
			}

			log.Println("Uploaded asset", asset)
		}(asset)
	}

	wg.Wait()
}

func newGithub(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return github.NewClient(tc)
}

func latestTag(client *github.Client, org, name string) string {
	opt := &github.ListOptions{PerPage: 100}

	latest, err := semver.Make("0.0.1")
	if err != nil {
		panic(err)
	}

	for {
		newTags, resp, err := client.Repositories.ListTags(org, name, opt)
		if err != nil {
			panic(err)
		}

		for _, tag := range newTags {
			version, err := semver.Make(*tag.Name)
			if err != nil {
				continue
			}

			if version.Compare(latest) > 0 {
				latest = version
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return latest.String()
}
