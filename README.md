# RelaLogi Backend

イラストロジック（お絵かきロジック）の画像から、ヒント（行・列の数字）を読み取って電子化する Go 製の API サーバー。

## 概要

パズル画像とヒント領域の指定を受け取り、OpenCV で前処理 → OCR → 数字グリッドへの変換。

- 言語: Go 1.24
- HTTP フレームワーク: [gin](https://github.com/gin-gonic/gin)
- 画像処理: [GoCV](https://gocv.io/)（OpenCV）
- OCR: [gosseract](https://github.com/otiai10/gosseract)（Tesseract）
- アーキテクチャ: クリーンアーキテクチャ

## アーキテクチャ

レイヤーごとのディレクトリ構成。

| ディレクトリ | 役割 |
| --- | --- |
| `entity/` | ドメインモデル（Puzzle, Quad, Point など） |
| `usecase/` | ユースケース（inputport / interactor / gateway） |
| `controller/` | ユースケースとトランスポートの仲介 |
| `transport/rest/` | gin によるルーティング・ハンドラ |
| `infra/puzzle/` | OpenCV 画像処理・Tesseract OCR の実装 |
| `api/` | OpenAPI 定義とコード生成設定 |
| `cmd/` | エントリポイント |

## API

| メソッド | パス | 説明 |
| --- | --- | --- |
| `GET` | `/api/health` | ヘルスチェック |
| `POST` | `/api/puzzles/recognize` | 画像とヒント領域から電子化したヒントの取得 |

詳細は [`api/openapi.yaml`](api/openapi.yaml)。

## 前提

GoCV / gosseract はネイティブライブラリに依存。事前に OpenCV と Tesseract のインストールが必要。

```sh
# macOS (Homebrew)
brew install opencv tesseract
```

## 実行

```sh
go run ./cmd
```

サーバーは `:8080` で起動。

## OpenAPI コード生成

`api/openapi.yaml` からの型定義の再生成。

```sh
make oapi-gen
```

## テスト

```sh
go test ./...
```
