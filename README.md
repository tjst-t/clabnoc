# clabnoc

Containerlab NOC — トポロジ可視化・ターミナルアクセス・障害注入ツール

## Features

- **プロジェクト自動検出**: Docker API でホスト上の全 Containerlab プロジェクトを検出
- **トポロジ可視化**: DC/ラック単位のグルーピング表示 (Cytoscape.js)
- **ターミナルアクセス**: docker exec / SSH / noVNC (qemu-bmc 連携)
- **障害注入**: リンク切断、遅延/パケットロス注入 (tc netem)
- **リアルタイム更新**: Docker Events 監視によるトポロジ自動更新

## Quick Start

```bash
docker run -d \
  --name clabnoc \
  --restart unless-stopped \
  --network host \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /tmp/containerlab:/tmp/containerlab:ro \
  ghcr.io/tjst-t/clabnoc:latest
```

ブラウザで http://localhost:8080 を開く。

## clab.yml Labels

clab.yml のラベルで可視化をカスタマイズ:

```yaml
topology:
  nodes:
    spine1:
      kind: nokia_srlinux
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "40"
        graph-role: "spine"

    server1:
      kind: linux
      image: qemu-bmc:latest
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-bmc: "true"
```

詳細は [docs/LABELS.md](docs/LABELS.md) を参照。

## Architecture

```
clabnoc (host daemon, --network host)
├── Go API Server (Docker SDK, SSH, netlink)
└── React SPA (Cytoscape.js, xterm.js)
```

詳細は [docs/DESIGN.md](docs/DESIGN.md) を参照。

## Development

```bash
# Build
make build

# Test
make test

# Docker
make docker-build
make docker-run
```

## Documentation

- [DESIGN.md](docs/DESIGN.md) — 設計ドキュメント
- [API.md](docs/API.md) — API 仕様
- [LABELS.md](docs/LABELS.md) — ラベル体系
- [TESTING.md](docs/TESTING.md) — テスト方針

## License

MIT
