# clabnoc — Testing Strategy

## 原則

- 各 Phase 完了時に全テストが通ること
- テストが失敗したら修正してから次に進む
- mock を活用して外部依存（Docker, ネットワーク）を分離

## テスト種別

### 1. Unit Tests (Go)

```bash
go test ./... -v -count=1
go test ./... -race -count=1
```

#### Phase 1

| パッケージ | テスト内容 |
|-----------|----------|
| `internal/topology` | topology-data.json パース (v0.73+ / 旧形式) |
| `internal/topology` | graph-hide フィルタリング |
| `internal/topology` | ラベルからグルーピング構造の構築 |
| `internal/docker` | プロジェクト検出ロジック (Docker client mock) |
| `internal/api` | 各 REST エンドポイントの HTTP テスト (httptest) |

#### Phase 2

| パッケージ | テスト内容 |
|-----------|----------|
| `internal/ssh` | SSH 接続パラメータ構築 |
| `internal/network` | veth peer 検出ロジック (mock) |
| `internal/network` | 障害注入/解除のコマンド構築 |
| `internal/api` | リンク操作 API テスト |

#### Phase 3

| パッケージ | テスト内容 |
|-----------|----------|
| `internal/network` | netem パラメータ構築 |
| `internal/api` | イベントストリーミング (WebSocket テスト) |

### 2. Frontend Tests

```bash
cd frontend && npm run test
```

Vitest + React Testing Library:

| コンポーネント | テスト内容 |
|--------------|----------|
| `ProjectSelector` | プロジェクト一覧表示、選択イベント |
| `TopologyView` | Cytoscape.js 初期化、ノード/リンク描画 |
| `NodePanel` | ノード詳細表示、アクセスボタン |
| `TerminalPanel` | タブ追加/削除/切替 |
| `useTerminalTabs` | プロジェクト別状態保持 |
| `LinkPanel` | リンク情報表示、障害注入 UI |
| `FaultDialog` | netem パラメータ入力バリデーション |
| `api.ts` | API クライアント (MSW でモック) |

### 3. Docker Client Mock

Docker SDK の操作を interface で抽象化し、テスト時は mock を注入:

```go
type DockerClient interface {
    ContainerList(ctx context.Context, opts container.ListOptions) ([]types.Container, error)
    ContainerInspect(ctx context.Context, id string) (types.ContainerJSON, error)
    ContainerExecCreate(ctx context.Context, id string, config container.ExecOptions) (types.IDResponse, error)
    ContainerExecAttach(ctx context.Context, id string, config container.ExecStartOptions) (types.HijackedResponse, error)
    CopyFromContainer(ctx context.Context, id, path string) (io.ReadCloser, types.ContainerPathStat, error)
    Events(ctx context.Context, opts events.ListOptions) (<-chan events.Message, <-chan error)
}
```

### 4. Test Fixtures

topology-data.json のテストフィクスチャを `testdata/` に配置:

```
testdata/
├── topology-v073.json      # v0.73.0+ 形式 (endpoints wrapper)
├── topology-legacy.json    # 旧形式 (flat a/z)
├── topology-with-groups.json  # DC/ラックラベル付き
├── topology-with-bmc.json     # qemu-bmc ノード付き
└── topology-minimal.json      # 最小構成 (1 node, 0 links)
```

### 5. Integration Tests

Docker Compose で clabnoc + テスト用 clab トポロジを起動して E2E テスト。

```yaml
# docker-compose.test.yml
services:
  clabnoc:
    build: .
    network_mode: host
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /tmp/containerlab:/tmp/containerlab:ro
```

テストシナリオ:
1. clab トポロジを `clab deploy` で起動
2. clabnoc の `/api/v1/projects` でプロジェクトが検出されること
3. トポロジデータが正しくパースされること
4. docker exec WebSocket でターミナルが開けること
5. リンク障害注入/解除が動作すること

統合テストは CI で自動実行は困難（Docker-in-Docker + clab が必要）なため、手動テスト用のスクリプトとして提供:

```bash
# tests/integration/run.sh
#!/bin/bash
set -e

# 1. テスト用 clab トポロジを起動
sudo clab deploy -t tests/integration/test-topology.clab.yml

# 2. clabnoc を起動
docker compose -f docker-compose.test.yml up -d

# 3. テスト実行
go test ./tests/integration/... -v -tags=integration

# 4. クリーンアップ
sudo clab destroy -t tests/integration/test-topology.clab.yml
docker compose -f docker-compose.test.yml down
```

## CI (GitHub Actions)

```yaml
# .github/workflows/test.yml
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - uses: actions/setup-node@v4
        with: { node-version: '22' }
      - run: go test ./... -v -count=1 -race
      - run: cd frontend && npm ci && npm run test && npm run lint
```

## Lint

```bash
# Go
go vet ./...
golangci-lint run

# Frontend
cd frontend && npm run lint
```
