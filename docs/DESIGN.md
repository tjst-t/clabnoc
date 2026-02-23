# clabnoc — Design Document

## Overview

clabnoc はホスト常駐型の Containerlab トポロジ可視化・操作ツール。
Docker API で全 clab プロジェクトを自動検出し、ブラウザからトポロジ表示、ターミナルアクセス、障害注入を行う。

## Background

Containerlab にはトポロジを可視化・操作するための適切なツールが存在しない。

- **Netreplica Graphite**: 2023年6月以降メンテナンスされておらず、Containerlab v0.73.0 が生成する topology-data.json のリンク形式（endpoints ラッパー）に対応していない
- **TopoViewer (asadarafat)**: 内蔵 Containerlab ライブラリが v0.60.0 で古く、cap-add / devices フィールドをパースできない。JSON 入力フラグにバグあり
- **clab graph (組み込み)**: 簡易的な NeXt UI ベースのビューアのみ。永続的な Web UI として不十分

clabnoc はこれらの問題を解決し、さらに以下の機能を提供する:
- 複数 clab プロジェクトの自動検出・切替
- docker exec / SSH ターミナルアクセス
- qemu-bmc noVNC 連携
- リンク障害注入 (up/down, tc netem)
- DC/ラック単位のグルーピング表示

## Deployment

```bash
docker run -d \
  --name clabnoc \
  --restart unless-stopped \
  --network host \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /tmp/containerlab:/tmp/containerlab:ro \
  ghcr.io/tjst-t/clabnoc:latest
```

| フラグ | 理由 |
|-------|------|
| `--network host` | 全 clab mgmt ブリッジへのアクセス、障害注入でホスト netns 操作 |
| `-v docker.sock` | コンテナ検出、docker exec、障害注入 |
| `-v /tmp/containerlab` | topology-data.json の読み取り |

個人 VM 専用が前提。Docker Socket はフル root 権限委譲と同義だが許容する。

## Architecture

```
┌──────────────────────────────────────────────┐
│  clabnoc (host daemon, --network host)        │
│  :8080                                        │
│                                               │
│  ┌────────────────────┐  ┌─────────────────┐ │
│  │  Go API Server     │  │  SPA Frontend   │ │
│  │                    │  │  (go:embed)     │ │
│  │  Docker SDK        │  │  Cytoscape.js   │ │
│  │  SSH client        │  │  xterm.js       │ │
│  │  netlink (fault)   │  │  React + TS     │ │
│  └────────┬───────────┘  └─────────────────┘ │
│           │                                   │
│     Docker Socket          Host Network       │
└──────────────────────────────────────────────┘
         │
    ┌────┴─────────────────────────────┐
    │  Docker Engine                    │
    │                                   │
    │  clab-project-A/                  │
    │    ├── spine1 (nokia_srlinux)     │
    │    ├── server1 (linux + qemu-bmc) │
    │    └── br-AAAA (mgmt bridge)      │
    │                                   │
    │  clab-project-B/                  │
    │    ├── router1                    │
    │    └── br-BBBB (mgmt bridge)      │
    └───────────────────────────────────┘
```

## Project Discovery

Docker API で `containerlab` ラベルを持つコンテナを列挙し、プロジェクトごとにグルーピング。

```go
filters.Add("label", "containerlab")
containers := cli.ContainerList(ctx, opts)

projects := map[string][]ContainerInfo{}
for _, c := range containers {
    proj := c.Labels["containerlab"]
    projects[proj] = append(projects[proj], c)
}
```

topology-data.json は各コンテナの `clab-node-lab-dir` ラベルから labdir パスを推定:
`{labdir}/topology-data.json` → `/tmp/containerlab/clab-{name}/topology-data.json`

## Data Source: topology-data.json Schema

Containerlab が生成する JSON。clabnoc はこれだけをデータソースとする（clab ライブラリ非依存）。

### Top-level

```json
{
  "name": "minimal",
  "type": "clab",
  "clab": {
    "config": {
      "prefix": "clab",
      "mgmt": {
        "network": "clab",
        "bridge": "br-XXXX",
        "ipv4-subnet": "172.20.20.0/24",
        "ipv4-gw": "172.20.20.1",
        "ipv6-subnet": "3fff:172:20:20::/64",
        "ipv6-gw": "3fff:172:20:20::1"
      }
    }
  },
  "nodes": { ... },
  "links": [ ... ]
}
```

### Node

```json
{
  "index": "0",
  "shortname": "dnsmasq",
  "longname": "clab-minimal-dnsmasq",
  "fqdn": "dnsmasq.minimal.io",
  "group": "",
  "labdir": "/path/to/clab-minimal/dnsmasq",
  "kind": "linux",
  "image": "clab-iaas-dnsmasq:latest",
  "mgmt-net": "",
  "mgmt-intf": "",
  "mgmt-ipv4-address": "172.20.20.8",
  "mgmt-ipv4-prefix-length": 24,
  "mgmt-ipv6-address": "3fff:172:20:20::8",
  "mgmt-ipv6-prefix-length": 64,
  "mac-address": "",
  "labels": {
    "clab-node-kind": "linux",
    "clab-node-name": "dnsmasq",
    "containerlab": "minimal",
    "graph-hide": "yes"
  },
  "port-bindings": [
    {
      "host-ip": "0.0.0.0",
      "host-port": 6080,
      "port": 6080,
      "protocol": "tcp"
    }
  ]
}
```

### Link (v0.73.0+)

```json
{
  "endpoints": {
    "a": {
      "node": "ops",
      "interface": "eth1",
      "mac": "aa:c1:ab:04:f3:6e",
      "peer": "z"
    },
    "z": {
      "node": "sw-mgmt",
      "interface": "e1-1",
      "mac": "aa:c1:ab:9b:98:f7",
      "peer": "a"
    }
  }
}
```

### Link (旧バージョン)

```json
{
  "a": {
    "node": "ops",
    "interface": "eth1",
    "mac": "aa:c1:ab:04:f3:6e"
  },
  "z": {
    "node": "sw-mgmt",
    "interface": "e1-1",
    "mac": "aa:c1:ab:9b:98:f7"
  }
}
```

## Terminal Access

### docker exec

WebSocket で docker exec の stdin/stdout をブリッジ。

```
WS /api/v1/projects/{name}/nodes/{node}/exec?cmd=/bin/bash
```

フロントエンド: xterm.js + WebSocket addon。タブ式で複数セッション管理。

### SSH

Go の `golang.org/x/crypto/ssh` で mgmt IP に接続し、WebSocket でブリッジ。

```
WS /api/v1/projects/{name}/nodes/{node}/ssh?user=admin&port=22
```

### noVNC (qemu-bmc)

qemu-bmc は Redfish HTTPS ポート(:443) で `/novnc/vnc.html` を提供。WebSocket proxy も内蔵。
clabnoc は proxy 不要。ノードの mgmt IP から URL を組み立てて新タブで開く。

```go
func getVNCURL(node NodeInfo) string {
    if node.Labels["graph-bmc"] != "true" {
        return ""
    }
    return fmt.Sprintf("https://%s/novnc/vnc.html", node.MgmtIPv4)
}
```

## Terminal Tab Management

### プロジェクト別状態保持

プロジェクト切替時に WebSocket 接続を切断せず、バックグラウンドで維持する。

```typescript
interface TerminalTab {
  id: string;
  node: string;
  type: "exec" | "ssh";
  label: string;          // "spine1 (exec)" / "spine1 (ssh)"
  terminal: Terminal;      // xterm.js instance
  socket: WebSocket;
}

// プロジェクト別にタブ状態を保持
const tabsByProject = useRef<Map<string, TerminalTab[]>>(new Map());
const [activeProject, setActiveProject] = useState<string>("");
const [activeTabId, setActiveTabId] = useState<string | null>(null);
```

プロジェクト切替時:
- 現在のタブ群を `tabsByProject.get(currentProject)` に保存
- 非アクティブタブの xterm.js は `display:none` で隠す（DOM から外さない）
- 新プロジェクトのタブ群を復元
- WebSocket はバックグラウンドで維持

## Link Fault Injection

### 方法

ホスト側の veth インターフェースを操作。`--network host` なのでホスト netns に直接アクセス可能。

Go の netlink ライブラリ (`vishvananda/netlink`) を使用:

```go
// リンクダウン
link, _ := netlink.LinkByName(hostVethName)
netlink.LinkSetDown(link)

// リンクアップ
netlink.LinkSetUp(link)

// tc netem
qdisc := &netlink.Netem{
    QdiscAttrs: netlink.QdiscAttrs{LinkIndex: link.Attrs().Index, Handle: netlink.MakeHandle(1, 0), Parent: netlink.HANDLE_ROOT},
    Latency: 100000, // 100ms in usec
    Loss:    30,     // 30%
}
netlink.QdiscAdd(qdisc)
```

### veth peer の特定

```go
func findHostVeth(containerPID int, ifName string) (string, error) {
    // 1. /proc/{pid}/net/igmp or nsenter でコンテナ内の ifindex を取得
    // 2. ホスト側でその peer index を持つ veth を検索
}
```

## Existing Tools: Lessons Learned

### Netreplica Graphite
- NeXt UI ベース
- topology-data.json → 内部 CMT 形式に変換
- **問題**: v0.73.0 の endpoints ラッパー未対応

### TopoViewer
- Cytoscape.js + Go バックエンド
- Containerlab ライブラリを直接使用して YAML パース
- **問題**: 内蔵ライブラリ (v0.60.0) が古い、JSON 入力フラグにバグ

### 教訓
- Containerlab ライブラリに依存しない（topology-data.json のみ使用）
- リンク形式の差異を吸収する互換レイヤーを持つ

## Key Design Decisions

| 決定 | 理由 |
|-----|------|
| ホスト常駐 | 複数プロジェクト対応、ライフサイクル独立 |
| `--network host` | 全 mgmt ブリッジへの到達性、障害注入でホスト netns 操作 |
| topology-data.json のみ | clab ライブラリ依存を避ける（TopoViewer の教訓） |
| noVNC は新タブリンク | proxy 自前実装不要、qemu-bmc 側に実装済み |
| Cytoscape.js | compound node 対応、DC/ラックグルーピングに最適 |
| Go embed | フロントエンドをバイナリに埋め込み、単一バイナリ配布 |
| vishvananda/netlink | 外部コマンド (ip, tc) への依存を避ける |
