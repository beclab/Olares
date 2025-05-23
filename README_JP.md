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
  <a href="https://olares.com">ウェブサイト</a> ·
  <a href="https://docs.olares.com">ドキュメント</a> ·
  <a href="https://olares.com/larepass">LarePassをダウンロード</a> ·
  <a href="https://github.com/beclab/apps">Olaresアプリ</a> ·
  <a href="https://space.olares.com">Olares Space</a>
</p>

> [!IMPORTANT]  
> 最近、TerminusからOlaresへのリブランディングを完了しました。詳細については、[リブランディングブログ](https://blog.olares.com/terminus-is-now-olares/)をご覧ください。

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

Olaresは以下のLinuxプラットフォームで動作検証を完了しています：

- Ubuntu 24.04 LTS 以降
- Debian 11 以降

### Olaresのセットアップ
自分のデバイスでOlaresを始めるには、[はじめにガイド](https://docs.olares.com/manual/get-started/)に従ってステップバイステップの手順を確認してください。

## アーキテクチャ

Olaresのアーキテクチャは、次の2つの基本原則に基づいています：
- Androidの設計思想を取り入れ、ソフトウェアの権限と対話性を制御することで、システムの安全かつ円滑な運用を実現します。
- クラウドネイティブ技術を活用し、ハードウェアとミドルウェアサービスを効率的に管理します。

  ![Olaresのアーキテクチ](https://file.bttcdn.com/github/terminus/v2/olares-arch-3.png)

各コンポーネントの詳細については、[Olares アーキテクチャ](https://docs.olares.com/manual/system-architecture.html)（英語版）をご参照ください。

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

## プロジェクトナビゲーション

このセクションでは、Olares リポジトリ内の主要なディレクトリをリストアップしています：

* **`apps`**: システムアプリケーションのコードが含まれており、主に `larepass` 用です。
* **`cli`**: Olares のコマンドラインインターフェースツールである `olares-cli` のコードが含まれています。
* **`daemon`**: システムデーモンプロセスである `olaresd` のコードが含まれています。
* **`docs`**: プロジェクトのドキュメントが含まれています。
* **`framework`**: Olares システムサービスが含まれています。
* **`infrastructure`**: コンピューティング、ストレージ、ネットワーキング、GPU などのインフラストラクチャコンポーネントに関連するコードが含まれています。
* **`platform`**: データベースやメッセージキューなどのクラウドネイティブコンポーネントのコードが含まれています。
* **`vendor`**: サードパーティのハードウェアベンダーからのコードが含まれています。

## Olaresへの貢献

あらゆる形での貢献を歓迎します：

- Olaresで独自のアプリケーションを開発したい場合は、以下を参照してください：<br>
https://docs.olares.com/developer/develop/


- Olaresの改善に協力したい場合は、以下を参照してください：<br>
https://docs.olares.com/developer/contribute/olares.html

## コミュニティと連絡先

* [**GitHub Discussion**](https://github.com/beclab/olares/discussions). フィードバックの共有や質問に最適です。
* [**GitHub Issues**](https://github.com/beclab/olares/issues). Olaresの使用中に遭遇したバグの報告や機能提案の提出に最適です。 
* [**Discord**](https://discord.com/invite/BzfqrgQPDK). Olaresに関するあらゆることを共有するのに最適です。

## 特別な感謝

Olaresプロジェクトは、次のような多数のサードパーティオープンソースプロジェクトを統合しています：[Kubernetes](https://kubernetes.io/)、[Kubesphere](https://github.com/kubesphere/kubesphere)、[Padloc](https://padloc.app/)、[K3S](https://k3s.io/)、[JuiceFS](https://github.com/juicedata/juicefs)、[MinIO](https://github.com/minio/minio)、[Envoy](https://github.com/envoyproxy/envoy)、[Authelia](https://github.com/authelia/authelia)、[Infisical](https://github.com/Infisical/infisical)、[Dify](https://github.com/langgenius/dify)、[Seafile](https://github.com/haiwen/seafile)、[HeadScale](https://headscale.net/)、 [tailscale](https://tailscale.com/)、[Redis Operator](https://github.com/spotahome/redis-operator)、[Nitro](https://nitro.jan.ai/)、[RssHub](http://rsshub.app/)、[predixy](https://github.com/joyieldInc/predixy)、[nvshare](https://github.com/grgalex/nvshare)、[LangChain](https://www.langchain.com/)、[Quasar](https://quasar.dev/)、[TrustWallet](https://trustwallet.com/)、[Restic](https://restic.net/)、[ZincSearch](https://zincsearch-docs.zinc.dev/)、[filebrowser](https://filebrowser.org/)、[lego](https://go-acme.github.io/lego/)、[Velero](https://velero.io/)、[s3rver](https://github.com/jamhall/s3rver)、[Citusdata](https://www.citusdata.com/)。
