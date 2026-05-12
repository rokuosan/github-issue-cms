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
  labels:
    - 'article'

output:
  articles:
    directory: 'content/posts'
    filename: '%Y-%m-%d_%H%M%S.md'
  images:
    directory: 'static/images/%Y-%m-%d_%H%M%S'
    filename: '[:id].png'
    url: '/images/%Y-%m-%d_%H%M%S'
```

## Configuration Items

### `github`

GitHub settings.

- `username`: GitHub username
- `repository`: Repository name to fetch issues from
- `labels`: Only fetch issues that have all specified labels

### `output`

Output settings.

#### `articles`

- `directory`: Directory to save articles
- `filename`: Article filename

#### `images`

- `directory`: Directory to save images
- `filename`: Image filename
- `url`: Image URL referenced from Markdown
- `targets`: URL prefixes to detect and replace in issue bodies

If `targets` is omitted, the built-in GitHub attachment URL rules are used.
If `targets: []` is specified, no image URLs are detected or replaced.
Wildcard host patterns such as `https://*.githubusercontent.com` are also supported.

`[:id]` will be replaced with the image ID. The image ID is unique within each issue and assigned sequentially.

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
output:
  articles:
    directory: 'content/posts/%Y-%m-%d_%H%M%S'
    filename: 'index.md'
  images:
    directory: 'content/posts/%Y-%m-%d_%H%M%S'
    filename: '[:id].png'
    url: ''
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
