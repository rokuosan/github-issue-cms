# GitHub Issue-based headless CMS

A headless CMS using GitHub Issues.

Issues are treated as articles.

## Prerequisites

- Go 1.25.0 or higher
- GitHub Token

## Installation

### 1. Install this application

```bash
$ go install github.com/rokuosan/github-issue-cms@v1.0.0
```

GitHub Releases also publish prebuilt binaries for macOS, Linux, and Windows.

### 2. Create Config file

Create a YAML file named ``gic.config.yaml`` and write your credentials.

```yaml
github:
  username: '<YOUR_GITHUB_USERNAME>'
  repository: '<YOUR_GITHUB_REPOSITORY>'

output:
  articles:
    directory: 'content/posts'
    filename: '%Y-%m-%d_%H%M%S.md'
  images:
    directory: 'static/images/%Y-%m-%d_%H%M%S'
    filename: '[:id].png'
    url: '/images/%Y-%m-%d_%H%M%S'
```

If you already have a legacy `hugo:` config section, it is still readable in `v1.0.0`.
Run `github-issue-cms migrate` to rewrite it to the canonical `output:` schema.

### 3. Run

Run this application with your GitHub Access Token

```bash
$ github-issue-cms generate --token="YOUR_GITHUB_TOKEN"
```

If your repository has issues and attached images, they will be exported like this tree.

These output paths are configurable, so you can adapt them to your site or build pipeline.

```bash
$ tree --dirsfirst
.
├── content
│   └── posts
│       ├── 2026-04-30_120000.md
│       └── 2026-04-30_121500.md
├── static
│   └── images
│       ├── 2026-04-30_120000
│       │   └── 0.png
│       └── 2026-04-30_121500
│           ├── 0.png
│           ├── 1.png
│           └── 2.png
└── gic.config.yaml
```

### 4. (Optional) Auto commit with GitHub Actions

GitHub Actions provides a built-in `GITHUB_TOKEN`, so you do not need to create a separate repository secret for this workflow.

Next, write this workflow with the permissions required to read issues and commit generated files.

```yaml
name: Go

on:
  push:
    branches: [ "main" ]
  issues:
    types: [reopened, closed]
permissions:
  contents: write
  issues: read

jobs:

  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2

    - name: Set up Go
      uses: actions/setup-go@4a3601121dd01d1626a1e23e37211e3254c1c06c # v6.4.0
      with:
        go-version: '1.25.0'
        # If your repository manages the Go version in go.mod, you can use this instead.
        # go-version-file: go.mod

    - name: Install
      run: go install github.com/rokuosan/github-issue-cms@v1.0.0

    - name: Generate
      run: github-issue-cms generate --token=${{ secrets.GITHUB_TOKEN }}

    - name: Auto Commit
      uses: stefanzweifel/git-auto-commit-action@04702edda442b2e678b25b537cec683a1493fcb9 # v7.1.0
      with:
        commit_message: "ci(github-issue-cms): :memo: Update content from GitHub Issues"
```

Congratulations.

Your Hugo site content will be regenerated and committed automatically when you push to `main` or an issue is closed or reopened.
