# GitHub Issue-based headless CMS

A headless CMS using GitHub Issues.

Issues are treated as articles.

## Prerequisites

- Go 1.25.0 or higher
- GitHub Token

## Installation

### 1. Install this application

```bash
$ go install github.com/rokuosan/github-issue-cms@v1.0.0
```

GitHub Releases also publish prebuilt binaries for macOS, Linux, and Windows.

### 2. Create Config file

Create a YAML file named ``gic.config.yaml`` and write your credentials.

```yaml
github:
  username: '<YOUR_GITHUB_USERNAME>'
  repository: '<YOUR_GITHUB_REPOSITORY>'

output:
  articles:
    directory: 'content/posts'
    filename: '%Y-%m-%d_%H%M%S.md'
  images:
    directory: 'static/images/%Y-%m-%d_%H%M%S'
    filename: '[:id].png'
    url: '/images/%Y-%m-%d_%H%M%S'
```

If you already have a legacy `hugo:` config section, it is still readable in `v1.0.0`.
Run `github-issue-cms migrate` to rewrite it to the canonical `output:` schema.

### 3. Run

Run this application with your GitHub Access Token

```bash
$ github-issue-cms generate --token="YOUR_GITHUB_TOKEN"
```

> [!NOTE]
> If your issues have images attached via drag-and-drop (`https://github.com/user-attachments/assets/...`) in a **private** repository, use a **classic** Personal Access Token. Fine-grained PATs and GitHub App installation tokens (including the Actions-provided `GITHUB_TOKEN`) are not accepted by GitHub's attachment download endpoint and will cause image downloads to fail with a 404.

### Preview an OGP template

Use the live preview server while editing a custom OGP template:

```bash
$ github-issue-cms ogimage preview ogp.html
```

Open `http://localhost:6140`. The template is rendered with sample article
data in a fixed 1200x630 OGP viewport and reloads automatically when the file
changes. Relative assets are served from the template's directory.

If your repository has issues and attached images, they will be exported like this tree.

These output paths are configurable, so you can adapt them to your site or build pipeline.

```bash
$ tree --dirsfirst
.
├── content
│   └── posts
│       ├── 2026-04-30_120000.md
│       └── 2026-04-30_121500.md
├── static
│   └── images
│       ├── 2026-04-30_120000
│       │   └── 0.png
│       └── 2026-04-30_121500
│           ├── 0.png
│           ├── 1.png
│           └── 2.png
└── gic.config.yaml
```

### Serve a JSON content API

Use `serve` to build a content snapshot from GitHub Issues at startup and serve
it over HTTP. API requests use the snapshot in memory, without contacting GitHub.
Restart the process to refresh articles and attachments.

```bash
export GITHUB_TOKEN='YOUR_GITHUB_TOKEN'
github-issue-cms serve --addr :8080 --page-size 20
```

Both `serve` and `export` read `gic.config.yaml` and respect `github.labels`.
They accept `--token` (`-t`), falling back to `GITHUB_TOKEN`. The default listen
address is `localhost:8080`; use `:8080` to listen on all interfaces for deployment.
SIGINT and SIGTERM shut down the server gracefully.

```bash
curl 'http://localhost:8080/data/articles?page=2&per_page=10&tag=Go'
curl 'http://localhost:8080/data/articles?q=Cloudflare&category=Engineering'
curl 'http://localhost:8080/data/articles?tag=Go&tag=Web&sort=title&order=asc'
curl 'http://localhost:8080/data/article/123'
```

| Parameter | Behavior |
| --- | --- |
| `page` | Positive page number, defaults to 1 |
| `per_page` | Articles per page, 1–100; defaults to `--page-size` (20) |
| `q` | Case-insensitive substring match in the title or Markdown body |
| `tag` | Exact tag match; repeated tags must all match |
| `category` | Exact category match |
| `sort` | `date` (default) or `title` |
| `order` | `desc` (default) or `asc` |

Filters combine with AND and run before sorting and pagination. Equal sort keys
are ordered by issue number in the requested direction. Invalid parameters return
HTTP 400; unknown articles return 404. A dynamic page beyond the result set returns
an empty `items` array with the actual totals.

```json
{
  "items": [
    {
      "id": 123,
      "author": "author",
      "title": "An article",
      "date": "2026-11-16T00:00:00Z",
      "category": "Engineering",
      "tags": ["Go"],
      "draft": false,
      "summary": "The beginning of the Markdown body…",
      "url": "/data/article/123.json"
    }
  ],
  "pagination": { "page": 1, "perPage": 20, "totalItems": 1, "totalPages": 1 }
}
```

Article IDs are repository-local issue numbers. Details contain `content` (the
Markdown body) and `metadata` (custom frontmatter), alongside the article fields.
Custom frontmatter must be JSON-serializable; unsupported nested YAML mappings
fail snapshot creation with the article ID. Recognized frontmatter fields override
issue metadata before filtering. Open issues
are drafts unless overridden with `draft: false`; drafts are excluded unless
`--include-drafts` is passed. Labels configured in `github.labels` select issues
and are removed from article tags, including tags supplied through frontmatter.

Attachments matching `output.images.targets` (or the default GitHub attachment
targets) are downloaded once, and their body URLs are rewritten to
`/images/{issue-number}/{image-id}.{extension}`. Other URLs remain external.
The new commands use these fixed image paths independently of the Markdown
generator's output path templates. Failed attachment downloads fail startup or
export so an incomplete snapshot is not published. Private repository attachment
downloads have the token requirements described above.

The API supports GET and HEAD and sends `Cache-Control: no-cache`. For a browser
frontend hosted on another origin, use `--cors-origin https://www.example.com`
or `--cors-origin '*'` for public cross-origin access. The default adds no CORS
headers. Credentialed CORS requests are not supported. In production, terminate
HTTPS at the hosting platform or proxy.

### Export static JSON content

Use `export` for hosting without a running content server:

```bash
github-issue-cms export --output dist --page-size 20
```

```text
dist/
├── .gic-content-export
├── _headers
├── data/
│   ├── metadata.json
│   ├── articles/page/1.json
│   ├── articles/page/2.json
│   ├── article/123.json
│   └── search/index.json
└── images/123/0.png
```

`metadata.json` includes `schemaVersion`, `generatedAt`, `totalArticles`,
`pageSize`, `totalPages`, the page URLs, an article path template, and the search
index URL and searchable fields. Page files have the same response format as
`/data/articles`. Each article detail includes its complete Markdown body.
The server also exposes these static JSON paths, including `.json` article URLs.

The static search index contains each article's ID, title, tags, summary, and
detail URL. Summaries are the first 200 Unicode characters of the trimmed Markdown
body, followed by an ellipsis when truncated. A browser client can fetch the index,
filter these fields, paginate the matches, and fetch details as needed. Search UI
and client logic are supplied by the consuming application. The index does not
include full article bodies; full-body search is available through `serve`'s `q`.
Query parameters do not filter static files. Fixed page URLs beyond `totalPages`
do not exist; an empty snapshot has zero pages and an empty search index.

The output directory is managed as a complete deployment: it must be absent,
empty, or a previous export containing `.gic-content-export`. A successful export
replaces the entire directory, removing old articles, pages, and images. Keep
handwritten files outside it. Files are built in a staging directory before
replacing the previous export, so a generation failure preserves existing output.

Deploy the complete directory to a static host, such as Cloudflare Workers Static
Assets, at the domain root. The generated `_headers` file sets JSON content types
and revalidation headers on hosts that support this convention; configure these
headers separately on other hosts. Regenerate and redeploy when issues change.
Do not configure an SPA fallback for `/data/*`: missing API resources should
return 404. The export does not generate a Worker search API or deploy resources.

### 4. (Optional) Auto commit with GitHub Actions

GitHub Actions provides a built-in `GITHUB_TOKEN`, so you do not need to create a separate repository secret for this workflow.

> [!NOTE]
> The built-in `GITHUB_TOKEN` is an installation token, so for a **private** repository it cannot download images attached via drag-and-drop (see the note above). If you need those images, use a classic PAT stored as a repository secret instead.

Next, write this workflow with the permissions required to read issues and commit generated files.

```yaml
name: Go

on:
  push:
    branches: [ "main" ]
  issues:
    types: [reopened, closed]
permissions:
  contents: write
  issues: read

jobs:

  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd # v6.0.2

    - name: Set up Go
      uses: actions/setup-go@4a3601121dd01d1626a1e23e37211e3254c1c06c # v6.4.0
      with:
        go-version: '1.25.0'
        # If your repository manages the Go version in go.mod, you can use this instead.
        # go-version-file: go.mod

    - name: Install
      run: go install github.com/rokuosan/github-issue-cms@v1.0.0

    - name: Generate
      run: github-issue-cms generate --token=${{ secrets.GITHUB_TOKEN }}

    - name: Auto Commit
      uses: stefanzweifel/git-auto-commit-action@04702edda442b2e678b25b537cec683a1493fcb9 # v7.1.0
      with:
        commit_message: "ci(github-issue-cms): :memo: Update content from GitHub Issues"
```

Congratulations.

Your Hugo site content will be regenerated and committed automatically when you push to `main` or an issue is closed or reopened.
