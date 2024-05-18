---
title: 'クイックスタート'
date: 2021-12-25
weight: 1
---

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
