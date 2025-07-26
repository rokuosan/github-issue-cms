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

{{% steps %}}

### 1. アプリケーションのインストール

以下のコマンドを実行します。

```shell
$ go install github.com/rokuosan/github-issue-cms@latest
```

### 2. コンフィグの作成

``gic.config.yaml``という名前でファイルを作成し、以下のような内容を記述します。

ここで指定するリポジトリは、Issueを取得するリポジトリです。
GitHub Access Tokenは、対象のリポジトリに対するアクセス権を持つものを準備してください。

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

出力されるファイルやディレクトリは、``gic.config.yaml``で変更することができます。

``gic.config.yaml``の設定については、[gic.config.yaml の設定](../configuration/parameters)を参照してください。

{{% /steps %}}
