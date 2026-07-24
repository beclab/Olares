<h1 align="center">Olares</h1>

<p align="center"><strong>あなたのAIエージェント。あなたのデータ。あなたのハードウェア。</strong></p>

<p align="center">
  <a href="#はじめに">Olaresをインストール</a> ·
  <a href="https://www.olares.com/docs/developer/cli-agent-skills">AIでOlaresを管理</a> ·
  <a href="#貢献">貢献する</a>
</p>

<div align="center">

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/beclab/Olares)](https://github.com/beclab/olares/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/beclab/Olares?style=social)](https://github.com/beclab/Olares/stargazers)
[![Discord](https://img.shields.io/badge/Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/olares)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue)](https://github.com/beclab/olares/blob/main/LICENSE)

<a href="https://trendshift.io/repositories/15376" target="_blank"><img src="https://trendshift.io/api/badge/repositories/15376" alt="beclab%2FOlares | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>

<p>
  <a href="./README.md"><img alt="Readme in English" src="https://img.shields.io/badge/English-FFFFFF"></a>
  <a href="./README_CN.md"><img alt="Readme in Chinese" src="https://img.shields.io/badge/简体中文-FFFFFF"></a>
  <a href="./README_JP.md"><img alt="Readme in Japanese" src="https://img.shields.io/badge/日本語-FFFFFF"></a>
</p>

</div>


**Olaresは、自然言語で操作できるオープンソースのパーソナルクラウドOSです。自分のハードウェア上でAIエージェントやLLMを動かすために生まれました。**

OlaresはKubernetesを基盤に、手元のマシンを、どのブラウザからでも使えるセルフホスト型のAIプラットフォームへと変えます。個人ユーザーから小さなチームまで、コンピュート・ストレージ・ネットワーク・アプリをひとつの場所でまとめて扱えます。

https://github.com/user-attachments/assets/01490c33-41ce-46fe-8450-6939b40db98e

> 🌟 *Olaresが役に立ったら、ぜひスターを付けてください。応援が改善を続ける励みになります。*

## なぜOlaresなのか

本当に使えるAIは、あなたのことをよく知っている必要があります。そのためには、ファイルやメッセージ、これまでの履歴にアクセスできることが欠かせません。ところが多くのクラウドAIは、こうした機微なデータをサードパーティのサーバーに預けさせ、しかも使った分だけ課金します。

OlaresはAIを手元に取り戻します。[OpenClaw](https://www.olares.com/docs/use-cases/openclaw) のようなエージェントを、自分のハードウェア上のローカルLLMで動かしながら、クラウドならではの手軽さと、どこからでもアクセスできる利便性はそのまま得られます。

![パブリッククラウドサービスで築いたデジタルライフと、Olaresパーソナルクラウド上のオープンソースアプリで動くデジタルライフの比較](https://app.cdn.olares.com/github/olares/public-cloud-to-personal-cloud.jpg)

主な機能：

- **ワンクリックでローカルAI**：[Olares Market](https://www.olares.com/market/) から、オープンソースのAIアプリやモデルをワンクリックで導入できます。
- **アクセラレーテッドコンピューティング管理**：複数ノードのGPUやアクセラレータをまとめて割り当て、タイムスライシング・メモリスライシング・GPU専有モードにより、AI・メディア・ゲームなど用途に応じて使い分けられます。
- **[ファイルとストレージ管理](https://www.olares.com/docs/manual/olares/files/)**：内蔵の「Files」アプリから、ローカルファイル・同期データ・接続済みのクラウドストレージ・外部のSMB/NFS共有をまとめて扱え、[バックアップも自由に設定](https://www.olares.com/docs/manual/olares/settings/backup)できます。
- **[プライベートネットワークとアクセス制御](https://www.olares.com/docs/developer/concepts/network)**：プライベートVPNとリバースプロキシに加え、公開・非公開・内部の3つの入口を用意。ポートを手動で開けなくても、各アプリにHTTPSのアドレスが割り当てられます。
- **いつでも、どこからでも**：Olares IDと [LarePass](https://www.olares.com/docs/manual/larepass/) があれば、スマホ・PC・ブラウザからすべてのサービスにアクセスできます。
- **充実のシステムアプリ**：Files、Vault、Market、Dashboard、Control Hubなどを最初から搭載。ログインすればすぐに使えます。

## はじめに

### Linuxスクリプトの動作要件

OlaresはLinuxホスト（物理マシンでも仮想マシンでも可）にインストールでき、Windows・macOS・Raspberry Pi向けには専用の方法も用意しています。要件はプラットフォームやインストール方法によって異なります。ここで使うLinuxスクリプトの要件は次のとおりです。

- **CPU**：4コア以上
- **メモリ**：8 GB以上の空き
- **ストレージ**：150 GB以上の空きSSD（HDDではインストールに失敗します）
- **OS**：Ubuntu 22.04〜25.04、またはDebian 12 / 13

専用GPUは任意で、あればローカルAIを高速化できます。

### インストールとアクティベーション

1. まず [LarePass](https://www.olares.com/docs/manual/larepass/) でOlares IDを作成します。LarePassは、安全なログイン・内蔵VPN・ファイル同期を備えたコンパニオンアプリです。

2. Linuxホストで次を実行します。

    ```bash
    curl -fsSL https://olares.sh | bash -
    ```

    このコマンドは `olares.sh` から公式インストーラーを取得し、Bashで実行します。詳しい要件やプラットフォーム別の手順、トラブルシューティングは [Linuxスクリプトインストールガイド](https://www.olares.com/docs/manual/get-started/install-linux-script) をご覧ください。

    Windows・macOS・Raspberry Pi・仮想マシンにインストールする場合は、[インストールガイド](https://www.olares.com/docs/manual/get-started/install-olares) で対象のプラットフォームを選んでください。

3. 画面のガイド付きウィザードに従うか、[Olares CLIでのアクティベーション](https://www.olares.com/docs/manual/best-practices/activate-olares-using-cli) を参考に、ターミナルだけで完結させることもできます。

アクティベーションが済んだら、Olares IDに紐づくアドレスから、どのブラウザでもOlaresを開けます。たとえばOlares IDが `marvin123` なら、デスクトップは `https://desktop.marvin123.olares.com` です。

## 主なユースケース

- **パーソナルAIエージェントに任せる**：リサーチ、コーディング、ファイル整理、日々の自動化を、自然言語で任せられます。
- **生成AIをローカルで動かす**：オープンモデルとチャットし、画像や動画を生成し、ローカルモデルを他のアプリにつなぐ。すべて自分のハードウェア上で完結します。
- **スマートホームとメディアを管理する**：ホームオートメーション機器をつなぎ、自分の音楽や動画ライブラリをいつでもストリーミングできます。
- **エージェンティックなアプリを開発・運用する**：隔離された環境で、アプリやワークフローを開発・テスト・実行できます。
- **音声をローカルで処理する**：会議の文字起こし、録音の翻訳、音声合成を、外部にアップロードせずに行えます。
- **自分たちのワークスペースを持つ**：家族やチームに、ドキュメント・自動化・プロジェクト管理・コミュニケーションの道具をまとめて用意できます。
- **個人データを自分で管理する**：デバイスをまたいで、ファイル・写真・ドキュメントを保存、同期、バックアップ、閲覧できます。

## システムアーキテクチャ

パブリッククラウドがIaaS・PaaS・SaaSという層に分かれているように、Olaresもそれぞれの層にオープンソースの代替を用意しています。

  ![オープンソースのコンポーネントをIaaS・PaaS・SaaSの各層に対応づけ、パブリッククラウドの同種サービスと並べて示したOlaresのアーキテクチャ](https://app.cdn.olares.com/github/olares/olares-architecture.jpg)

各コンポーネントの詳細は、[Olaresアーキテクチャ](https://www.olares.com/docs/developer/concepts/system-architecture)（英語）をご覧ください。

> 🔍 **OlaresとNASの違いは何ですか？**
>
> Olaresが目指すのは、オールインワンのセルフホスト型パーソナルクラウド体験です。その中心となる機能や対象ユーザーは、ネットワークストレージが主眼の従来型NASとは大きく異なります。詳しくは [OlaresとNASの比較](https://www.olares.com/blog/compare-olares-and-nas/)（英語）をご覧ください。

## プロジェクト構成

Olaresリポジトリの主なディレクトリ：

```
Olares/
├── apps/            # Olares内蔵のシステムアプリ
├── cli/             # olares-cli：Agent Skillsを備えた、Olaresのインストールと操作のためのエージェントネイティブCLI
├── daemon/          # olaresd（システムデーモンプロセス）
├── docs/            # プロジェクトのドキュメント
├── framework/       # Olaresのシステムサービス
├── infrastructure/  # コンピュート、ストレージ、ネットワーク、GPUのコンポーネント
├── platform/        # データベースやメッセージキューなどのクラウドネイティブコンポーネント
└── vendor/          # Olaresデバイス向けのハードウェア固有コード
```

## 貢献

Olaresはプロジェクト全体への貢献を歓迎します。改善したいことに合わせて、次から選んでください。

- **コア開発**：まずは[未対応のissue](https://github.com/beclab/Olares/issues)を見てみてください。大きな変更に取りかかる前に、issueで方針を相談しておくと安心です。
- **ドキュメント**：[`docs/`](./docs) のガイドを改善しましょう。[ドキュメント貢献ガイド](./docs/README.md)と[コンテンツ・スタイルガイド](https://github.com/beclab/Olares/wiki/General-style-reference)を参照してください。
- **アプリの配布**：アプリを[パッケージ化して申請](https://www.olares.com/docs/developer/develop/distribute-index)し、Olares Marketに公開できます。
- **バグ報告・機能要望**：調査や検討がしやすいよう、背景を添えて [GitHub issue](https://github.com/beclab/Olares/issues) を作成してください。
- **セキュリティ報告**：[セキュリティポリシー](./SECURITY.md)に従ってください。脆弱性は、公開のissueやディスカッション、コミュニティチャンネルでは報告しないでください。

## さらに詳しく

- **[インストールガイド](https://www.olares.com/docs/manual/get-started/install-olares)**：インストール方法を選んでOlaresをアクティベーションします。
- **[ユースケース](https://www.olares.com/docs/use-cases/)**：ローカルAI、メディア、生産性、セルフホストの活用例を紹介します。
- **[CLIガイド](https://www.olares.com/docs/developer/install/cli/olares-cli)**：コマンドラインからOlaresをインストール・管理・診断します。
- **[Agent Skills](https://www.olares.com/docs/developer/cli-agent-skills)**：AIエージェントが `olares-cli` を通じてOlaresを操作できるようにします。
- **[高度なチュートリアル](https://www.olares.com/docs/manual/best-practices/)**：GPU、マルチノード構成、カスタムドメイン、ストレージ拡張などを設定します。

## コミュニティ

- **[Discord](https://discord.gg/olares)**：コミュニティのサポートを受けつつ、デプロイやエージェントの活用について話せます。
- **[Olaresフォーラム](https://www.olares.com/forum/)**：製品へのフィードバックを共有し、じっくり議論できます。
- [X](https://x.com/Olares_OS) と [YouTube](https://www.youtube.com/@OlaresOS) でOlaresをフォローしてください。

## 謝辞

Olaresは、数多くの優れたオープンソースプロジェクトの上に成り立っています。この場を借りて感謝します：[Kubernetes](https://kubernetes.io/)、[Kubesphere](https://github.com/kubesphere/kubesphere)、[Padloc](https://padloc.app/)、[K3S](https://k3s.io/)、[JuiceFS](https://github.com/juicedata/juicefs)、[MinIO](https://github.com/minio/minio)、[Envoy](https://github.com/envoyproxy/envoy)、[Authelia](https://github.com/authelia/authelia)、[Infisical](https://github.com/Infisical/infisical)、[Dify](https://github.com/langgenius/dify)、[Seafile](https://github.com/haiwen/seafile)、[HeadScale](https://headscale.net/)、[Tailscale](https://tailscale.com/)、[Redis Operator](https://github.com/spotahome/redis-operator)、[Nitro](https://nitro.jan.ai/)、[RSSHub](http://rsshub.app/)、[predixy](https://github.com/joyieldInc/predixy)、[nvshare](https://github.com/grgalex/nvshare)、[LangChain](https://www.langchain.com/)、[Quasar](https://quasar.dev/)、[TrustWallet](https://trustwallet.com/)、[Restic](https://restic.net/)、[ZincSearch](https://zincsearch-docs.zinc.dev/)、[filebrowser](https://filebrowser.org/)、[lego](https://go-acme.github.io/lego/)、[Velero](https://velero.io/)、[s3rver](https://github.com/jamhall/s3rver)、[Citusdata](https://www.citusdata.com/)。
