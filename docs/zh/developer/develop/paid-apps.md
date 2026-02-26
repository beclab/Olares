---
outline: [2, 3]
description: 学习如何发布付费应用，并使用 Merchant 管理销售与收款。
---
# 发布付费应用

从 Olares v1.12.3 版本开始，Olares 应用市场支持付费应用分发。要销售应用，开发者必须在区块链上注册 Olares ID (DID) 并安装 Merchant 应用来管理许可证、订单和收款记录。

:::info 封闭测试中
该功能目前处于封闭测试阶段。如需发布付费应用，请通过邮件 [hi@olares.com](mailto:hi@olares.com) 联系我们申请。

目前仅支持付费下载应用。应用内购买（IAP）与订阅收费模式仍在开发中。
:::

本文将介绍如何发布付费应用并配置 Merchant。

## 前提条件

开始前，请确保你已具备：

- **Olares ID**：有效的 Olares ID（例如 `alice123@olares.com`）。
- **Olares OS**：运行 v1.12.3 或更高版本的 Olares 主机。
- **开发者资格**：已获得开发者资格（通过邮件申请）。
- **本地环境**：本地计算机已安装 Node.js。

## 设置钱包并充值 Gas 费

要启用支付功能，需要在区块链上注册你的 Olares ID (DID)。这属于链上交互，需要少量的网络手续费 (Gas Fee)。

### 设置加密钱包

本教程以 MetaMask 为例，你也可以使用其他加密钱包。

1. 在手机端打开 **LarePass**。
2. 进入**设置** > **安全** > **助记词**查看助记词。
   :::warning 安全提示
   助记词是账户的最高权限凭证。切勿将其泄露给他人或上传至公共代码库。请在安全的本地环境中执行以下操作。
   :::
3. 在浏览器中安装 MetaMask 插件。
4. 在 MetaMask 中选择**添加钱包** > **导入钱包**，输入你的 Olares 助记词完成导入。

### 充值 Gas 费

注册过程会消耗 Optimism 网络上的少量 ETH 作为手续费。

1. 在 MetaMask 中切换网络至 **Optimism (OP Mainnet)**。
2. 复制你的钱包地址（以 `0x` 开头）。

   ![Copy MetaMask wallet address](/images/developer/develop/distribute/metamask-wallet-add.png#bordered){width=50%}

3. 向该地址转入少量 ETH。建议至少 **0.0005 ETH** 以支付后续的 Gas 费。

## 生成并注册 RSA 密钥

为确保交易安全与许可证唯一性，你需要生成 RSA 密钥对，并将公钥注册到你 Olares ID 的链上记录中。

### 生成 RSA 密钥对

1. 安装 `did-cli` 工具：
    ```bash
    npm install -g @beclab/olaresid
    ```
2. 生成默认的 2048 位 RSA 密钥对：
    ```bash
    did-cli rsa generate
    ```
    命令成功后，你会看到类似如下输出：
    ![RSA key pair generated](/images/developer/develop/distribute/metamask-rsa-key-pairs.png#bordered){width=60%}
3. 查看生成的密钥内容：
    ```bash
    cat ./rsa-public.pem
    cat ./rsa-private.pem
    ```
    :::info 安全警告
    请妥善保管 `rsa-private.pem`。稍后配置 Merchant 应用时需要用到它。
    :::
### 注册 RSA 公钥

将生成的公钥绑定到你的 Olares ID。此操作需要使用助记词进行签名，并会消耗 Gas。

1. 将助记词导出为临时环境变量，并将值用双引号包裹：
    ```bash
    export PRIVATE_KEY_OR_MNEMONIC="xx xxx xx ..."
    echo $PRIVATE_KEY_OR_MNEMONIC
    ```
2. 检查你的 Olares ID 是否已存在 RSA 公钥：
    ```bash
    did-cli rsa get alice123@olares.com --network mainnet
    ```
    若看到以下提示，说明未设置公钥，可以继续：
    
    ```text
    ❌ No RSA public key set for this domain
    ```
3. 验证 Owner 钱包地址，确保其与你之前充值的钱包一致：
    
    ```bash
    did-cli owner alice123@olares.com --network mainnet
    ```

    示例输出：

    ```text
    👤 Owner: 0x3.....
    ```
4. 将 RSA 公钥注册上链：
    
    ```bash
    did-cli rsa set alice123@olares.com ./rsa-public.pem --network mainnet
    ```
5. 验证注册结果：
    ```bash
    did-cli rsa get alice123@olares.com --network mainnet
    ```
6. 从环境中清除助记词：
    ```bash
    unset PRIVATE_KEY_OR_MNEMONIC
    echo $PRIVATE_KEY_OR_MNEMONIC
    ```
    若无输出，说明已清除。

## 配置你的应用（OAC）

付费应用与免费应用的 OAC 基本一致，你只需要额外完成两项配置：
- 在 Chart 根目录（OAC 根目录）添加 `price.yaml` 文件。
- 将 `VERIFIABLE_CREDENTIAL` 注入到应用中，以便应用读取用户的购买凭证。

### 添加定价配置

在 Chart 根目录下创建 `price.yaml`：
```yaml
# 开发者的 Olares 实例地址
developer: alice123.olares.com

# 付费应用
paid:
  # product_id 格式：repoName-appName-productId
  product_id: apps-appname-paid
  price:
    - chain: optimism | eth
      token_symbol: USDC | USDT
      receive_wallet: "0xcbbcd55960eC62F1dCFBac17C2a2341E4f0e81c8"
      product_prize: 100000
  description:
    - lang: en
      title: Purchase item title
      description: Purchase item description
      icon: https://item.icon.url

# 应用内购买或订阅项目（尚未实现）
products: []
```

### 启用许可证校验

你的应用可以通过注入的环境变量读取用户的购买凭证。

在需要进行付费校验的 Pod 模板中添加该环境变量：
```yaml
- name: VERIFIABLE_CREDENTIAL
  value: "{{ .Values.olaresEnv.VERIFIABLE_CREDENTIAL }}"
```

在 `OlaresManifest.yaml` 中声明该变量：
```yaml
envs:
  - envName: VERIFIABLE_CREDENTIAL
    required: true
    type: string
    editable: false
    applyOnChange: true
``` 
更多环境变量行为说明，请参考 [`OlaresManifest.yaml` 中的环境变量](/zh/developer/develop/package/manifest.md#envs)。

### 提交应用

按照[标准提交流程](/zh/developer/develop/submit-apps.md)创建 PR，提交你的应用。

## 安装并配置 Merchant

Merchant 应用是你的收银台和管理面板，可用于：
- 查看开发者信息（产品列表、RSA 密钥、收款钱包、订单等）。
- 发放产品许可证。
- 应用启动时验证许可证。

:::info 保持 Merchant 在线
购买请求需要实时回调你的 Olares 主机完成验证。请确保运行 Merchant 的主机 24 小时在线，网络连接稳定快速。主机离线或网络不稳定可能导致用户无法完成购买。
:::

### 安装 Merchant

1. 打开 Olares 应用市场，搜索“Merchant”。
2. 点击**获取**，然后点击**安装**。

### 初始化配置

安装完成后，从启动台打开 Merchant。
1. 在登录界面，输入你的开发者 Olares ID，并点击**导入开发者**。
2. Merchant 会自动加载链上产品信息与 RSA 公钥。
3. 打开你的私钥文件：
    ```bash
    cat ./rsa-private.pem
    ```
    复制完整内容，粘贴到 Merchant 的对话框中，然后点击**导入私钥**。

## 在 Merchant 中管理销售

设置完成后，Merchant 会打开 Home 面板，你可以在此监控身份状态、密钥配置、钱包资产和交易历史。

:::info 
左侧导航栏中的 Store 与 Buy App 页面目前仅用于调试/测试。在当前发布流程中可忽略。
:::

### 同步状态

交易列表上方的状态指示器支持手动刷新：
- `Synced`：已同步至最新。
- `Syncing`：正在从链上拉取数据。
- `Not synced`：未同步。
- `Error`：同步失败，请检查网络连接。

### 交易字段说明

| 字段 | 说明 |
|--|--|
| Wallet | Merchant 管理的开发者收款钱包地址。 |
| Type | 交易类型：<ul><li>`Incoming`：收入（作为接收方</li><li>`Outgoing`：支出（作为发送方）</li></ul>|
| Tx Hash | 交易记录哈希，链上交易的唯一标识符。 |
| From/To | 交易对手地址，根据交易类型显示：<ul><li>`Incoming`：发送方地址</li><li>`Outgoing`：接收方地址</li></ul> |
| Contract | 代币合约地址，表示资产类型与来源：<ul><li>**Native**：原生 ETH 转账，无合约交互</li><li>**[合约地址]**：ERC-20 代币智能合约地址</li></ul> |
| Amount | 交易金额，单位与 Contract 指定的代币一致。|
| Time | 交易在链上被确认并记录的时间戳。|
| Product | <ul><li>**Product ID**：与 Olares 应用购买相关的交易</li><li>**-**：与应用销售无关的普通转账 </li></ul>|

### 更新开发者信息

如需修改 RSA 密钥、更新产品信息、调整价格或更改收款地址：
1. [重新生成 RSA 密钥](#生成并注册-rsa-密钥)或[更新应用配置](#配置你的应用-oac)，然后提交包含更新信息的 PR。
2. 等待 PR 审核合并。
3. PR 合并后，从启动台打开 Merchant。收到提示时，点击**更新配置/重新安装**以应用最新配置。更改将立即生效。