<div align="center">

# Olares: ローカルAIのためのオープンソース主権クラウドOS<!-- omit in toc -->

[![Mission](https://img.shields.io/badge/Mission-Let%20people%20own%20their%20data%20again-purple)](#)<br/>
[![Last Commit](https://img.shields.io/github/last-commit/beclab/olares)](https://github.com/beclab/olares/commits/main)
![Build Status](https://github.com/beclab/olares/actions/workflows/release-daily.yaml/badge.svg)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/beclab/olares)](https://github.com/beclab/olares/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/beclab/olares?style=social)](https://github.com/beclab/olares/stargazers)
[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.com/invite/BzfqrgQPDK)
[![License](https://img.shields.io/badge/License-Olares-darkblue)](https://github.com/beclab/olares/blob/main/LICENSE.md)

<p>
  <a href="./README.md"><img alt="Readme in English" src="https://img.shields.io/badge/English-FFFFFF"></a>
  <a href="./README_CN.md"><img alt="Readme in Chinese" src="https://img.shields.io/badge/简体中文-FFFFFF"></a>
  <a href="./README_JP.md"><img alt="Readme in Japanese" src="https://img.shields.io/badge/日本語-FFFFFF"></a>
</p>

</div>

https://github.com/user-attachments/assets/3089a524-c135-4f96-ad2b-c66bf4ee7471

*Olaresを使って、ローカルAIアシスタントを構築し、データを場所を問わず同期し、ワークスペースをセルフホストし、独自のメディアをストリーミングし、その他多くのことを実現できます。*

<p align="center">
  <a href="https://olares.xyz">ウェブサイト</a> ·
  <a href="https://docs.olares.xyz">ドキュメント</a> ·
  <a href="https://olares.xyz/larepass">LarePassをダウンロード</a> ·
  <a href="https://github.com/beclab/apps">Olaresアプリ</a> ·
  <a href="https://space.olares.xyz">Olares Space</a>
</p>

> [!IMPORTANT]  
> 最近、TerminusからOlaresへのリブランディングを完了しました。詳細については、[リブランディングブログ](https://blog.olares.xyz/terminus-is-now-olares/)をご覧ください。

Olaresを使用して、ハードウェアをAIホームサーバーに変換します。Olaresは、ローカルAIのためのオープンソース主権クラウドOSです。

- **最先端のAIモデルを自分の条件で実行**: LLaMA、Stable Diffusion、Whisper、Flux.1などの強力なオープンAIモデルをハードウェア上で簡単にホストし、AI環境を完全に制御します。
- **簡単にデプロイ**: Olares Marketから幅広いオープンソースAIアプリを数クリックで発見してインストールします。複雑な設定やセットアップは不要です。
- **いつでもどこでもアクセス**: ブラウザを通じて、必要なときにAIアプリやモデルにアクセスします。
- **統合されたAIでスマートなAI体験**: [Model Context Protocol](https://spec.modelcontextprotocol.io/specification/)（MCP）に似たメカニズムを使用して、OlaresはAIモデルとAIアプリ、およびプライベートデータセットをシームレスに接続します。これにより、ニーズに応じて適応する高度にパーソナライズされたコンテキスト対応のAIインタラクションが実現します。

> 🌟 *新しいリリースや更新についての通知を受け取るために、スターを付けてください。*

## なぜOlaresなのか？

以下の理由とシナリオで、Olaresはプライベートで強力かつ安全な主権クラウド体験を提供します：

🤖 **エッジAI**: 最先端のオープンAIモデルをローカルで実行し、大規模言語モデル、コンピュータビジョン、音声認識などを含みます。データに合わせてプライベートAIサービスを作成し、機能性とプライバシーを向上させます。<br>

📊 **個人データリポジトリ**: 重要なファイル、写真、ドキュメントを安全に保存し、デバイスや場所を問わず同期および管理します。<br>

🚀 **セルフホストワークスペース**: 安全なオープンソースSaaS代替品を使用して、チームのための無料のコラボレーションワークスペースを構築します。<br>

🎥 **プライベートメディアサーバー**: 個人のメディアコレクションをホストし、独自のストリーミングサービスを提供します。<br>

🏡 **スマートホームハブ**: IoTデバイスやホームオートメーションの中央制御ポイントを作成します。<br>

🤝 **ユーザー所有の分散型ソーシャルメディア**: Mastodon、Ghost、WordPressなどの分散型ソーシャルメディアアプリをOlaresに簡単にインストールし、プラットフォームの手数料やアカウント停止のリスクなしに個人ブランドを構築します。<br>

📚 **学習プラットフォーム**: セルフホスティング、コンテナオーケストレーション、クラウド技術を実践的に学びます。

## はじめに

### システム互換性
OlaresはLinux、Raspberry Pi、Mac、Windowsで利用可能です。以下のシステムでテストおよび検証されています：

| プラットフォーム            | オペレーティングシステム                     | 備考                                                 |
|---------------------|--------------------------------------|-------------------------------------------------------|
| Linux               | Ubuntu 20.04 LTS以降 <br/> Debian 11以降 |                                          |
| Raspberry Pi        | RaspbianOS                           | Raspberry Pi 4 Model BおよびRaspberry Pi 5で検証済み |
| Windows             | Windows 11 23H2以降 <br/>Windows 10 22H2以降<br/> WSL2 |                                     |
| Mac                 | Monterey (12)以降              |                                                        |
| Proxmox VE (PVE)    | Proxmox Virtual Environment 8.0      |                                                       |

> **注意**
> 
> 互換性テーブルに記載されていないオペレーティングシステムでOlaresを正常にインストールした場合は、お知らせください！GitHubリポジトリで[問題を開く](https://github.com/beclab/Olares/issues/new)か、プルリクエストを送信できます。

### Olaresのセットアップ
自分のデバイスでOlaresを始めるには、[はじめにガイド](https://docs.olares.xyz/manual/get-started/)に従ってステップバイステップの手順を確認してください。

## 技術スタック

パブリッククラウドにはIaaS、PaaS、SaaSのレイヤーがあります。Olaresはこれらのレイヤーに対するオープンソースの代替手段を提供します。

  ![技術スタック](https://file.bttcdn.com/github/terminus/v2/tech-stack-olares.jpeg)

## 機能

Olaresは、セキュリティ、使いやすさ、開発の柔軟性を向上させるための幅広い機能を提供します：

- **エンタープライズグレードのセキュリティ**: Tailscale、Headscale、Cloudflare Tunnel、FRPを使用してネットワーク構成を簡素化します。
- **安全で許可のないアプリケーションエコシステム**: サンドボックス化によりアプリケーションの分離とセキュリティを確保します。
- **統一ファイルシステムとデータベース**: 自動スケーリング、バックアップ、高可用性を提供します。
- **シングルサインオン**: 一度ログインするだけで、Olares内のすべてのアプリケーションに共有認証サービスを使用してアクセスできます。
- **AI機能**: GPU管理、ローカルAIモデルホスティング、プライベートナレッジベースの包括的なソリューションを提供し、データプライバシーを維持します。
- **内蔵アプリケーション**: ファイルマネージャー、同期ドライブ、ボールト、リーダー、アプリマーケット、設定、ダッシュボードを含みます。
- **どこからでもシームレスにアクセス**: モバイル、デスクトップ、ブラウザ用の専用クライアントを使用して、どこからでもデバイスにアクセスできます。
- **開発ツール**: アプリケーションの開発と移植を容易にする包括的な開発ツールを提供します。

## 他のセルフホスティングソリューションとの比較

ローカルAIのためのオープンソース主権クラウドOSとして、Olaresはセルフホスティングの可能性を再考します。Olaresが市場の他のセルフホスティングソリューションとどのように異なるかを理解するために、Olaresの機能を市場の他のセルフホスティングソリューションと比較した表を作成しました。

**凡例：** 

- 🚀: **自動**、システムがタスクを自動的に完了することを示します。
- ✅: **はい**、開発者の背景がないユーザーが製品のUIプロンプトを通じて設定を完了できることを示します。
- 🛠️: **手動設定**、エンジニアリングの背景を持つユーザーでもチュートリアルを参照して設定を完了する必要があることを示します。
- ❌:  **いいえ**、機能がサポートされていないことを示します。

| | Olares | Synology | TrueNAS | CasaOS | Unraid |
| --- | --- | --- | --- | --- | --- |
| ソースコードライセンス | Olaresライセンス | クローズド | GPL 3.0 | Apache 2.0 | クローズド |
| 基盤 | Kubernetes | Linux | Kubernetes | Docker | Docker |
| ローカルLLMホスティング | 🚀  | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| ローカルLLMアプリ開発 | 🚀 | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| マルチノード | ✅   | ❌   | ✅   | ❌   | ❌   |
| 内蔵アプリ | ✅ (豊富なデスクトップアプリ) | ✅ (豊富なデスクトップアプリ) | ❌ (CLI) | ✅ (シンプルなデスクトップアプリ) | ✅ (ダッシュボード) |
| 無料ドメイン名 | ✅   | ✅   | ❌   | ❌   | ❌   |
| 自動SSL証明書 | 🚀  | ✅   | 🛠️ | 🛠️ | 🛠️ |
| リバースプロキシ | 🚀  | ✅   | 🛠️ | 🛠️ | 🛠️ |
| VPN管理 | 🚀  | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| グレード付きアプリエントランス | 🚀  | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| マルチユーザー管理 | ✅ ユーザー管理 <br>🚀 リソース分離 | ✅ ユーザー管理<br>🛠️ リソース分離 | ✅ ユーザー管理<br>🛠️ リソース分離 | ❌   | ✅ ユーザー管理  <br>🛠️ リソース分離 |
| すべてのアプリのシングルログイン | 🚀  | ❌   | ❌   | ❌   |  ❌   |
| クロスノードストレージ | 🚀 (Juicefs+<br>MinIO) | ❌   | ❌   | ❌   | ❌   |
| データベースソリューション | 🚀 (内蔵クラウドネイティブソリューション) | 🛠️ | 🛠️ | 🛠️ | 🛠️ |
| 災害復旧 | 🚀 (MinIOの[**エラーレート**](https://min.io/docs/minio/linux/operations/concepts/erasure-coding.html)**)** | ✅ RAID | ✅ RAID | ✅ RAID | ✅ Unraid Storage |
| バックアップ | ✅ アプリデータ  <br>✅ ユーザーデータ | ✅ ユーザーデータ | ✅ ユーザーデータ | ✅ ユーザーデータ | ✅ ユーザーデータ |
| アプリサンドボックス | ✅   | ❌   | ❌ (K8Sの名前空間) | ❌   | ❌   |
| アプリエコシステム | ✅ (公式 + サードパーティ) | ✅ (主に公式アプリ) | ✅ (公式 + サードパーティの提出) | ✅ 主に公式アプリ | ✅ (コミュニティアプリマーケット) |
| 開発者フレンドリー | ✅ IDE  <br>✅ CLI  <br>✅ SDK  <br>✅ ドキュメント | ✅ CLI  <br>✅ SDK  <br>✅ ドキュメント | ✅ CLI  <br>✅ ドキュメント | ✅ CLI  <br>✅ ドキュメント | ✅ ドキュメント |
| クライアントプラットフォーム | ✅ Android  <br>✅ iOS  <br>✅ Windows  <br>✅ Mac  <br>✅ Chromeプラグイン | ✅ Android  <br>✅ iOS | ❌   | ❌   | ❌   |
| クライアント機能 | ✅ (オールインワンクライアントアプリ) | ✅ (14個の別々のクライアントアプリ) | ❌   | ❌   |  ❌   |

## プロジェクトナビゲーション

Olaresは、GitHubで公開されている多数のコードリポジトリで構成されています。現在のリポジトリは、オペレーティングシステムの最終コンパイル、パッケージング、インストール、およびアップグレードを担当しており、特定の変更は主に対応するリポジトリで行われます。

以下の表は、Olaresのプロジェクトディレクトリと対応するリポジトリを一覧にしたものです。興味のあるものを見つけてください：

<details>
<summary><b>フレームワークコンポーネント</b></summary>
  
| ディレクトリ | リポジトリ | 説明 |
| --- | --- | --- |
| [frameworks/app-service](https://github.com/beclab/olares/tree/main/frameworks/app-service) | <https://github.com/beclab/app-service> | システムフレームワークコンポーネントで、システム内のすべてのアプリのライフサイクル管理とさまざまなセキュリティ制御を提供します。 |
| [frameworks/backup-server](https://github.com/beclab/olares/tree/main/frameworks/backup-server) | <https://github.com/beclab/backup-server> | システムフレームワークコンポーネントで、定期的なフルまたは増分クラスターのバックアップサービスを提供します。 |
| [frameworks/bfl](https://github.com/beclab/olares/tree/main/frameworks/bfl) | <https://github.com/beclab/bfl> | ランチャーのバックエンド（BFL）、ユーザーアクセスポイントとして機能し、さまざまなバックエンドサービスのインターフェースを集約およびプロキシします。 |
| [frameworks/GPU](https://github.com/beclab/olares/tree/main/frameworks/GPU) | <https://github.com/grgalex/nvshare> | 複数のプロセス（またはKubernetes上で実行されるコンテナ）が同じ物理GPU上で同時に安全に実行できるようにするGPU共有メカニズムで、各プロセスが全GPUメモリを利用できます。 |
| [frameworks/l4-bfl-proxy](https://github.com/beclab/olares/tree/main/frameworks/l4-bfl-proxy) | <https://github.com/beclab/l4-bfl-proxy> | BFLの第4層ネットワークプロキシ。SNIを事前に読み取ることで、ユーザーのIngressに通過する動的ルートを提供します。 |
| [frameworks/osnode-init](https://github.com/beclab/olares/tree/main/frameworks/osnode-init) | <https://github.com/beclab/osnode-init> | 新しいノードがクラスターに参加する際にノードデータを初期化するシステムフレームワークコンポーネント。 |
| [frameworks/system-server](https://github.com/beclab/olares/tree/main/frameworks/system-server) | <https://github.com/beclab/system-server> | システムランタイムフレームワークの一部として、アプリ間のセキュリティコールのメカニズムを提供します。 |
| [frameworks/tapr](https://github.com/beclab/olares/tree/main/frameworks/tapr) | <https://github.com/beclab/tapr> | Olaresアプリケーションランタイムコンポーネント。 |
</details>

<details>
<summary><b>システムレベルのアプリケーションとサービス</b></summary>
  
| ディレクトリ | リポジトリ | 説明 |
| --- | --- | --- |
| [apps/analytic](https://github.com/beclab/olares/tree/main/apps/analytic) | <https://github.com/beclab/analytic> | [Umami](https://github.com/umami-software/umami)に基づいて開発されたAnalyticは、Google Analyticsのシンプルで高速、プライバシー重視の代替品です。 |
| [apps/market](https://github.com/beclab/olares/tree/main/apps/market) | <https://github.com/beclab/market> | このリポジトリは、Olaresのアプリケーションマーケットのフロントエンド部分をデプロイします。 |
| [apps/market-server](https://github.com/beclab/olares/tree/main/apps/market-server) | <https://github.com/beclab/market> | このリポジトリは、Olaresのアプリケーションマーケットのバックエンド部分をデプロイします。 |
| [apps/argo](https://github.com/beclab/olares/tree/main/apps/argo) | <https://github.com/argoproj/argo-workflows> | ローカル推奨アルゴリズムのコンテナ実行をオーケストレーションするワークフローエンジン。 |
| [apps/desktop](https://github.com/beclab/olares/tree/main/apps/desktop) | <https://github.com/beclab/desktop> | システムの内蔵デスクトップアプリケーション。 |
| [apps/devbox](https://github.com/beclab/olares/tree/main/apps/devbox) | <https://github.com/beclab/devbox> | Olaresアプリケーションの移植と開発のための開発者向けIDE。 |
| [apps/vault](https://github.com/beclab/olares/tree/main/apps/vault) | <https://github.com/beclab/termipass> | [Padloc](https://github.com/padloc/padloc)に基づいて開発された、あらゆる規模のチームや企業向けの無料の1PasswordおよびBitwardenの代替品。DID、Olares ID、およびOlaresデバイスの管理を支援するクライアントとして機能します。 |
| [apps/files](https://github.com/beclab/olares/tree/main/apps/files) | <https://github.com/beclab/files> | [Filebrowser](https://github.com/filebrowser/filebrowser)から変更された内蔵ファイルマネージャーで、Drive、Sync、およびさまざまなOlares物理ノード上のファイルの管理を提供します。 |
| [apps/notifications](https://github.com/beclab/olares/tree/main/apps/notifications) | <https://github.com/beclab/notifications> | Olaresの通知システム |
| [apps/profile](https://github.com/beclab/olares/tree/main/apps/profile) | <https://github.com/beclab/profile> | OlaresのLinktree代替品 |
| [apps/rsshub](https://github.com/beclab/olares/tree/main/apps/rsshub) | <https://github.com/beclab/rsshub> | [RssHub](https://github.com/DIYgod/RSSHub)に基づいたRSS購読管理ツール。 |
| [apps/settings](https://github.com/beclab/olares/tree/main/apps/settings) | <https://github.com/beclab/settings> | 内蔵システム設定。 |
| [apps/system-apps](https://github.com/beclab/olares/tree/main/apps/system-apps) | <https://github.com/beclab/system-apps> | _kubesphere/console_プロジェクトに基づいて構築されたsystem-serviceは、視覚的なダッシュボードと機能豊富なControlHubを通じて、システムの実行状態とリソース使用状況を理解し、制御するためのセルフホストクラウドプラットフォームを提供します。 |
| [apps/wizard](https://github.com/beclab/olares/tree/main/apps/wizard) | <https://github.com/beclab/wizard> | ユーザーにシステムのアクティベーションプロセスを案内するウィザードアプリケーション。 |
</details>

<details>
<summary><b>サードパーティコンポーネントとサービス</b></summary>

| ディレクトリ | リポジトリ | 説明 |
| --- | --- | --- |
| [third-party/authelia](https://github.com/beclab/olares/tree/main/third-party/authelia) | <https://github.com/beclab/authelia> | Webポータルを介してアプリケーションに二要素認証とシングルサインオン（SSO）を提供するオープンソースの認証および認可サーバー。 |
| [third-party/headscale](https://github.com/beclab/olares/tree/main/third-party/headscale) | <https://github.com/beclab/headscale> | OlaresでのTailscaleコントロールサーバーのオープンソース自ホスト実装で、LarePassで異なるデバイス間でTailscaleを管理します。 |
| [third-party/infisical](https://github.com/beclab/olares/tree/main/third-party/infisical) | <https://github.com/beclab/infisical> | チーム/インフラストラクチャ間でシークレットを同期し、シークレットの漏洩を防ぐオープンソースのシーク��ッ��管理プラットフォーム。 |
| [third-party/juicefs](https://github.com/beclab/olares/tree/main/third-party/juicefs) | <https://github.com/beclab/juicefs-ext> | RedisとS3の上に構築された分散POSIXファイルシステムで、異なるノード上のアプリがPOSIXインターフェースを介して同じデータにアクセスできるようにします。 |
| [third-party/ks-console](https://github.com/beclab/olares/tree/main/third-party/ks-console) | <https://github.com/kubesphere/console> | Web GUIを介してクラスター管理を可能にするKubesphereコンソール。 |
| [third-party/ks-installer](https://github.com/beclab/olares/tree/main/third-party/ks-installer) | <https://github.com/beclab/ks-installer-ext> | クラスターリソース定義に基づいて自動的にKubesphereクラスターを作成するKubesphereインストーラーコンポーネント。 |
| [third-party/kube-state-metrics](https://github.com/beclab/olares/tree/main/third-party/kube-state-metrics) | <https://github.com/beclab/kube-state-metrics> | kube-state-metrics（KSM）は、Kubernetes APIサーバーをリッスンし、オブジェクトの状態に関するメトリックを生成するシンプルなサービスです。 |
| [third-party/notification-manager](https://github.com/beclab/olares/tree/main/third-party/notification-manager) | <https://github.com/beclab/notification-manager-ext> | 複数の通知チャネルの統一管理と通知内容のカスタム集約を提供するKubesphereの通知管��コンポーネント。 |
| [third-party/predixy](https://github.com/beclab/olares/tree/main/third-party/predixy) | <https://github.com/beclab/predixy> | 利用可能なノードを自動的に識別し、名前空間の分離を追加するRedisクラスターのプロキシサービス。 |
| [third-party/redis-cluster-operator](https://github.com/beclab/olares/tree/main/third-party/redis-cluster-operator) | <https://github.com/beclab/redis-cluster-operator> | Kubernetesに基づいてRedisクラスターを作成および管理するためのクラウドネイティブツール。 |
| [third-party/seafile-server](https://github.com/beclab/olares/tree/main/third-party/seafile-server) | <https://github.com/beclab/seafile-server> | データストレージを処理するSeafile（同期ドライブ）のバックエンドサービス。 |
| [third-party/seahub](https://github.com/beclab/olares/tree/main/third-party/seahub) | <https://github.com/beclab/seahub> | ファイル共有、データ同期などを処理するSeafile（同期ドライブ）のフロントエンドおよびミドルウェアサービス。 |
| [third-party/tailscale](https://github.com/beclab/olares/tree/main/third-party/tailscale) | <https://github.com/tailscale/tailscale> | TailscaleはすべてのプラットフォームのLarePassに統合されています。 |
</details>

<details>
<summary><b>追加のライブラリとコンポーネント</b></summary>

| ディレクトリ | リポジトリ | 説明 |
| --- | --- | --- |
| [build/installer](https://github.com/beclab/olares/tree/main/build/installer) |     | インストーラービルドを生成するためのテンプレート。 |
| [build/manifest](https://github.com/beclab/olares/tree/main/build/manifest) |     | インストールビルドイメージリストテンプレート。 |
| [libs/fs-lib](https://github.com/beclab/olares/tree/main/libs) | <https://github.com/beclab/fs-lib> | JuiceFSに基づいて実装されたiNotify互換インターフェースのSDKライブラリ。 |
| [scripts](https://github.com/beclab/olares/tree/main/scripts) |     | インストーラービルドを生成するための補助スクリプト。 |
</details>

## Olaresへの貢献

あらゆる形での貢献を歓迎します：

- Olaresで独自のアプリケーションを開発したい場合は、以下を参照してください：<br>
https://docs.olares.xyz/developer/develop/


- Olaresの改善に協力したい場合は、以下を参照してください：<br>
https://docs.olares.xyz/developer/contribute/olares.html

## コミュニティと連絡先

* [**GitHub Discussion**](https://github.com/beclab/olares/discussions). フィードバックの共有や質問に最適です。
* [**GitHub Issues**](https://github.com/beclab/olares/issues). Olaresの使用中に遭遇したバグの報告や機能提案の提出に最適です。 
* [**Discord**](https://discord.com/invite/BzfqrgQPDK). Olaresに関するあらゆることを共有するのに最適です。

## 特別な感謝

Olaresプロジェクトは、次のような多数のサードパーティオープンソースプロジェクトを統合しています：[Kubernetes](https://kubernetes.io/)、[Kubesphere](https://github.com/kubesphere/kubesphere)、[Padloc](https://padloc.app/)、[K3S](https://k3s.io/)、[JuiceFS](https://github.com/juicedata/juicefs)、[MinIO](https://github.com/minio/minio)、[Envoy](https://github.com/envoyproxy/envoy)、[Authelia](https://github.com/authelia/authelia)、[Infisical](https://github.com/Infisical/infisical)、[Dify](https://github.com/langgenius/dify)、[Seafile](https://github.com/haiwen/seafile)、[HeadScale](https://headscale.net/)、 [tailscale](https://tailscale.com/)、[Redis Operator](https://github.com/spotahome/redis-operator)、[Nitro](https://nitro.jan.ai/)、[RssHub](http://rsshub.app/)、[predixy](https://github.com/joyieldInc/predixy)、[nvshare](https://github.com/grgalex/nvshare)、[LangChain](https://www.langchain.com/)、[Quasar](https://quasar.dev/)、[TrustWallet](https://trustwallet.com/)、[Restic](https://restic.net/)、[ZincSearch](https://zincsearch-docs.zinc.dev/)、[filebrowser](https://filebrowser.org/)、[lego](https://go-acme.github.io/lego/)、[Velero](https://velero.io/)、[s3rver](https://github.com/jamhall/s3rver)、[Citusdata](https://www.citusdata.com/)。
