# go-release

Script that will create and upload assets for a Github release for a given tag.

# Installation
Download the [binaries](https://github.com/segmentio/go-release/releases) or `go get github.com/segmentio/go-release`

# Usage
```
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
  --token <token>  Github Token. Checks $GITHUB_TOKEN if not provided.
  --tag <tag>      Git tag. Uses latest published tag if not provided. 
  --name <name>    Release name. Uses tag if not provided.
  --body <body>    Release body. Empty by default.
  --draft          Identify the release as a draft.
  --prerelease     Identify the release as a prerelease.
```