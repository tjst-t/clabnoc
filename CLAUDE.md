# CLAUDE.md — clabnoc Development Guide

## Project Overview

clabnoc は Containerlab トポロジの可視化・操作ツール。ホスト常駐デーモンとして動作し、Docker API で全 clab プロジェクトを自動検出、ブラウザからトポロジ表示・ターミナルアクセス・障害注入を行う。

## Quick Reference

- **Language**: Go (backend) + React/TypeScript (frontend)
- **Key dependencies**: Docker SDK, Cytoscape.js, xterm.js, chi (HTTP router)
- **Deploy**: `docker run --network host -v /var/run/docker.sock:/var/run/docker.sock:ro`
- **Docs**: `docs/DESIGN.md` (設計), `docs/API.md` (API仕様), `docs/LABELS.md` (ラベル体系), `docs/TESTING.md` (テスト方針)

## Architecture Summary

```
clab-viewer (host daemon, --network host, :8080)
├── Go API Server
│   ├── Docker SDK → プロジェクト検出、docker exec
│   ├── SSH client → mgmt IP経由のSSHアクセス
│   ├── netlink → veth操作による障害注入
│   └── static file server (go:embed)
└── React SPA (embedded)
    ├── Cytoscape.js → トポロジグラフ (DC/ラックグルーピング)
    ├── xterm.js → ターミナル (タブ式、プロジェクト別状態保持)
    └── noVNC → 新タブでqemu-bmcのnoVNCを開く
```

## Repository Structure

```
cmd/
└── clabnoc/
    └── main.go                 # エントリポイント

internal/
├── api/
│   ├── router.go               # HTTP/WSルーティング (chi)
│   ├── projects.go             # GET /api/v1/projects
│   ├── topology.go             # GET /api/v1/projects/{name}/topology
│   ├── nodes.go                # GET /api/v1/projects/{name}/nodes
│   ├── terminal.go             # WS exec/ssh
│   ├── links.go                # GET/POST リンク操作
│   └── events.go               # WS イベント通知
├── docker/
│   ├── client.go               # Docker SDK wrapper
│   ├── discovery.go            # clabプロジェクト検出
│   └── exec.go                 # docker exec WebSocket bridge
├── topology/
│   ├── parser.go               # topology-data.json パーサー
│   ├── model.go                # 内部データモデル
│   └── compat.go               # v0.73+ / 旧バージョン互換
├── network/
│   ├── veth.go                 # veth peer検出
│   ├── fault.go                # リンク障害注入
│   └── state.go                # 障害状態管理
├── ssh/
│   └── proxy.go                # SSH WebSocket proxy
└── frontend/
    └── embed.go                # go:embed

frontend/                       # React SPA
├── package.json
├── tsconfig.json
├── vite.config.ts
├── src/
│   ├── App.tsx
│   ├── components/
│   │   ├── ProjectSelector.tsx
│   │   ├── TopologyView.tsx
│   │   ├── NodePanel.tsx
│   │   ├── LinkPanel.tsx
│   │   ├── TerminalPanel.tsx   # タブ式ターミナル管理
│   │   ├── TerminalTab.tsx     # 個別ターミナルタブ
│   │   └── FaultDialog.tsx
│   ├── hooks/
│   │   ├── useTopology.ts
│   │   ├── useWebSocket.ts
│   │   ├── useProjects.ts
│   │   └── useTerminalTabs.ts  # プロジェクト別タブ状態管理
│   ├── lib/
│   │   ├── cytoscape-config.ts
│   │   └── api.ts
│   └── types/
│       └── topology.ts
└── dist/                       # ビルド成果物 → go:embed

docs/
├── DESIGN.md
├── API.md
├── LABELS.md
└── TESTING.md

Dockerfile
Makefile
```

## Development Commands

```bash
# Backend
go build -o bin/clabnoc ./cmd/clabnoc
go test ./... -v -count=1
go test ./... -race -count=1
go vet ./...

# Frontend
cd frontend && npm install
cd frontend && npm run dev          # 開発サーバー (Vite)
cd frontend && npm run build        # プロダクションビルド
cd frontend && npm run test         # Vitest
cd frontend && npm run lint         # ESLint

# All-in-one
make build          # frontend build + go build
make test           # 全テスト実行
make docker-build   # Docker イメージビルド
make dev            # 開発モード (frontend dev + go run)
make lint           # 全lint
```

## Development Phases — 自律的に Phase 3 まで一気通貫で実装する

各 Phase 完了時にテストを通してから次の Phase に進むこと。
テストが通らない場合は修正してから進む。途中で止まらない。

### Phase 1: Core (MVP)

**ゴール**: プロジェクト検出 + トポロジ表示 + docker exec ターミナル

実装順序:
1. Go プロジェクト初期化 (go mod, cmd/clabnoc/main.go)
2. Docker SDK wrapper + clabプロジェクト検出
3. topology-data.json パーサー (v0.73+ と旧形式両対応)
4. REST API: projects, topology, nodes
5. React SPA 初期化 (Vite + TypeScript)
6. Cytoscape.js トポロジ表示 (DC/ラック compound nodes)
7. ノード詳細サイドパネル
8. WebSocket: docker exec ターミナル
9. xterm.js タブ式ターミナル (プロジェクト別状態保持)
10. go:embed でフロントエンド埋め込み

テスト要件:
- `internal/topology/parser_test.go`: JSON パース (両形式)
- `internal/docker/discovery_test.go`: プロジェクト検出ロジック (mock)
- `internal/api/*_test.go`: 各APIエンドポイントのHTTPテスト
- `frontend/src/**/*.test.ts`: コンポーネント単体テスト
- 手動テスト: 実際のclabトポロジで画面確認

### Phase 2: Extended Access

**ゴール**: SSH + noVNC + 障害注入

実装順序:
1. SSH WebSocket proxy (golang.org/x/crypto/ssh)
2. noVNC統合 (新タブでqemu-bmc noVNC URLを開く)
3. アクセス方法の自動検出 (exec/ssh/vnc)
4. リンク状態取得 (veth peer 検出)
5. リンク障害注入 API (ip link down/up)
6. フロントエンド: リンク右クリックメニュー、リンク色による状態表示
7. フロントエンド: SSH ターミナルタブ対応

テスト要件:
- `internal/ssh/proxy_test.go`: SSH接続ロジック
- `internal/network/veth_test.go`: veth検出ロジック
- `internal/network/fault_test.go`: 障害注入/解除
- `internal/api/links_test.go`: リンクAPI
- フロントエンド: リンク操作UIのテスト

### Phase 3: Advanced

**ゴール**: netem + イベント + ノード操作

実装順序:
1. tc netem による遅延/パケットロス注入
2. Docker Events API 監視 → WebSocket イベント通知
3. フロントエンド: リアルタイムトポロジ更新
4. ノード start/stop 操作 API
5. 障害注入パラメータ設定ダイアログ (delay_ms, loss_percent)
6. FaultDialog UI の洗練

テスト要件:
- `internal/network/fault_test.go`: netem操作テスト追加
- `internal/api/events_test.go`: イベントストリーミング
- 統合テスト: Docker Compose でclabnoc + テスト用clabトポロジを起動

## 重要な設計判断

### topology-data.json の取得
clabライブラリを直接使わない（TopoViewer の教訓: 内蔵ライブラリが古くなる問題）。
topology-data.json を読むだけにする。取得方法:
- デフォルト: `/tmp/containerlab/clab-{name}/topology-data.json`
- Docker起動時に `-v /tmp/containerlab:/tmp/containerlab:ro` をバインド
- フォールバック: Docker API `CopyFromContainer` で取得

### ターミナルタブのプロジェクト別状態保持
プロジェクト切替時:
- WebSocket接続は切断しない（バックグラウンド維持）
- xterm.js インスタンスは `display:none` で隠す
- プロジェクトに戻ると復元される
- タブ状態は `Map<projectName, TerminalTab[]>` で管理

### noVNC
qemu-bmc が Redfish HTTPS ポート(:443) で `/novnc/vnc.html` を提供済み。
clabnoc 側では proxy 不要。ノードの mgmt IP + labels から URL を組み立てて新タブで開く。

### 障害注入
ホスト側の veth インターフェースを操作:
- `--network host` なのでホスト netns に直接アクセス可能
- Go の netlink ライブラリ (vishvananda/netlink) で ip link / tc netem を操作
- iproute2 コマンドへの依存を避ける

### Containerlab バージョン互換
topology-data.json のリンク形式:
- v0.73.0+: `links[].endpoints.a` / `links[].endpoints.z`
- 旧バージョン: `links[].a` / `links[].z`
両方をパースできるようにする。

## コーディング規約

- Go: 標準的な Go スタイル、`go vet` / `golangci-lint` でチェック
- エラーハンドリング: エラーは wrap して返す (`fmt.Errorf("...: %w", err)`)
- ログ: `slog` パッケージ (Go 1.21+)
- フロントエンド: ESLint + Prettier、React hooks ルール遵守
- テスト: テーブル駆動テスト、mockは interface で注入
- コミットメッセージ: Conventional Commits (feat:, fix:, docs:, test:, refactor:)

## テスト戦略の原則

詳細は `docs/TESTING.md` を参照。

- 各 Phase 完了時に全テストが通ること
- Docker SDK のテストは interface mock で実施
- ネットワーク操作 (veth, netem) は root 権限が必要なため、CI では mock テスト、手動で実機テスト
- フロントエンドは Vitest + React Testing Library
- 統合テストは Docker Compose で clabnoc + テスト用 clab トポロジを起動

## 参考資料

- Containerlab topology-data.json スキーマ: `docs/DESIGN.md` の Data Source セクション
- qemu-bmc: https://github.com/tjst-t/qemu-bmc (noVNC は `/novnc/vnc.html`)
- 既存ツールの問題点: `docs/DESIGN.md` の Existing Tools セクション
