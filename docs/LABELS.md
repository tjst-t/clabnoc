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

---

## .clabnoc.yml — 外部エンティティ設定

`.clabnoc.yml` でトポロジ上に外部ノード・ネットワーク・リンクを追加表示できる。

### auto_mgmt

clab ノードの管理ネットワーク接続を自動生成する。

```yaml
auto_mgmt:
  enabled: true        # 有効化 (default: false)
  position: bottom     # 表示位置: "top" | "bottom" (default: "bottom")
  collapsed: true      # 折りたたみ表示で ×N バッジ (default: true)
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | 管理ネットワーク表示を有効化 |
| `position` | string | `"bottom"` | ネットワークバーの表示位置 |
| `collapsed` | bool | `true` | `true`: 1本のバーに ×N バッジ / `false`: ノードごとに個別リンク |

各ノードの `mgmt-net` フィールドからネットワーク名を取得。未設定の場合 `clab.config.mgmt.network` → `"clab"` の順でフォールバック。

### external_nodes

Containerlab 管理外の外部デバイス（NTP, DNS, OOB スイッチ等）。

```yaml
external_nodes:
  ntp-server:
    label: NTP Server
    description: Campus NTP service
    icon: service           # default: "service"
    interfaces:
      - eth0
    placement:
      dc: dc-a              # DC レベル配置 → Services エリアに表示
  oob-switch:
    label: OOB Switch
    icon: switch
    interfaces:
      - ge-0/0/0
      - ge-0/0/1
    placement:
      dc: dc-a
      rack: rack-a01         # ラック内配置
      rack_unit: 20
      size: 1                # default: 1
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `label` | string | required | 表示名 |
| `description` | string | - | 説明テキスト（詳細パネルに表示） |
| `icon` | string | `"service"` | アイコン種別 |
| `interfaces` | []string | - | インターフェース名一覧（外部リンク用） |
| `placement.dc` | string | required | 配置先 DC 名 |
| `placement.rack` | string | - | 配置先ラック名（省略で Services エリア） |
| `placement.rack_unit` | int | - | ラック内 U 位置 |
| `placement.size` | int | `1` | 占有 U 数 |

配置ルール:
- `dc` のみ: DC ボックス内の Services エリアに水平配置
- `dc` + `rack` + `rack_unit`: ラック内に clab ノードと同様に配置

### external_networks

トポロジ上部（クラウド形状）または下部（バー形状）に表示する外部ネットワーク。

```yaml
external_networks:
  internet:
    label: Internet
    position: top            # クラウド形状で DC ボックス上部に表示
  oob-mgmt:
    label: OOB Management
    position: bottom         # バー形状で DC ボックス下部に表示
    dc: dc-a                 # 特定 DC に紐付け（省略で全 DC 共通）
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `label` | string | required | 表示名 |
| `position` | string | `"bottom"` | `"top"`: クラウド形状 / `"bottom"`: バー形状 |
| `dc` | string | - | 紐付け先 DC（省略で全体表示） |

### external_links

clab ノード、外部ノード、外部ネットワーク間の接続。破線で表示される。

```yaml
external_links:
  - a:
      node: spine1           # clab ノード
      interface: e1-48
    z:
      network: internet      # 外部ネットワーク
  - a:
      external: oob-switch   # 外部ノード
      interface: ge-0/0/0
    z:
      network: oob-mgmt
  - a:
      external: ntp-server
      interface: eth0
    z:
      node: leaf1
      interface: e1-10
```

各エンドポイントには以下のいずれか1つを指定:

| Field | Description |
|-------|-------------|
| `node` | clab ノード名 |
| `external` | 外部ノード名 |
| `network` | 外部ネットワーク名 |

`interface` は `node` または `external` エンドポイントに任意で指定。

### 表示上の効果（外部エンティティ）

```
          ☁ Internet ☁
              │
┌─ DC: dc1 ──│────────────────────────────────┐
│             │                                │
│  ┌ SERVICES ─────────────────────────┐       │
│  │  [NTP Server]    [DNS Server]     │       │
│  └───────────────────────────────────┘       │
│                                              │
│  ┌─ Rack: rack1 ──┐  ┌─ Rack: rack2 ──┐    │
│  │  [spine1] ──────│──│── [spine2]     │    │
│  │  [OOB Switch]┄┄┄│┄┄│┄┄┄┄┄┄┄┄┄┄┄┄┄┄ │    │
│  │  [server1]      │  │  [server2]     │    │
│  └─────────────────┘  └───────────────-┘    │
│                                              │
│  ┄┄┄┄┄ OOB Management ┄┄┄┄┄┄ ×4 ┄┄┄┄┄     │
│  ┄┄┄┄┄ clab (mgmt) ┄┄┄┄┄┄┄┄┄ ×6 ┄┄┄┄┄     │
│                                              │
└──────────────────────────────────────────────┘
```

- 外部ノード: 破線ボーダー、グレー配色、ステータス LED なし
- 外部ネットワーク (top): クラウド形状（破線）
- 外部ネットワーク (bottom): 水平バー（破線）、折りたたみ時は ×N バッジ
- 外部リンク: 破線、グレー（右クリックメニューなし）
- Services エリア: DC ボックス内上部に薄い破線境界
