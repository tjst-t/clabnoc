# GNS3 vs clabnoc 機能比較レポート

調査日: 2026-02-27

## 概要

GNS3 はネットワーク機器のエミュレーション環境を提供するオープンソースツール。clabnoc は Containerlab トポロジの可視化・操作ツール。
両者の機能を比較し、clabnoc に追加すべき機能を特定する。

---

## 1. GNS3 機能一覧

### 1.1 トポロジ可視化・管理

| 機能 | 詳細 |
|------|------|
| ドラッグ＆ドロップ | デバイスをツールバーからワークスペースにドラッグ |
| ズーム/パン | 拡大・縮小・スクロール |
| スクリーンショット | ワークスペースの画面キャプチャ |
| 背景画像 | フロアプラン等をワークスペース背景に配置 |
| グリッド表示 | グリッドサイズ設定可能、スナップ対応 |
| レイヤー管理 | ビジュアル要素の重ね順序制御 |
| トポロジサマリ | テーブル形式で全ノード・リンクの状態表示 |
| インターフェースラベル | リンク上にポート名 (`f0/0`, `e0/1`) を表示/非表示 |
| `.gns3` プロジェクトファイル | JSON形式でノード位置、リンク、描画オブジェクトを保存 |
| アノテーション・描画 | テキストメモ（フォント・色・サイズ設定可能）、矩形・楕円（塗りつぶし・枠線設定可能）、線 |

### 1.2 デバイス/ノード管理

| 機能 | 詳細 |
|------|------|
| エミュレーションバックエンド | Dynamips (Cisco IOS), QEMU/KVM, Docker, VirtualBox, VMware, VPCS, IOU/IOL |
| デバイスカテゴリ | ルーター、スイッチ、エンドデバイス、セキュリティデバイス |
| ノードライフサイクル | Start, Stop, Suspend, Restart（個別 + 一括操作） |
| ステータスインジケータ | 緑(running), 黄(suspended), 赤(stopped) |
| ノード設定 | ホスト名、コンソールタイプ、RAM、ディスク、NIC数、Idle-PC |
| カスタムシンボル | Classic / Affinity アイコンセット、SVG/PNG カスタムアイコン対応 |
| アプライアンスシステム | GNS3 Marketplace からテンプレートダウンロード (`.gns3a` ファイル) |
| イメージマネージャ (v3.0) | QEMU/Dynamips/IOU イメージの一元管理 |

### 1.3 コンソール/ターミナルアクセス

| 機能 | 詳細 |
|------|------|
| プロトコル | Telnet, VNC, SPICE |
| Console Connect to All | 全ノードに一括ターミナル接続 |
| 外部ターミナル | PuTTY, SecureCRT, iTerm2 等の外部アプリ連携 |
| プロトコルハンドラ | `gns3+telnet://`, `gns3+vnc://`, `gns3+spice://` |
| Web Client Pack | ブラウザからローカルコンソールアプリを起動 |
| コンソールポート | 自動割当 or 手動指定、Auxiliary ポート対応 |

### 1.4 ネットワークシミュレーション

| 機能 | 詳細 |
|------|------|
| リンク作成 | ドラッグ＆ドロップ、インターフェース選択ダイアログ |
| リンクステータス | 色別表示 (緑/黄/赤) |
| パケットフィルタ | Frequency Drop, Packet Loss (%), Delay/Jitter (ms), Corruption (%), BPF フィルタ |
| フィルタ組み合わせ | 複数フィルタを1リンクに同時適用可能 |
| リンク中断 | ケーブル切断シミュレーション |
| パケットキャプチャ | リンク右クリック → `.pcap` 保存 → Wireshark 自動起動 |
| NAT ノード | 内蔵DHCP付きのNAT接続 (192.168.122.0/24) |
| Cloud ノード | ホストネットワークへのブリッジ |
| L2スイッチ/ハブ | ポート数設定可能、VLAN対応 |
| ATM/Frame Relay | レガシーWANプロトコルシミュレーション |

### 1.5 コンフィグ管理

| 機能 | 詳細 |
|------|------|
| Startup Config | ノードごとの起動時コンフィグ保存 |
| コンフィグエクスポート/インポート | 全ノードの startup-config を一括エクスポート/インポート |
| NVRAM/ディスク永続化 | セッション間でコンフィグ・ディスク内容保持 |

### 1.6 コラボレーション

| 機能 | 詳細 |
|------|------|
| マルチユーザー/マルチクライアント | 1サーバーに複数GUI同時接続 |
| リアルタイム同期 | ノード追加・リンク変更が全クライアントに即時反映 |
| 共有コンソール | 同一デバイスのターミナルを複数ユーザーが共有 |
| RBAC (v3.0) | ロールベースアクセス制御 |

### 1.7 インポート/エクスポート

| 機能 | 詳細 |
|------|------|
| ポータブルプロジェクト | `.gns3project` アーカイブ（ディスクイメージ含む） |
| プロジェクト API | `POST /export`, `POST /import` |
| アプライアンスインポート | `.gns3a` テンプレートファイル |
| シンボルインポート | カスタムアイコンの追加 |

### 1.8 ラボ/プロジェクト管理

| 機能 | 詳細 |
|------|------|
| プロジェクト操作 | 作成、開く、閉じる、削除、複製 |
| スナップショット | 名前付きスナップショットの作成・復元・削除 |
| GNS3 VM | VMware/Hyper-V/VirtualBox 上の専用コンピュートVM |
| サーバーモード | ローカル、GNS3 VM、リモートサーバー |

### 1.9 UI/UX

| 機能 | 詳細 |
|------|------|
| 右クリックメニュー | ノード/リンク/アノテーション/ワークスペース別のコンテキストメニュー |
| ツールバー | プロジェクト管理、スナップショット、ラベル、一括操作、アノテーション、ズーム |
| コンソールパネル | ログ/デバッグメッセージの表示 |
| サーバーサマリ | 接続状態、CPU/RAM使用率 |
| 設定画面 | プロジェクトディレクトリ、コンソール設定、エミュレータ設定 |
| Web UI (v3.0) | ブラウザベースのトポロジビルダー、ダークテーマ |

### 1.10 モニタリング・デバッグ

| 機能 | 詳細 |
|------|------|
| ステータスモニタリング | ノード/リンクのリアルタイム状態表示 |
| サーバーリソース | CPU/RAM使用率の表示 |
| ログ表示 | サーバーログ、エラー、デバッグ出力 |
| 通知システム | WebSocket ベースのイベントストリーム |
| パケットキャプチャ | ライブキャプチャ + pcap 保存 |
| デバッグAPI | `/v2/debug` エンドポイント |

---

## 2. clabnoc 現在の機能一覧

### 2.1 バックエンド (Go)

#### API エンドポイント
- `GET /api/v1/projects` — プロジェクト一覧（ノード数・ステータス付き）
- `GET /api/v1/projects/{name}/topology` — トポロジ取得（ノード・リンク・グループ）
- `GET /api/v1/projects/{name}/nodes` — ノード一覧
- `GET /api/v1/projects/{name}/nodes/{node}` — ノード詳細
- `POST /api/v1/projects/{name}/nodes/{node}/action` — ノード操作 (start/stop/restart)
- `GET /api/v1/projects/{name}/nodes/{node}/ssh-credentials` — SSH認証情報の解決
- `GET /api/v1/projects/{name}/links` — リンク一覧（障害状態付き）
- `GET /api/v1/projects/{name}/links/{id}` — リンク詳細
- `POST /api/v1/projects/{name}/links/{id}/fault` — 障害注入制御

#### WebSocket エンドポイント
- `GET /api/v1/projects/{name}/nodes/{node}/exec` — Docker exec ターミナル
- `GET /api/v1/projects/{name}/nodes/{node}/ssh` — SSH ターミナル
- `GET /api/v1/events` — Docker イベントストリーム

#### Docker 統合
- Containerlab プロジェクト自動検出（Docker ラベルベース）
- コンテナ状態取得・ID解決
- Docker exec WebSocket ブリッジ（stdin/stdout/stderr）
- ターミナルリサイズ対応
- Docker Events リアルタイム監視

#### トポロジパーサー
- v0.73+ (endpoints wrapper) と旧形式 (flat a/z) 両対応
- `.clabnoc.yml` 設定: ラック定義、ノード配置、SSH認証情報
- グラフアノテーション: DC, ラック, ユニット, サイズ, ロール, アイコン

#### 障害注入
- リンク状態制御: `ip link set down/up`
- tc netem: delay, jitter, loss, corruption, duplication
- 状態追跡と解除

#### SSH アクセス
- WebSocket SSH プロキシ
- パスワード認証 + keyboard-interactive フォールバック
- 13種以上のノードkindにデフォルト認証情報内蔵
- 3層の認証情報解決（内蔵 → kind_defaults → ノード固有）

#### noVNC プロキシ
- HTTP/WebSocket リバースプロキシ（TLS/平文自動検出）
- Authorization ヘッダキャッシュ

### 2.2 フロントエンド (React/TypeScript)

#### メインUI
- プロジェクトセレクタ（ドロップダウン）
- ダーク/ライトテーマ切替
- ノード/リンク数表示

#### トポロジビュー
- ラックベース可視化（デバイスフェースプレート + ポート）
- データセンターグルーピング
- ケーブル描画（状態別色分け）
- ズーム & パン（マウスホイール、ドラッグ）
- デバイス選択 → 接続ケーブルハイライト
- ポートレベル選択
- 情報オーバーレイ（選択デバイス情報）
- リンク右クリックメニュー（Down/Up, Netem 適用/クリア）
- 設定警告表示

#### 詳細パネル
- ノード詳細: 名前, kind, image, ステータス, mgmt IP, コンテナID, ポートバインディング
- アクセスボタン: exec, ssh, vnc
- ノード制御: start/stop/restart
- リンク詳細: 状態, エンドポイントA/Z, MAC, Netem パラメータ, 障害注入コントロール

#### ターミナルパネル
- タブ式ターミナル（exec / ssh）
- プロジェクト別タブ状態保持
- xterm.js (WebSocket, リサイズ対応)
- 折りたたみ/展開

#### ダイアログ
- SSH接続ダイアログ（デフォルト認証情報ロード、手動入力、パスワード表示切替）
- Netem ダイアログ（delay, jitter, loss, corruption, duplication）

#### UI機能
- レスポンシブデザイン
- リサイズ可能パネル
- Tailwind CSS + NOCカラーパレット
- JetBrains Mono / Space Grotesk / DM Sans フォント

### 2.3 設定

#### .clabnoc.yml
- ラック定義（DC名、ユニット数）
- ノード配置（ラック、ユニット、サイズ、ロール）
- SSH認証情報（kind_defaults、ノード固有）

### 2.4 デプロイ
- Docker コンテナ (`--network host`, Docker ソケットマウント)
- `-addr`, `-dev`, `-version` フラグ
- go:embed によるフロントエンド埋め込み

---

## 3. 機能差分分析

### 3.1 GNS3 にあって clabnoc にない機能

以下、clabnoc のコンテキスト（Containerlab ビューア/オペレータ）で意味がある機能のみ抽出。
エミュレーション固有の機能（Dynamips, QEMU VM作成, VirtualBox連携等）は対象外。

#### パケットキャプチャ

GNS3 ではリンク右クリックから `.pcap` キャプチャを開始し、Wireshark で自動表示する。
clabnoc は `--network host` で veth に直接アクセスできるため、`tcpdump` で同等のキャプチャが可能。

**実装案**:
- API: `POST /api/v1/projects/{name}/links/{id}/capture` (start/stop)
- バックエンド: veth インターフェースに対して `tcpdump` を起動
- フロントエンド: キャプチャ状態表示 + `.pcap` ファイルダウンロード
- 発展: WebSocket でリアルタイムパケットストリーム表示（ブラウザ内デコード）

#### インターフェースラベル表示

GNS3 ではトポロジ上のリンクにポート名を表示/非表示できる。

**実装案**:
- トポロジビューのケーブル上 or ホバーでインターフェース名表示
- トグルスイッチで表示/非表示切替
- ラック図のポート名にも対応

#### ノード検索・フィルタ

大規模トポロジでのノード検索機能。

**実装案**:
- ヘッダーに検索バー追加
- 名前/Kind/ステータスでフィルタ
- 検索結果をトポロジ上でハイライト + 自動スクロール

#### コンフィグ閲覧・エクスポート

ノードの running-config を取得・表示・ダウンロードする機能。

**実装案**:
- API: `GET /api/v1/projects/{name}/nodes/{node}/config`
- バックエンド: Kind 別にコンフィグ取得コマンドを定義
  - SR Linux: `sr_cli info flat`
  - cEOS: `Cli -p 15 -c "show running-config"`
  - SONiC: `show runningconfiguration all`
- フロントエンド: シンタックスハイライト付きコンフィグビューア
- 一括エクスポート（zip ダウンロード）

#### Console Connect to All（一括ターミナル接続）

全ノードまたは選択ノードに対して一括でターミナルタブを作成する。

**実装案**:
- ツールバーにボタン追加
- 選択ノードがある場合は選択分のみ、なければ全ノード
- exec / ssh の選択ダイアログ

#### スナップショット

トポロジの状態を保存・復元する。

**実装案**:
- `containerlab save` コマンド相当の機能
- コンフィグスナップショットの保存・復元
- スナップショット間の差分表示

#### リソースモニタリング

各コンテナのCPU/メモリ使用率をリアルタイム表示する。

**実装案**:
- API: `GET /api/v1/projects/{name}/stats`
- バックエンド: Docker Stats API (`ContainerStats`) でストリーミング
- フロントエンド: ノード上にミニチャート or ヒートマップ表示
- テーブルビューでソート可能な一覧

#### トポロジアノテーション

トポロジ図上にテキストメモや図形を配置する。

**実装案**:
- テキスト、矩形、線の描画ツール
- 色、フォント、サイズの設定
- `.clabnoc.yml` にアノテーション情報を保存
- 障害対応手順や設計メモの記録用途

#### BPF パケットフィルタ

特定トラフィックのみをドロップする精密な障害注入。

**実装案**:
- `tc filter` + BPF で特定プロトコル/ポート/IP のパケットをドロップ
- 既存の netem と組み合わせ可能
- プリセット: DNS ドロップ、BGP ドロップ、特定 VLAN ドロップ等

#### ノードステータス一覧（テーブルビュー）

トポロジ図とは別に、テーブル形式で全ノードの状態を一覧表示する。

**実装案**:
- 新しいビューモード or サイドパネル
- カラム: 名前, Kind, ステータス, mgmt IP, ラック, ユニット
- ソート・フィルタ対応
- 行クリックでトポロジ上のノードにフォーカス

#### マルチユーザー/RBAC

複数ユーザーの同時接続とアクセス制御。

**実装案**:
- 認証: OAuth2 / OIDC or ベーシック認証
- 権限: viewer (閲覧のみ), operator (操作可), admin (設定変更可)
- セッション管理
- 操作ログ/監査ログ

#### 背景画像

DC レイアウト図をトポロジの背景に配置する。

**実装案**:
- `.clabnoc.yml` で背景画像パスを指定
- トポロジビューの背景レイヤーに表示
- ズームに追従

---

## 4. 追加機能の優先度評価

### 評価基準
- **実用性**: Containerlab 運用でどれだけ頻繁に使うか
- **差別化**: GNS3 や他ツールとの差別化になるか
- **実装コスト**: 実装の複雑さと工数
- **依存関係**: 他の機能の前提となるか

### 優先度 高

| # | 機能 | 実用性 | 差別化 | コスト | 理由 |
|---|------|--------|--------|--------|------|
| 1 | パケットキャプチャ | ★★★ | ★★★ | 中 | ネットワークラボで最頻出の操作。ブラウザ完結ならGNS3超え |
| 2 | インターフェースラベル | ★★★ | ★☆☆ | 小 | トポロジ確認時に常に必要。実装コスト低く効果大 |
| 3 | ノード検索・フィルタ | ★★★ | ★☆☆ | 小 | 大規模トポロジで必須。実装コスト低い |
| 4 | コンフィグ閲覧・エクスポート | ★★★ | ★★☆ | 中 | ラボ運用の基本操作。Kind別コマンド定義が必要 |
| 5 | Console Connect to All | ★★☆ | ★☆☆ | 小 | 頻出ワークフロー。既存のターミナル基盤に追加するだけ |

### 優先度 中

| # | 機能 | 実用性 | 差別化 | コスト | 理由 |
|---|------|--------|--------|--------|------|
| 6 | リソースモニタリング | ★★☆ | ★★★ | 中 | GNS3にない独自機能。Docker Stats APIで実装可能 |
| 7 | ノード一覧テーブルビュー | ★★☆ | ★☆☆ | 小 | 大規模トポロジでの管理性向上 |
| 8 | スナップショット | ★★☆ | ★★☆ | 大 | ラボの再現性確保に有用だが実装が複雑 |
| 9 | アノテーション | ★☆☆ | ★☆☆ | 大 | あると便利だが優先度は低い |
| 10 | BPF フィルタ | ★★☆ | ★★☆ | 中 | 精密な障害注入。netemの延長で実装可能 |

### 優先度 低

| # | 機能 | 実用性 | 差別化 | コスト | 理由 |
|---|------|--------|--------|--------|------|
| 11 | マルチユーザー/RBAC | ★☆☆ | ★☆☆ | 大 | チーム運用時のみ必要 |
| 12 | 背景画像 | ★☆☆ | ★☆☆ | 小 | ニッチな用途 |
| 13 | プロジェクトエクスポート | ★☆☆ | ★☆☆ | 中 | clab自体のエクスポートで代替可能 |

---

## 5. 推奨実装ロードマップ

### Phase 4: Observability & Search

1. インターフェースラベル表示（トグル）
2. ノード検索・フィルタ
3. ノード一覧テーブルビュー
4. コンテナリソースモニタリング（CPU/メモリ）

### Phase 5: Packet Analysis

1. パケットキャプチャ（tcpdump → pcap ダウンロード）
2. リアルタイムパケットストリーム（WebSocket）
3. BPF パケットフィルタ

### Phase 6: Configuration Management

1. コンフィグ閲覧（Kind別コマンド定義）
2. コンフィグエクスポート（一括 zip）
3. Console Connect to All
4. スナップショット（保存・復元・差分）

### Phase 7: Collaboration (将来)

1. 認証・認可
2. マルチユーザー対応
3. 操作ログ/監査ログ
4. アノテーション
5. 背景画像

---

## 参考: GNS3 情報ソース

- [GNS3 Documentation](https://docs.gns3.com/docs/)
- [GNS3 GUI Documentation](https://docs.gns3.com/docs/using-gns3/beginners/the-gns3-gui/)
- [GNS3 3.0.0 New Features](https://www.packettracernetwork.com/features/gns3-300-newfeatures.html)
- [GNS3 API Endpoints](https://gns3-server.readthedocs.io/en/stable/endpoints.html)
- [GNS3 Link Control](https://docs.gns3.com/docs/using-gns3/beginners/link-control/)
- [GNS3 Custom Symbols](https://docs.gns3.com/docs/using-gns3/beginners/change-node-symbol/)
- [GNS3 Docker Support](https://docs.gns3.com/docs/emulators/docker-support-in-gns3/)
- [GNS3 Multi-Client Setup](https://docs.gns3.com/docs/getting-started/installation/one-server-multiple-clients/)
- [GNS3 Web Client Pack](https://github.com/GNS3/gns3-webclient-pack)
- [GNS3 Marketplace](https://gns3.com/marketplace/appliances)
- [GNS3 Configuration Export/Import](https://www.n-study.com/en/how-to-use-gns3/config-export-import/)
- [GNS3 Registry (GitHub)](https://github.com/GNS3/gns3-registry)
- [GNS3 Web UI (GitHub)](https://github.com/GNS3/gns3-web-ui)
