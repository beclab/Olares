<div align="center">

# Olares: ローカルAIのためのオープンソース主権クラウドOS<!-- omit in toc -->

[![Mission](https://img.shields.io/badge/Mission-Let%20people%20own%20their%20data%20again-purple)](#)<br/>
[![Last Commit](https://img.shields.io/github/last-commit/beclab/olares)](https://github.com/beclab/olares/commits/main)
![Build Status](https://github.com/beclab/olares/actions/workflows/release-daily.yaml/badge.svg)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/beclab/olares)](https://github.com/beclab/olares/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/beclab/olares?style=social)](https://github.com/beclab/olares/stargazers)
[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/olares)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue)](https://github.com/beclab/olares/blob/main/LICENSE)

<p>
  <a href="./README.md"><img alt="Readme in English" src="https://img.shields.io/badge/English-FFFFFF"></a>
  <a href="./README_CN.md"><img alt="Readme in Chinese" src="https://img.shields.io/badge/简体中文-FFFFFF"></a>
  <a href="./README_JP.md"><img alt="Readme in Japanese" src="https://img.shields.io/badge/日本語-FFFFFF"></a>
</p>

</div>

<p align="center">
  <a href="https://olares.com">ウェブサイト</a> ·
  <a href="https://docs.olares.com">ドキュメント</a> ·
  <a href="https://larepass.olares.com">LarePassをダウンロード</a> ·
  <a href="https://github.com/beclab/apps">Olaresアプリ</a> ·
  <a href="https://space.olares.com">Olares Space</a>
</p>

> *パブリッククラウドを基盤とする現代のインターネットは、あなたの個人データのプライバシーをますます脅かしています。ChatGPT、Midjourney、Facebookといったサービスへの依存が深まるにつれ、デジタル主権に対するあなたのコントロールも弱まっています。あなたのデータは他者のサーバーに保存され、その利用規約に縛られ、追跡され、検閲されているのです。*
>
>*今こそ、変革の時です。*

![自身のデジタル](https://app.cdn.olares.com/github/olares/public-cloud-to-personal-cloud.jpg)

私たちは、あなたが自身のデジタルライフをコントロールする基本的な権利を有すると確信しています。この権利を守る最も効果的な方法は、あなたのデータをローカルの、あなた自身のハードウェア上でホストすることです。

Olaresは、あなたが自身のデジタル資産をローカルで容易に所有し管理できるよう設計された、オープンソースのパーソナルクラウドOSです。もはやパブリッククラウドサービスに依存する必要はありません。Olares上で、例えばOllamaを利用した大規模言語モデルのホスティング、SD WebUIによる画像生成、Mastodonを用いた検閲のないソーシャルスペースの構築など、強力なオープンソースの代替サービスやアプリケーションをローカルにデプロイできます。Olaresは、クラウドコンピューティングの絶大な力を活用しつつ、それを完全に自身のコントロール下に置くことを可能にします。

> 🌟 *新しいリリースや更新についての通知を受け取るために、スターを付けてください。*

## アーキテクチャ

パブリッククラウドは、IaaS (Infrastructure as a Service)、PaaS (Platform as a Service)、SaaS (Software as a Service) といったサービスレイヤーで構成されています。Olaresは、これら各レイヤーに対するオープンソースの代替ソリューションを提供しています。

  ![Olaresのアーキテクチ](https://app.cdn.olares.com/github/olares/olares-architecture.jpg)

各コンポーネントの詳細については、[Olares アーキテクチャ](https://docs.olares.com/manual/concepts/system-architecture.html)（英語版）をご参照ください。

> 🔍**OlaresとNASの違いは何ですか？**
>
> Olaresは、ワンストップのセルフホスティング・パーソナルクラウド体験の実現を目指しています。そのコア機能とユーザーの位置付けは、ネットワークストレージに特化した従来のNASとは大きく異なります。詳細は、[OlaresとNASの比較](https://docs.olares.com/manual/olares-vs-nas.html)（英語版）をご参照ください。

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

以下はUIのスクリーンショットプレビューです。

| **デスクトップ：馴染みやすく効率的なアクセスポイント** |  **ファイルマネージャー：データを安全に保管** |
| :--------: | :-------: |
| ![桌面](https://app.cdn.olares.com/github/terminus/v2/desktop.jpg) | ![文件](https://app.cdn.olares.com/github/terminus/v2/files.jpg) |
| **Vault：安心のパスワード管理**|**マーケット：コントロール可能なアプリエコシステム** |
| ![vault](https://app.cdn.olares.com/github/terminus/v2/vault.jpg) | ![市场](https://app.cdn.olares.com/github/terminus/v2/market.jpg) |
| **Wise：あなただけのデジタルガーデン** | **設定：Olaresを効率的に管理** |
| ![设置](https://app.cdn.olares.com/github/terminus/v2/wise.jpg) | ![](https://app.cdn.olares.com/github/terminus/v2/settings.jpg) |
| **ダッシュボード：Olaresを継続的に監視** | **プロフィール：ユニークなパーソナルページ** |
| ![面板](https://app.cdn.olares.com/github/terminus/v2/dashboard.jpg) | ![profile](https://app.cdn.olares.com/github/terminus/v2/profile.jpg) |
| **Studio：開発、デバッグ、デプロイをワンストップで**|**コントロールパネル：Kubernetesクラスターを簡単に管理** |
| ![Devbox](https://app.cdn.olares.com/github/terminus/v2/devbox.jpg) | ![控制中心](https://app.cdn.olares.com/github/terminus/v2/controlhub.jpg)|

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


## プロジェクトナビゲーションx
このセクションでは、Olares リポジトリ内の主要なディレクトリをリストアップしています：

* **[`apps`](./apps)**: システムアプリケーションのコードが含まれており、主に `larepass` 用です。
* **[`cli`](./cli)**: Olares のコマンドラインインターフェースツールである `olares-cli` のコードが含まれています。
* **[`daemon`](./daemon)**: システムデーモンプロセスである `olaresd` のコードが含まれています。
* **[`docs`](./docs)**: プロジェクトのドキュメントが含まれています。
* **[`framework`](./framework)**: Olares システムサービスが含まれています。
* **[`infrastructure`](./infrastructure)**: コンピューティング、ストレージ、ネットワーキング、GPU などのインフラストラクチャコンポーネントに関連するコードが含まれています。
* **[`platform`](./platform)**: データベースやメッセージキューなどのクラウドネイティブコンポーネントのコードが含まれています。
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
* [**Discord**](https://discord.gg/olares). Olaresに関するあらゆることを共有するのに最適です。

## 特別な感謝

Olaresプロジェクトは、次のような多数のサードパーティオープンソースプロジェクトを統合しています：[Kubernetes](https://kubernetes.io/)、[Kubesphere](https://github.com/kubesphere/kubesphere)、[Padloc](https://padloc.app/)、[K3S](https://k3s.io/)、[JuiceFS](https://github.com/juicedata/juicefs)、[MinIO](https://github.com/minio/minio)、[Envoy](https://github.com/envoyproxy/envoy)、[Authelia](https://github.com/authelia/authelia)、[Infisical](https://github.com/Infisical/infisical)、[Dify](https://github.com/langgenius/dify)、[Seafile](https://github.com/haiwen/seafile)、[HeadScale](https://headscale.net/)、 [tailscale](https://tailscale.com/)、[Redis Operator](https://github.com/spotahome/redis-operator)、[Nitro](https://nitro.jan.ai/)、[RssHub](http://rsshub.app/)、[predixy](https://github.com/joyieldInc/predixy)、[nvshare](https://github.com/grgalex/nvshare)、[LangChain](https://www.langchain.com/)、[Quasar](https://quasar.dev/)、[TrustWallet](https://trustwallet.com/)、[Restic](https://restic.net/)、[ZincSearch](https://zincsearch-docs.zinc.dev/)、[filebrowser](https://filebrowser.org/)、[lego](https://go-acme.github.io/lego/)、[Velero](https://velero.io/)、[s3rver](https://github.com/jamhall/s3rver)、[Citusdata](https://www.citusdata.com/)。
