---
title: 'Quick Start'
date: 2021-12-25
weight: 1
---
A headless CMS for Hugo using GitHub Issues.

Issues are treated as articles.

## Prerequisites

- Go
- Hugo
- GitHub Token

## Installation

### 1. Install this application

```bash
$ go install github.com/rokuosan/github-issue-cms@latest
```

### 2. Create Config file

Create a YAML file named ``gic.config.yaml`` and write your credentials.

```yaml
github:
  username: 'your-name'
  repository: 'your-repository'

hugo:
  url:
    images: '/images'
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
