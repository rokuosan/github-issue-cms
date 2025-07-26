---
title: 'gic.config.yaml Configuration'
date: 2021-12-25
weight: 1
---

This page explains the configuration of `gic.config.yaml`.

## Overall Structure

`gic.config.yaml` has the following structure:

```yaml
github:
  username: 'rokuosan'
  repository: 'github-issue-cms'

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

## Configuration Items

### `github`

GitHub settings.

- `username`: GitHub username
- `repository`: Repository name to fetch issues from

### `hugo`

Hugo settings.

#### `filename`

- `articles`: Article filename
- `images`: Image filename

`[:id]` will be replaced with the image ID. The image ID is unique within each issue and assigned sequentially.

#### `directory`

- `articles`: Directory to save articles
- `images`: Directory to save images

#### `url`

- `images`: Image URL referenced from Markdown

## Placeholders

The following placeholders are available in `gic.config.yaml`:

- `%Y`: Year
- `%m`: Month
- `%d`: Day
- `%H`: Hour
- `%M`: Minute
- `%S`: Second

These placeholders can be used in the same format as `strftime`.

## Configuration Examples

### Using Hugo Page Bundles

#### `gic.config.yaml`
```yaml
hugo:
  filename:
    articles: 'index.md'
    images: '[:id].png'
  directory:
    articles: 'content/posts/%Y-%m-%d_%H%M%S'
    images: 'content/posts/%Y-%m-%d_%H%M%S'
  url:
    images: ''
```

#### Output Example

{{< filetree/container >}}
  {{< filetree/folder name="content" >}}
    {{< filetree/folder name="posts" >}}
      {{< filetree/folder name="2021-12-24_000000" >}}
        {{< filetree/file name="index.md" >}}
        {{< filetree/file name="0.png" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="2021-12-25_000000" >}}
        {{< filetree/file name="index.md" >}}
        {{< filetree/file name="0.png" >}}
      {{< /filetree/folder >}}
    {{< /filetree/folder >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}