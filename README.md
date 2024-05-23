# Terminus OS

<p align="center">
    <picture>
         <img alt="Terminus OS" src="https://raw.githubusercontent.com/beclab/terminus/main/images/banner1.jpg"/>
    </picture>
    <i>Let people own their data again</i>
    <br>
  <a href="https://www.jointerminus.com">Website</a> ·
  <a href="https://docs.jointerminus.com">Documentation</a> ·
  <a href="https://docs.jointerminus.com/how-to/termipass/overview.html#download">Download TermiPass</a> ·
  <a href="https://github.com/beclab/apps">Terminus Apps</a> ·
  <a href="https://space.jointerminus.com">Terminus Space</a>
</p>


## Introduction

Terminus OS is a source-available, cloud-native operating system built on Kubernetes, designed to run on edge devices owned by users. Our goal is to enable users to securely store their most important data on their own hardware and access services based on this private data from anywhere in the world.

In essence, we want you to use Terminus OS like a regular computer. We hope that Terminus OS can assist individuals and organizations in managing data, business, and life effectively, all while fully owning and controlling their data.

- For users, we aim to make Terminus OS as easy to use as a smartphone.
- For developers, we strive to provide an experience consistent with that of public clouds.

## Features

Terminus OS offers a wide array of features designed to enhance security, ease of use, and development flexibility, making it a powerful tool for both users and developers.

- [**Enterprise-Grade Security with Ease**](https://docs.jointerminus.com/overview/terminus/network.html)

   Terminus seamlessly integrates Tailscale, Headscale, Cloudflare Tunnel, and FRP, simplifying network configuration while providing enterprise-grade security. Users no longer need to worry about managing domain names, HTTPS certificates, and other details; each service can be accessed in the most secure and convenient way.

- [**Secure and Permissionless Application Ecosystem**](https://docs.jointerminus.com/overview/terminus/application.html)

   Terminus offers a secure and permissionless app ecosystem via sandboxing, ensuring application isolation and security. Developers can freely distribute and run applications without the constraints of traditional app stores.

- [**Manage Data with Peace of Mind**](https://docs.jointerminus.com/overview/terminus/data.html)

   Terminus provides a unified filesystem and database at the OS level, with the OS handling scaling, backups, and high availability.

- [**One Login for All Applications**](https://docs.jointerminus.com/overview/terminus/account.html)
  
   Terminus offers a seamless integration with third-party application accounts, allowing users to log in the system once and access all applications within Terminus. Say goodbye to the hassle of logging into each app individually.

- [**Effortlessly Enjoy AI Benefits**](https://docs.jointerminus.com/overview/terminus/ai.html)
   
   Terminus provides a one-stop solution for GPU management, model hosting, private knowledge base maintenance, and agent and workflow construction. Users can enjoy the benefits of AI without writing any code, all while protecting their privacy.

- [**Versatile Built-in Applications**](https://docs.jointerminus.com/how-to/terminus/)
  
   Terminus comes with a suite of built-in applications such as a file manager, sync drive, vault, reader, app marketplace, settings, and dashboard, allowing users to use Terminus as easily as they would use a smartphone or any desktop. 

- [**Access Your Devices Anytime, Anywhere**](https://github.com/beclab/TermiPass)
   
   Terminus offers various clients, including mobile, desktop, and browser extensions, enabling users to access their machines anytime, anywhere.

- [**Easily Port and Develop Applications**](https://docs.jointerminus.com/overview/terminus/network.html)
  
   Terminus provides development tools to help users port existing applications to Terminus or develop new ones.


## Getting Started

Before you get started, make sure your hardware meet the following minimum system requirements:

- Hardware congigurations

  - CPU >= 4 Core
  - RAM >= 8GB
  - Free Disk >= 100GB
- Supported systems:
   
   | Linux Version | Architecture |
   | -------------- | ------ |
   | Ubuntu 24.04   | x86-64, amd64 |
   | Ubuntu 22.04   | x86-64, amd64 |
   | Ubuntu 20.04   | x86-64, amd64 |
   | Debian 12  | amd64 |
   | Debian 11  | amd64 |
      

1. [Apply for A Terminus Name](https://docs.jointerminus.com/how-to/termipass/account/#create-terminus-name). 
   
2. Install Terminus in your machine with the following command: 
   ```
   curl -fsSL https://terminus.sh |  bash -
   ```
   For more detailed instructions, see [Install Terminus with commands](https://docs.jointerminus.com/how-to/terminus/setup/install.html#install).

3. Access the URL required for Terminus activation in the browser, and complete the initial setups and system activation following the on-screen instructions. For more detailed instructions, see the [Activation Guide](../../how-to/terminus/setup/wizard.md).
   
4. Log in with the password you reset during activation and complete two-step verification on TermiPass. For more detailed instructions, see the [Login Doc](../../how-to/terminus/setup/login.md).
   
5. [Back up your mnemonic phrase](../../how-to/termipass/account/index.md#backup-mnemonic-phrase.md) to ensure account and data security.



## Contributing

We are welcoming anyways of contributions:

- If you want to develop your own applications on Terminus, refer to:
https://docs.jointerminus.com/developer/develop/


- If you want to help improve Terminus, refer to:
https://docs.jointerminus.com/developer/contribute/terminus-os.html

## Community & Contact

* [**Github Discussion**](https://github.com/beclab/terminus/discussions). Best for sharing feedback and asking questions.
* [**GitHub Issues**](https://github.com/beclab/terminus/issues). Best for filing bugs you encounter using Terminus and submitting feature proposals. 
* [**Discord**](https://discord.gg/ShjkCBs2). Best for sharing your applications and hanging out with the community.

## Staying Ahead

Star Terminus on GitHub and be instantly notified of new releases and status updates. 

<p align="center">
    <picture>
        <img alt="star Terminus OS" src="https://raw.githubusercontent.com/beclab/terminus/main/images/star.gif"/>
    </picture>
</p>


## Special Thanks 

The Terminus OS project has incorporated numerous third-party open source projects, including: [Kubernetes](https://kubernetes.io/), [Kubesphere](https://github.com/kubesphere/kubesphere), [Padloc](https://padloc.app/), [K3S](https://k3s.io/), [JuiceFS](https://github.com/juicedata/juicefs), [MinIO](https://github.com/minio/minio), [Envoy](https://github.com/envoyproxy/envoy), [Authelia](https://github.com/authelia/authelia), [Infisical](https://github.com/Infisical/infisical), [Dify](https://github.com/langgenius/dify), [Seafile](https://github.com/haiwen/seafile),[HeadScale](https://headscale.net/), [tailscale](https://tailscale.com/), [Redis Operator](https://github.com/spotahome/redis-operator), [Nitro](https://nitro.jan.ai/), [RssHub](http://rsshub.app/), [predixy](https://github.com/joyieldInc/predixy), [nvshare](https://github.com/grgalex/nvshare), [LangChain](https://www.langchain.com/), [Quasar](https://quasar.dev/), [TrustWallet](https://trustwallet.com/), [Restic](https://restic.net/), [ZincSearch](https://zincsearch-docs.zinc.dev/), [filebrowser](https://filebrowser.org/), [lego](https://go-acme.github.io/lego/), [Velero](https://velero.io/), [s3rver](https://github.com/jamhall/s3rver), [Citusdata](https://www.citusdata.com/).

## Contributors
<a href="https://github.com/beclab/terminus/graphs/contributors"> <img src="https://contrib.rocks/image?repo=beclab/terminus" /> </a>