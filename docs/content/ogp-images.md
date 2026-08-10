---
title: 'OGP Images'
weight: 2
---

`github-issue-cms` can render a 1200x630 JPEG Open Graph Protocol (OGP) image from article metadata.

## Generate images with articles

Add `--with-ogimage` when generating content.

```shell
$ github-issue-cms generate --token="YOUR_GITHUB_TOKEN" --with-ogimage
```

The command creates one OGP image for each generated article.

For a page bundle (`index.md`), the image is saved as `ogp.jpeg` in the bundle directory.

For a flat Markdown file, `post.md` becomes `post.ogp.jpeg` in the same directory.

## Generate one image

Use `ogimage` to render a local Markdown article.

```shell
$ github-issue-cms ogimage --file content/posts/2024-01-15_103000.md
```

This command writes `ogp.jpeg` to the image output directory configured in `gic.config.yaml`.

The renderer uses headless Chromium.

It downloads Chromium automatically when necessary.

Set `GIC_CHROMIUM_BIN` to use an existing Chromium-compatible browser binary instead.

## Customize the template

Pass an HTML template with `--template`.

```shell
$ github-issue-cms ogimage --file article.md --template ogp.html
```

Templates use Go's `html/template` syntax and receive these values:

- `.Title`
- `.Author`
- `.Date`
- `.Category`
- `.Tags`

For example:

```html
<h1>{{ .Title }}</h1>
{{ range .Tags }}<span>#{{ . }}</span>{{ end }}
```

## Preview a template in a browser

Start the local preview server while editing the template.

```shell
$ github-issue-cms ogimage preview ogp.html
```

Open `http://localhost:6140` in a browser.

The preview uses sample article data, displays the template in a fixed 1200x630 viewport, and reloads automatically after the template file changes.

Relative URLs in the template are served from the template directory, so local CSS, images, and fonts can be previewed without copying them elsewhere.

Use `--port` or `--host` when the default address is unavailable.

```shell
$ github-issue-cms ogimage preview ogp.html --port 8080
```
