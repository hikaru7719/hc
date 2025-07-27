# HC (HTTP Client) Design Document

## プロジェクト概要

HC は Go ベースの CLI ツールで、`hc serve` コマンドによりローカルサーバーを起動し、ブラウザベースの GUI HTTP クライアントを提供します。

## アーキテクチャ

### システム構成
- **バックエンド**: Go 製 Web サーバー（HTTP プロキシ機能付き）
- **フロントエンド**: Next.js による SPA（静的エクスポート）
- **データストア**: SQLite によるリクエスト履歴の永続化
- **配布形式**: go:embed で静的ファイルをバイナリに埋め込んだ単一実行ファイル

### プロジェクト構造
```
hc/
├── main.go              # CLI エントリーポイント
├── go.mod              # Go モジュール定義
├── cmd/
│   └── serve.go        # serve コマンドの実装
├── internal/
│   ├── server/
│   │   └── server.go   # Web サーバーの実装
│   ├── proxy/
│   │   └── proxy.go    # HTTP リクエスト代理実行
│   ├── storage/
│   │   └── sqlite.go   # SQLite データベース操作
│   └── models/
│       └── request.go  # データモデル定義
├── frontend/           # Next.js プロジェクト
│   ├── package.json
│   ├── pages/
│   ├── components/
│   └── out/           # 静的ビルド出力（.gitignore）
└── embed.go           # go:embed 定義
```

## 主要機能

### 1. HTTP リクエストの作成・送信
- URL、メソッド、ヘッダー、ボディの指定
- リクエストの保存と履歴管理
- レスポンスの表示（ステータスコード、ヘッダー、ボディ）

### 2. 階層構造でのリクエスト管理
- フォルダー形式での整理（例: A > B > request）
- ドラッグ&ドロップでの整理
- リクエストのグルーピングと検索

### 3. API エンドポイント
- `POST /api/request` - HTTP リクエストの実行
- `GET /api/requests` - 保存されたリクエスト一覧の取得
- `GET /api/requests/:id` - 特定リクエストの取得
- `POST /api/requests` - リクエストの保存
- `PUT /api/requests/:id` - リクエストの更新
- `DELETE /api/requests/:id` - リクエストの削除
- `GET /api/folders` - フォルダー構造の取得
- `POST /api/folders` - フォルダーの作成
- `PUT /api/folders/:id` - フォルダーの更新
- `DELETE /api/folders/:id` - フォルダーの削除

## データモデル

### SQLite スキーマ

```sql
-- フォルダー管理
CREATE TABLE folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    parent_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- HTTP リクエスト
CREATE TABLE requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    folder_id INTEGER,
    method TEXT NOT NULL,
    url TEXT NOT NULL,
    headers TEXT, -- JSON 形式
    body TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
);
```

## 技術スタック

### バックエンド
- Go 1.24+
- Cobra (CLI フレームワーク)
- 標準 net/http パッケージ
- SQLite3 (github.com/mattn/go-sqlite3)
- go:embed (静的ファイル埋め込み)

### フロントエンド
- Next.js 14+
- TypeScript
- React 18+
- TailwindCSS
- DaisyUI (UI コンポーネントライブラリ)

## ビルドプロセス

1. フロントエンドのビルド
   ```bash
   cd frontend
   npm install
   npm run build
   npm run export
   ```

2. Go バイナリのビルド
   ```bash
   go build -o hc
   ```

3. 配布
   - 単一の実行ファイル `hc` として配布
   - SQLite データベースはユーザーのホームディレクトリに自動作成

## 使用方法

```bash
# サーバーの起動（デフォルトポート: 8080）
hc serve

# カスタムポートでの起動
hc serve --port 3000

# ヘルプの表示
hc --help
```
