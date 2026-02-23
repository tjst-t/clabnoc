# clabnoc — API Specification

Base URL: `http://localhost:8080/api/v1`

## Projects

### GET /projects

全 clab プロジェクトの一覧。

```json
[
  {
    "name": "dc-fabric",
    "nodes": 12,
    "status": "running",
    "labdir": "/tmp/containerlab/clab-dc-fabric"
  }
]
```

`status`:
- `running`: 全ノード稼働中
- `partial`: 一部ノード停止
- `stopped`: 全ノード停止

### GET /projects/{name}/topology

トポロジデータ（topology-data.json をパース・加工）。

```json
{
  "name": "dc-fabric",
  "nodes": [
    {
      "name": "spine1",
      "kind": "nokia_srlinux",
      "image": "ghcr.io/nokia/srlinux:24.10.1",
      "status": "running",
      "mgmt_ipv4": "172.20.20.2",
      "mgmt_ipv6": "3fff:172:20:20::2",
      "container_id": "abc123...",
      "labels": {
        "graph-dc": "dc1",
        "graph-rack": "rack1",
        "graph-role": "spine"
      },
      "port_bindings": [],
      "access_methods": [
        { "type": "exec", "label": "Console (docker exec)" },
        { "type": "ssh", "label": "SSH", "target": "172.20.20.2:22" }
      ],
      "graph": {
        "dc": "dc1",
        "rack": "rack1",
        "rack_unit": 40,
        "role": "spine",
        "icon": "switch",
        "hidden": false
      }
    }
  ],
  "links": [
    {
      "id": "spine1:e1-1__leaf1:e1-49",
      "a": { "node": "spine1", "interface": "e1-1", "mac": "aa:c1:ab:..." },
      "z": { "node": "leaf1", "interface": "e1-49", "mac": "aa:c1:ab:..." },
      "state": "up",
      "netem": null
    }
  ],
  "groups": {
    "dcs": ["dc1"],
    "racks": {
      "dc1": ["rack1", "rack2"]
    }
  }
}
```

## Nodes

### GET /projects/{name}/nodes

プロジェクト内の全ノード一覧。

### GET /projects/{name}/nodes/{node}

ノード詳細情報。レスポンスは topology の nodes 要素と同一構造。

### POST /projects/{name}/nodes/{node}/action

ノード操作。

```json
{ "action": "start" }
{ "action": "stop" }
{ "action": "restart" }
```

## Terminal (WebSocket)

### WS /projects/{name}/nodes/{node}/exec

docker exec ターミナル。

Query parameters:
- `cmd`: 実行コマンド (デフォルト: `/bin/bash`)

プロトコル: バイナリ WebSocket フレーム（stdin/stdout の raw バイト列）。
xterm.js の WebSocket addon と直結。

### WS /projects/{name}/nodes/{node}/ssh

SSH ターミナル。

Query parameters:
- `user`: SSH ユーザー名 (デフォルト: `admin`)
- `port`: SSH ポート (デフォルト: `22`)
- `password`: SSH パスワード (optional)

プロトコル: exec と同一（バイナリ WebSocket フレーム）。

## Links

### GET /projects/{name}/links

全リンク一覧 + 現在の状態。

```json
[
  {
    "id": "spine1:e1-1__leaf1:e1-49",
    "a": { "node": "spine1", "interface": "e1-1" },
    "z": { "node": "leaf1", "interface": "e1-49" },
    "state": "up",
    "netem": null,
    "host_veth_a": "veth123abc",
    "host_veth_z": "veth456def"
  }
]
```

`state`:
- `up`: 正常
- `down`: 切断 (ip link down)
- `degraded`: netem 適用中

### POST /projects/{name}/links/{id}/fault

障害注入/解除。

**リンクダウン:**
```json
{ "action": "down" }
```

**リンクアップ (復旧):**
```json
{ "action": "up" }
```

**netem (遅延/パケットロス):**
```json
{
  "action": "netem",
  "netem": {
    "delay_ms": 100,
    "jitter_ms": 10,
    "loss_percent": 30,
    "corrupt_percent": 0,
    "duplicate_percent": 0
  }
}
```

**netem 解除:**
```json
{ "action": "clear_netem" }
```

### GET /projects/{name}/links/{id}

個別リンクの状態。

## Events (WebSocket)

### WS /events

Query parameters:
- `project`: プロジェクト名 (optional、省略時は全プロジェクト)

イベント種別:

```json
{ "type": "node_status_changed", "project": "dc-fabric", "data": { "node": "server1", "status": "stopped" } }
{ "type": "link_fault_changed", "project": "dc-fabric", "data": { "link_id": "spine1:e1-1__leaf1:e1-49", "state": "down" } }
{ "type": "project_changed", "data": { "action": "created", "project": "new-lab" } }
{ "type": "topology_changed", "project": "dc-fabric", "data": {} }
```
