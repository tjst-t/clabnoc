# clabnoc — clab.yml Label Schema

clab.yml の `labels` フィールドで clabnoc の可視化をカスタマイズする。

## Label Reference

### 可視化制御（既存ツール互換）

| Label | Type | Default | Description |
|-------|------|---------|-------------|
| `graph-hide` | `"yes"` | - | ノードを非表示にする |
| `graph-icon` | string | auto | アイコン種別 (switch, router, server, host, bmc) |
| `graph-level` | integer | auto | 階層表示での段 (1=上段) |
| `graph-group` | string | - | 論理グループ名 |

### DC/ラック構造（clabnoc 拡張）

| Label | Type | Default | Description |
|-------|------|---------|-------------|
| `graph-dc` | string | - | データセンター名 |
| `graph-rack` | string | - | ラック名 |
| `graph-rack-unit` | integer | - | ラック内 U 位置 (1=最下段) |
| `graph-rack-height` | integer | 1 | 機器の U 数 |
| `graph-role` | string | auto | 役割 (spine, leaf, server, oob, bmc) |

### BMC 連携

| Label | Type | Default | Description |
|-------|------|---------|-------------|
| `graph-bmc` | `"true"` | - | qemu-bmc 搭載フラグ |
| `graph-bmc-pair` | string | - | BMC が管理する対象ノード名 |

## アイコン自動判定

`graph-icon` が未設定の場合、`kind` と `graph-role` から自動判定:

| kind | graph-role | アイコン |
|------|-----------|--------|
| `nokia_srlinux` | * | Switch/Router |
| `ceos` | * | Switch/Router |
| `crpd` | * | Router |
| `vr-sros`, `vr-xrv9k`, etc. | * | Router |
| `linux` | `server` | Server |
| `linux` | `bmc` or `graph-bmc=true` | BMC |
| `linux` | `spine` or `leaf` | Switch |
| `linux` | - | Generic Host |

## グルーピング動作

Cytoscape.js の compound nodes で階層表現:

```
DC (graph-dc)
└── Rack (graph-rack)
    └── Node
```

- `graph-dc` のみ設定: DC グループ内にフラットに配置
- `graph-rack` のみ設定: ラックグループで囲む（DC なし）
- 両方設定: DC → Rack の2段階層
- どちらも未設定: グルーピングなし

## 使用例

### DC ファブリック構成

```yaml
topology:
  nodes:
    spine1:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "40"
        graph-role: "spine"

    leaf1:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "20"
        graph-role: "leaf"

    server1:
      kind: linux
      image: qemu-bmc:latest
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "1"
        graph-role: "server"
        graph-bmc: "true"

    server1-bmc:
      kind: linux
      image: qemu-bmc:latest
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "1"
        graph-role: "bmc"
        graph-bmc-pair: "server1"
        graph-bmc: "true"
```

### OOB ネットワーク

```yaml
    oob-switch:
      kind: linux
      image: bridge-container:latest
      labels:
        graph-group: "management"
        graph-role: "oob"
        graph-icon: "switch"
```

### 非表示ノード (clabnoc 自身、DNS 等)

```yaml
    dnsmasq:
      kind: linux
      image: dnsmasq:latest
      labels:
        graph-hide: "yes"
```

## 表示上の効果

```
┌─ DC: dc1 ──────────────────────────────────────┐
│                                                  │
│  ┌─ Rack: rack1 ──┐    ┌─ Rack: rack2 ──┐      │
│  │                 │    │                 │      │
│  │  [spine1]  ─────│────│── [spine2]      │      │
│  │     │           │    │     │           │      │
│  │  [leaf1]        │    │  [leaf2]        │      │
│  │     │           │    │     │           │      │
│  │  [server1]      │    │  [server2]      │      │
│  │  [server1-bmc]  │    │  [server2-bmc]  │      │
│  │                 │    │                 │      │
│  └─────────────────┘    └─────────────────┘      │
│                                                  │
└──────────────────────────────────────────────────┘
```

リンクの色:
- 🟢 緑: up (正常)
- 🔴 赤: down (障害注入で切断)
- 🟡 黄: degraded (netem 適用中)
