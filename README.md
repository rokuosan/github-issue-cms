# GitHub Issue-based headless CMS for Hugo

A headless CMS for Hugo using GitHub Issues.

Issues are treated as articles.

## Prerequisites

- Go
- Hugo
- GitHub Token

## Installation

### 1. Install this application

```bash
$ go install github.com/rokuosan/github-issue-cms@v0.6.1
```

### 2. Create Config file

Create a YAML file named ``gic.config.yaml`` and write your credentials.

```yaml
github:
  username: '<YOUR_GITHUB_USERNAME>'
  repository: '<YOUR_GITHUB_REPOSITORY>'

hugo:
  filename:
    articles: '%Y-%m-%d_%H%M%S.md'
    images: '[:id].png'
  directory:
    articles: 'content/posts'
    images: 'static/images/%Y-%m-%d_%H%M%S'
  url:
    images: '/images/%Y-%m-%d_%H%M%S'
```

### 3. Run

Run this application with your GitHub Access Token

```bash
$ github-issue-cms generate --token="YOUR_GITHUB_TOKEN"
```

If your repository has issues and attached images, they will be exported like this tree.

These directories are compatible with Hugo directory structure, so you can quickly deploy this application to your Hugo site.

```bash
$ tree --dirsfirst
.
├── content
│   └── posts
│       ├── 2004501283.md
│       └── 2006779255.md
├── static
│   └── images
│       ├── 2004501283
│       │   └── 0.png
│       └── 2006779255
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
      run: go install github.com/rokuosan/github-issue-cms@v0.6.1

    - name: Generate
      run: github-issue-cms generate --token=${{ secrets.GITHUB_TOKEN }}

    - name: Auto Commit
      uses: stefanzweifel/git-auto-commit-action@04702edda442b2e678b25b537cec683a1493fcb9 # v7.1.0
      with:
        commit_message: "ci(github-issue-cms): :memo: Update content from GitHub Issues"
```

Congratulations.

Your Hugo site content will be regenerated and committed automatically when you push to `main` or an issue is closed or reopened.
