---
title: 'GitHub Actions Integration'
date: 2021-12-25
weight: 2
---

This page explains how to regenerate content automatically with GitHub Actions.

## Overview

`github-issue-cms` can run in GitHub Actions with the built-in `GITHUB_TOKEN`.

The workflow below regenerates Markdown files when:

- you push to `main`
- an issue is reopened
- an issue is closed

## Example Workflow

```yaml
name: Generate content from GitHub Issues

on:
  push:
    branches: [ "main" ]
  issues:
    types: [reopened, closed]

permissions:
  contents: write
  issues: read

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2

      - name: Set up Go
        uses: actions/setup-go@4a3601121dd01d1626a1e23e37211e3254c1c06c # v6.4.0
        with:
          go-version: '1.25.0'

      - name: Install github-issue-cms
        run: go install github.com/rokuosan/github-issue-cms@v1.0.0

      - name: Generate content
        run: github-issue-cms generate --token=${{ secrets.GITHUB_TOKEN }}

      - name: Commit generated files
        uses: stefanzweifel/git-auto-commit-action@04702edda442b2e678b25b537cec683a1493fcb9 # v7.1.0
        with:
          commit_message: "ci(github-issue-cms): update content from GitHub Issues"
```

## Notes

- `contents: write` is required to commit generated files back to the repository.
- `issues: read` is required to read issues through the GitHub API.
- If you keep `gic.config.yaml` in the repository root, no extra setup is required.
- If your repository is a Go module and has a root `go.mod`, you can replace `go-version` with `go-version-file: go.mod`.
- Use `github-issue-cms -v generate --token=...` if you want more logs in the Actions output.
