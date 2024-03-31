---
weight: 1
title: 'インストール'
bookcase_cover_src: 'cover/catalogue.png'
bookcase_cover_src_dark: 'cover/catalogue_dark.png'
---

# GitHub Issue CMS の導入

このページではGitHub Issue CMSの導入について説明します。

## 前提条件

以下のものを利用します。事前に準備しておいてください。

- Go v1.22.1 (or higher)
- Hugo extended
- GitHub Access Token

## 手順

### 1. アプリケーションのインストール

以下のコマンドを実行します。

```shell
$ go install github.com/rokuosan/github-issue-cms@latest
```

### 2. コンフィグの作成

``gic.config.yaml``という名前でファイルを作成し、以下のような内容を記述します。

```yaml
github:
  username: '<YOUR_GITHUB_USERNAME>'
  repository: '<YOUR_GITHUB_REPOSITORY>'

hugo:
  directory:
    articles: 'content/posts'
  url:
    appendSlash: false
    images: '/images'
```

### 3. 実行

以下のコマンドを実行して、対象のリポジトリからすべてのIssueをMarkdownに変換します。

```shell
$ github-issue-cms generate --token="<YOUR_GITHUB_ACCESS_TOKEN>" -d
```

もし、Issueに添付画像がある場合は以下のように出力されます。

```shell
$ tree --dirsfirst
.
├── content
│   └── posts
│       ├── 2023-12-21_151921.md
│       └── 2023-12-22_063216.md
├── static
│   └── images
│       ├── 2023-12-21_151921
│       │   └── 0.png
│       └── 2023-12-22_063216
│           ├── 0.png
│           ├── 1.png
│           └── 2.png
└── gic.config.yaml
```

出力されるディレクトリはHugoで利用される一般的なディレクトリに一致します。
もし別のディレクトリに出力する際は、``gic.config.yaml``を編集してください。

### 4. GitHub Actions との連携(任意)

GitHub Access Tokenをリポジトリのシークレット変数に格納します。(ここでは、``GH_TOKEN``という名前で登録しています。)

次に、以下のようなワークフローを定義します。

```yaml
name: Go

on:
  push:
    branches: [ "main" ]
  issues:
    types: [reopened, closed]
  pull_request:
    branches: [ "main" ]
permissions: write-all

jobs:

  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.22.1'

    - name: Install
      run: go install github.com/rokuosan/github-issue-cms@latest

    - name: Generate
      run: github-issue-cms generate --token=${{ secrets.GH_TOKEN }}

    - name: Auto Commit
      uses: stefanzweifel/git-auto-commit-action@v4
      with:
        commit_message: "ci: :memo: Add new articles"
```

最後にIssueを書き、Closeすると自動で記事が生成されます。
