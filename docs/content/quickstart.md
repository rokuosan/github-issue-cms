---
title: 'Quick Start'
date: 2021-12-25
weight: 1
---

This page explains how to get started with GitHub Issue CMS.

## Prerequisites

The following are required. Please prepare them in advance.

- Go v1.22.1 (or higher)
- Hugo extended
- GitHub Access Token

## Steps

{{% steps %}}

### 1. Install the application

Run the following command:

```shell
$ go install github.com/rokuosan/github-issue-cms@latest
```

### 2. Create configuration file

Create a file named ``gic.config.yaml`` and write the following content:

The repository specified here is the one from which issues will be fetched.
Please prepare a GitHub Access Token that has access permissions to the target repository.

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

### 3. Execute

Run the following command to convert all issues from the target repository to Markdown:

```shell
$ github-issue-cms generate --token="<YOUR_GITHUB_ACCESS_TOKEN>" -d
```

If issues have attached images, the output will be as follows:

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

The output files and directories can be changed in ``gic.config.yaml``.

For more information about ``gic.config.yaml`` settings, please refer to [gic.config.yaml Configuration](../configuration/parameters).

{{% /steps %}}
