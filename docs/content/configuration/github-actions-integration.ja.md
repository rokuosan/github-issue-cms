---
title: 'GitHub Actions との連携'
date: 2021-12-25
weight: 2
---

このページでは GitHub Actions を使ってコンテンツを自動生成する方法を説明します。

## 概要

`github-issue-cms` は GitHub Actions の組み込み `GITHUB_TOKEN` で実行できます。

以下のワークフローでは、次の場合に Markdown を再生成します。

- `main` への push
- Issue の reopen
- Issue の close

## ワークフロー例

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

## 補足

- 生成ファイルをリポジトリへコミットするために `contents: write` が必要です。
- GitHub API で Issue を読むために `issues: read` が必要です。
- `gic.config.yaml` をリポジトリルートに置く構成であれば、追加設定は不要です。
- リポジトリが Go module で、ルートに `go.mod` がある場合は `go-version` の代わりに `go-version-file: go.mod` を利用できます。
- Actions のログを増やしたい場合は `github-issue-cms -v generate --token=...` を利用してください。
