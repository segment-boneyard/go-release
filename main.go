package main

import (
	"log"
	"os"

	"golang.org/x/oauth2"

	"github.com/blang/semver"
	"github.com/docopt/docopt.go"
	"github.com/google/go-github/github"
)

const (
	usage = `release.

Usage:
  release <org> <repo> [--assets=<assets>...]
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
	version = "0.1.0"
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	check(err)

	// Grab inputs.
	org := args["<org>"].(string)
	repo := args["<repo>"].(string)
	assets := args["--assets"].([]string)
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
	check(err)

	log.Println("Created release", *release.ID)

	for _, asset := range assets {
		file, err := os.Open(asset)
		check(err)

		log.Println("Uploading asset", asset)

		_, _, err = client.Repositories.UploadReleaseAsset(org, repo, *release.ID, &github.UploadOptions{asset}, file)
		check(err)
	}
}

func check(err error) {
	if err != nil {
		log.Fatalln("error:", err)
	}
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
	check(err)

	for {
		newTags, resp, err := client.Repositories.ListTags(org, name, opt)
		check(err)

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
