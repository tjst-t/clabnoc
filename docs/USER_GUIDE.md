# clabnoc ユーザーガイド

clabnoc で Containerlab トポロジを可視化・操作するためのガイド。
clab.yml のラベル設定と `.clabnoc.yml` 設定ファイルの書き方を解説する。

## 目次

1. [クイックスタート](#クイックスタート)
2. [clab.yml ラベルリファレンス](#clab-yml-ラベルリファレンス)
3. [.clabnoc.yml 設定ファイル](#clabnoc-yml-設定ファイル)
4. [SSH クレデンシャル](#ssh-クレデンシャル)
5. [BMC / noVNC 連携](#bmc--novnc-連携)
6. [障害注入](#障害注入)
7. [パケットキャプチャ](#パケットキャプチャ)
8. [実践例](#実践例)

---

## クイックスタート

### 1. clabnoc を起動する

```bash
docker run -d \
  --name clabnoc \
  --restart unless-stopped \
  --network host \
  --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /tmp/containerlab:/tmp/containerlab:ro \
  ghcr.io/tjst-t/clabnoc:latest
```

| フラグ | 理由 |
|-------|------|
| `--network host` | clab mgmt ブリッジへのアクセス、障害注入でホスト netns 操作 |
| `--privileged` | パケットキャプチャ (tcpdump)、ネットワーク操作に必要 |
| `-v docker.sock` | コンテナ検出、docker exec |
| `-v /tmp/containerlab` | topology-data.json の読み取り |

起動後、ブラウザで `http://<host>:8080` を開く。
稼働中の全 Containerlab プロジェクトが自動検出される。

### 2. clab.yml にラベルを追加する

ラベルなしでもトポロジの基本表示は可能だが、ラベルを追加することで
DC/ラック階層表示やアイコンの最適化が行える。

```yaml
topology:
  nodes:
    spine1:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "dc1"
        graph-rack: "rack-a01"
        graph-role: "spine"
```

### 3. .clabnoc.yml で詳細設定する（任意）

clab.yml のラベルだけでは足りない場合、`.clabnoc.yml` で補完・上書きできる。

---

## clab.yml ラベルリファレンス

clab.yml の各ノードに `labels` を設定することで clabnoc の表示をカスタマイズする。
全ラベルは `graph-` プレフィックスを持つ。

### 可視化制御

| ラベル | 値 | デフォルト | 説明 |
|--------|-----|-----------|------|
| `graph-hide` | `"yes"` | — | ノードをトポロジから非表示にする |
| `graph-icon` | string | 自動判定 | アイコン種別を明示指定（後述） |
| `graph-role` | string | 自動判定 | ノードの役割（後述） |

### DC/ラック構造

| ラベル | 値 | デフォルト | 説明 |
|--------|-----|-----------|------|
| `graph-dc` | string | — | データセンター名 |
| `graph-rack` | string | — | ラック名 |
| `graph-rack-unit` | integer | — | ラック内 U ポジション（1 = 最下段） |
| `graph-rack-unit-size` | integer | `1` | 機器の高さ（U 数） |

### BMC 連携

| ラベル | 値 | デフォルト | 説明 |
|--------|-----|-----------|------|
| `graph-bmc` | `"true"` | — | BMC ノードとしてマーク（noVNC アクセスを有効化） |
| `graph-bmc-pair` | string | — | BMC が管理する対象ノードの名前 |

### graph-role の値

| 値 | 意味 | アイコンへの影響 |
|-----|------|----------------|
| `spine` | スパインスイッチ | linux kind → switch アイコン |
| `leaf` | リーフスイッチ | linux kind → switch アイコン |
| `server` | サーバー | linux kind → server アイコン |
| `bmc` | ベースボード管理コントローラ | linux kind → bmc アイコン |
| `oob` | OOB 管理機器 | — |

### graph-icon の値

明示指定すると自動判定を上書きする。

| 値 | アイコン |
|-----|---------|
| `switch` | スイッチ |
| `router` | ルーター |
| `server` | サーバー |
| `host` | 汎用ホスト |
| `bmc` | BMC |

### アイコン自動判定ルール

`graph-icon` が未設定の場合、`kind` と `graph-role` から自動で決定される。

| kind | 条件 | アイコン |
|------|------|---------|
| `nokia_srlinux` / `srl` | — | switch |
| `arista_ceos` / `ceos` | — | switch |
| `crpd` | — | router |
| `vr-sros`, `vr-vmx`, `vr-xrv9k` 等 | — | router |
| `linux` | `graph-role: spine` or `leaf` | switch |
| `linux` | `graph-role: server` | server |
| `linux` | `graph-bmc: "true"` or `graph-role: bmc` | bmc |
| `linux` | image 名に `qemu-bmc` を含む | bmc |
| `linux` | その他 | host |
| その他 | — | host |

### グルーピングの動作

`graph-dc` と `graph-rack` の組み合わせにより、Cytoscape.js の compound node として階層表示される。

```
DC (graph-dc)
└── Rack (graph-rack)
    └── Node
```

| 設定パターン | 表示 |
|-------------|------|
| `graph-dc` + `graph-rack` 両方 | DC → Rack の 2 段階層 |
| `graph-dc` のみ | DC グループ内にフラット配置 |
| `graph-rack` のみ | ラックグループのみ |
| どちらも未設定 | グルーピングなし |

---

## .clabnoc.yml 設定ファイル

clab.yml のラベルを補完・上書きする設定ファイル。
ラベルに書ききれない詳細設定や、SSH クレデンシャルの指定に使う。

### 配置場所

以下の順で検索され、最初に見つかったファイルが使用される。

| 優先度 | パス | 説明 |
|--------|------|------|
| 1 | `<labDir>/../<projectName>.clabnoc.yml` | プロジェクトディレクトリの親（推奨） |
| 2 | `<labDir>/clabnoc.yml` | labdir 内（Docker マウント用） |

例: プロジェクト名が `dc-fabric`、labdir が `/tmp/containerlab/clab-dc-fabric` の場合:

```
/tmp/containerlab/
├── dc-fabric.clabnoc.yml        ← 優先度 1
└── clab-dc-fabric/
    ├── topology-data.json
    └── clabnoc.yml              ← 優先度 2
```

### 完全なスキーマ

```yaml
# ラック定義
racks:
  <rack-name>:
    dc: <dc-name>        # 所属 DC（必須）
    units: <integer>      # ラックの総 U 数（デフォルト: 42）

# kind 別 SSH デフォルト
kind_defaults:
  <kind-name>:
    ssh:
      username: <string>
      password: <string>
      port: <integer>

# ノード別設定
nodes:
  <node-name>:
    rack: <rack-name>     # ラック割り当て（graph-rack を上書き）
    unit: <integer>       # U ポジション（graph-rack-unit を上書き）
    size: <integer>       # U サイズ（graph-rack-unit-size を上書き、デフォルト: 1）
    role: <string>        # 役割（graph-role を上書き → アイコンも再判定）
    ssh:                  # ノード固有の SSH クレデンシャル
      username: <string>
      password: <string>
      port: <integer>
```

### racks セクション

ラック定義により DC 階層を構築する。
ここで定義されたラック名は `nodes` セクションの `rack` フィールドで参照できる。

```yaml
racks:
  rack-a01:
    dc: dc-tokyo
    units: 42

  rack-a02:
    dc: dc-tokyo
    units: 48       # 高さの異なるラック

  rack-b01:
    dc: dc-osaka
    units: 42
```

### nodes セクション

ノードごとにラック配置や役割を設定する。
clab.yml の `graph-*` ラベルを上書きする。

```yaml
nodes:
  spine-sw-01:
    rack: rack-a01
    unit: 42           # 最上部に配置
    size: 1
    role: spine

  compute-01:
    rack: rack-a01
    unit: 39
    size: 2            # 2U サーバー
    role: server

  gpu-node-01:
    rack: rack-a01
    unit: 31
    size: 4            # 4U GPU サーバー
    role: server
```

### ラベルと .clabnoc.yml の優先度

`.clabnoc.yml` の設定は clab.yml のラベルを**上書き**する。

| 設定項目 | clab.yml ラベル | .clabnoc.yml | 優先 |
|----------|----------------|-------------|------|
| DC | `graph-dc` | `racks.<name>.dc` | .clabnoc.yml |
| ラック | `graph-rack` | `nodes.<name>.rack` | .clabnoc.yml |
| U ポジション | `graph-rack-unit` | `nodes.<name>.unit` | .clabnoc.yml |
| U サイズ | `graph-rack-unit-size` | `nodes.<name>.size` | .clabnoc.yml |
| 役割 | `graph-role` | `nodes.<name>.role` | .clabnoc.yml |

### レイアウトバリデーション

clabnoc は起動時にラック配置を検証し、問題があれば警告を出す。

| 検証項目 | 内容 |
|----------|------|
| ラック高さ超過 | unit + size がラックの units を超えている |
| 機器重複 | 同一ラック内で U ポジションが重なっている |
| 配置未設定 | rack や unit が設定されていないノード |

---

## SSH クレデンシャル

clabnoc は 3 層のマージモデルで SSH 接続情報を解決する。

### 解決順序（優先度順）

```
1. ビルトインデフォルト（kind ごと）
        ↓ 上書き
2. .clabnoc.yml の kind_defaults
        ↓ 上書き
3. .clabnoc.yml のノード個別指定
```

各層では**明示的に設定された項目だけ**が上書きされる（部分上書き可能）。

### ビルトインデフォルト

以下の kind にはデフォルトの SSH クレデンシャルが組み込まれている。

| kind | ユーザー名 | パスワード | ポート |
|------|-----------|-----------|--------|
| `nokia_srlinux` / `srl` | admin | NokiaSrl1! | 22 |
| `arista_ceos` / `ceos` | admin | admin | 22 |
| `sonic-vs` | admin | YourPaSsWoRd | 22 |
| `linux` | root | *(空)* | 22 |
| `crpd` | root | clab123 | 22 |
| `vr-sros` | admin | admin | 22 |
| `vr-vmx` | root | Embe1mpls | 22 |
| `vr-xrv9k` | clab | clab@123 | 22 |
| `vr-veos` | admin | admin | 22 |
| `vr-csr` | admin | admin | 22 |
| `vr-n9kv` | admin | admin | 22 |

上記以外の kind は `admin` / *(空)* / 22 がデフォルト。

### .clabnoc.yml での上書き

```yaml
# kind 全体のデフォルトを変更
kind_defaults:
  nokia_srlinux:
    ssh:
      password: MyCustomPassword!    # パスワードだけ上書き

  linux:
    ssh:
      username: ubuntu               # ユーザー名だけ上書き

# 特定ノードの SSH 設定
nodes:
  jump-server:
    ssh:
      username: admin
      password: secret123
      port: 2222                     # 非標準ポート
```

---

## BMC / noVNC 連携

[qemu-bmc](https://github.com/tjst-t/qemu-bmc) を使用している場合、clabnoc から直接 noVNC コンソールを開ける。

### 自動検出

以下のいずれかの条件を満たすとノードが BMC として認識される:

1. `graph-bmc: "true"` ラベルが設定されている
2. コンテナイメージ名に `qemu-bmc` が含まれている（大文字小文字を区別しない）

BMC として認識されると:
- アイコンが BMC に変更される
- アクセス方法に **VNC** が追加される（`https://<mgmt-ipv4>/novnc/vnc.html`）
- ノード詳細パネルから noVNC を新タブで開ける

### clab.yml での設定

```yaml
topology:
  nodes:
    server1-bmc:
      kind: linux
      image: ghcr.io/tjst-t/qemu-bmc:latest
      labels:
        graph-bmc: "true"
        graph-bmc-pair: "server1"     # 管理対象ノード（メタデータ）
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "2"
```

---

## 障害注入

リンクを右クリック（コンテキストメニュー）から以下の操作が可能。

### リンクダウン/アップ

リンクの veth インターフェースを down/up する。

### tc netem

遅延・パケットロスなどのネットワーク障害をシミュレートする。

| パラメータ | 説明 | 例 |
|-----------|------|-----|
| `delay_ms` | 遅延（ミリ秒） | 100 |
| `jitter_ms` | ジッター（ミリ秒） | 10 |
| `loss_percent` | パケットロス率（%） | 5 |
| `corrupt_percent` | パケット破損率（%） | 1 |
| `duplicate_percent` | パケット重複率（%） | 0 |

### BPF フィルタ（選択的障害注入）

特定のトラフィックだけに障害を注入できる。
FaultDialog の BPF フィルタセクションでプリセットを選択するか、カスタム式を入力する。

**プリセット:**

| 名前 | フィルタ式 | 用途 |
|------|-----------|------|
| DNS | `udp port 53` | DNS トラフィックだけに障害 |
| BGP | `tcp port 179` | BGP セッションだけに障害 |
| HTTPS | `tcp port 443` | HTTPS トラフィックだけに障害 |
| HTTP | `tcp port 80` | HTTP トラフィックだけに障害 |
| ICMP | `icmp` | ICMP (ping) だけに障害 |
| ARP | `arp` | ARP だけに障害 |

BPF フィルタが設定されると、内部的に `prio` qdisc + `netem` + `tc filter` のチェーンが構成され、
マッチしたパケットだけが netem を通過する。

### リンク状態の表示

| 色 | 状態 | 説明 |
|----|------|------|
| 緑 | up | 正常 |
| 赤 | down | リンクダウン（障害注入で切断） |
| 黄 | degraded | netem 適用中 |

---

## パケットキャプチャ

リンクを右クリック（コンテキストメニュー）からパケットキャプチャを開始できる。

### pcap キャプチャ

1. リンクの右クリックメニューから **Start Capture** を選択
2. キャプチャ中はリンク詳細パネルに REC インジケーターが表示される
3. **Stop Capture** でキャプチャ停止
4. **Download Pcap** で pcap ファイルをダウンロード → Wireshark 等で解析

### ライブストリーム

1. リンクの右クリックメニューから **Live Stream** を選択
2. 画面下部にパケットテーブルが表示される
3. リアルタイムでパケットが流れてくる

**パケットテーブルの機能:**

| 操作 | 説明 |
|------|------|
| pause / resume | ストリームの一時停止・再開 |
| clear | 表示パケットのクリア |
| scroll | 自動スクロールの ON/OFF |
| Display Filter | プロトコル・IP アドレスなどのテキストフィルタ |

パケットテーブルはプロトコル別に色分けされる:
- TCP: シアン
- UDP: グリーン
- ICMP: アンバー
- ARP: レッド

---

## 実践例

### 例 1: 最小構成（ラベルなし）

ラベルがなくてもトポロジの表示は可能。ノードの kind に基づいてアイコンが自動判定される。

```yaml
name: simple-lab

topology:
  nodes:
    router1:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
    router2:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
    server1:
      kind: linux
      image: alpine:latest

  links:
    - endpoints: ["router1:e1-1", "router2:e1-1"]
    - endpoints: ["router1:e1-2", "server1:eth1"]
```

### 例 2: DC ファブリック構成

```yaml
name: dc-fabric

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

    spine2:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "dc1"
        graph-rack: "rack2"
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

    leaf2:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "dc1"
        graph-rack: "rack2"
        graph-rack-unit: "20"
        graph-role: "leaf"

    server1:
      kind: linux
      image: alpine:latest
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "1"
        graph-role: "server"

    # ユーティリティノード（非表示）
    dnsmasq:
      kind: linux
      image: dnsmasq:latest
      labels:
        graph-hide: "yes"

  links:
    - endpoints: ["spine1:e1-1", "leaf1:e1-49"]
    - endpoints: ["spine1:e1-2", "leaf2:e1-49"]
    - endpoints: ["spine2:e1-1", "leaf1:e1-50"]
    - endpoints: ["spine2:e1-2", "leaf2:e1-50"]
    - endpoints: ["leaf1:e1-1", "server1:eth1"]
```

表示イメージ:

```
┌─ DC: dc1 ──────────────────────────────────────┐
│                                                  │
│  ┌─ Rack: rack1 ──┐    ┌─ Rack: rack2 ──┐      │
│  │                 │    │                 │      │
│  │  [spine1] ──────│────│── [spine2]      │      │
│  │     │           │    │     │           │      │
│  │  [leaf1]  ──────│────│── [leaf2]       │      │
│  │     │           │    │                 │      │
│  │  [server1]      │    │                 │      │
│  │                 │    │                 │      │
│  └─────────────────┘    └─────────────────┘      │
│                                                  │
└──────────────────────────────────────────────────┘
```

### 例 3: .clabnoc.yml でラベルを補完

clab.yml にラベルを追加せず、`.clabnoc.yml` だけで配置を制御する方法。

**clab.yml** (ラベルなし):

```yaml
name: dc-fabric

topology:
  nodes:
    spine1:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
    leaf1:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
    server1:
      kind: linux
      image: alpine:latest
```

**dc-fabric.clabnoc.yml**:

```yaml
racks:
  rack-a01:
    dc: dc-tokyo
    units: 42
  rack-a02:
    dc: dc-tokyo
    units: 42

kind_defaults:
  linux:
    ssh:
      username: ubuntu
      password: ubuntu

nodes:
  spine1:
    rack: rack-a01
    unit: 42
    role: spine

  leaf1:
    rack: rack-a01
    unit: 20
    role: leaf

  server1:
    rack: rack-a02
    unit: 1
    size: 2
    role: server
    ssh:
      port: 2222
```

### 例 4: BMC 付きサーバー

```yaml
name: bmc-lab

topology:
  nodes:
    tor-sw:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "42"
        graph-role: "leaf"

    server1:
      kind: linux
      image: alpine:latest
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "1"
        graph-role: "server"

    server1-bmc:
      kind: linux
      image: ghcr.io/tjst-t/qemu-bmc:latest
      labels:
        graph-dc: "dc1"
        graph-rack: "rack1"
        graph-rack-unit: "1"
        graph-bmc: "true"
        graph-bmc-pair: "server1"

    oob-switch:
      kind: linux
      image: bridge-container:latest
      labels:
        graph-role: "oob"
        graph-icon: "switch"
```

### 例 5: マルチ DC 構成

```yaml
name: multi-dc

topology:
  nodes:
    dc1-spine:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "tokyo"
        graph-rack: "rack-t01"
        graph-role: "spine"

    dc1-leaf:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "tokyo"
        graph-rack: "rack-t01"
        graph-role: "leaf"

    dc2-spine:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "osaka"
        graph-rack: "rack-o01"
        graph-role: "spine"

    dc2-leaf:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:24.10.1
      labels:
        graph-dc: "osaka"
        graph-rack: "rack-o01"
        graph-role: "leaf"

    # DC 間接続ルーター
    wan-router:
      kind: crpd
      image: crpd:latest
      labels:
        graph-icon: "router"
```

---

## Tips

### ラベル vs .clabnoc.yml の使い分け

| 観点 | clab.yml ラベル | .clabnoc.yml |
|------|----------------|-------------|
| バージョン管理 | clab.yml と共にコミット | 別ファイルで管理可能 |
| 適用範囲 | ノードごとに個別設定 | kind 全体のデフォルト指定可能 |
| SSH 設定 | 不可 | 可能 |
| ラック U 数 | ラベルでは指定不可 | `racks.<name>.units` で指定 |
| 用途 | インフラ定義の一部として | 可視化専用の設定として |

**推奨**: DC/ラック/役割などインフラに密接な情報は clab.yml のラベルに、
SSH クレデンシャルやラック高さなど表示・接続に関する設定は `.clabnoc.yml` に書く。

### Containerlab バージョン互換

clabnoc は topology-data.json の新旧フォーマットを自動判別する。

| バージョン | リンク形式 |
|-----------|-----------|
| v0.73.0 以降 | `links[].endpoints.a` / `.z` |
| v0.73.0 未満 | `links[].a` / `.z` |

明示的な設定は不要。

### graph-hide の活用

clabnoc 自身や DNS サーバーなど、トポロジ表示に不要なノードは
`graph-hide: "yes"` で非表示にすることを推奨する。

```yaml
    clabnoc:
      kind: linux
      image: ghcr.io/tjst-t/clabnoc:latest
      labels:
        graph-hide: "yes"

    dnsmasq:
      kind: linux
      image: dnsmasq:latest
      labels:
        graph-hide: "yes"
```
