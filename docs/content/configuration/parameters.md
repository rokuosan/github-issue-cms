---
title: 'gic.config.yaml の設定'
date: 2021-12-25
weight: 1
---

このページでは `gic.config.yaml` の設定について説明します。

## 全体構成

`gic.config.yaml` は以下のような構成になっています。

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

## 設定項目

### `github`

GitHub の設定です。

- `username`: GitHub のユーザー名
- `repository`: Issue を取得するリポジトリ名

### `hugo`

Hugo の設定です。

#### `filename`

- `articles`: 記事のファイル名
- `images`: 画像のファイル名

``[:id]`` は画像の ID に置き換わります。画像の ID はそのIssue内部で一意で、連番で割り振られます。

#### `directory`

- `articles`: 記事の保存先ディレクトリ
- `images`: 画像の保存先ディレクトリ

#### `url`

- `images`: 画像の URL

## プレースホルダ

`gic.config.yaml` では以下のプレースホルダを利用できます。

- `%Y`: 年
- `%m`: 月
- `%d`: 日
- `%H`: 時
- `%M`: 分
- `%S`: 秒

これらのプレースホルダは、`strftime` と同様の書式で利用できます。

## 設定例

### Hugo のページバンドルを使う場合

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

#### 出力例

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
