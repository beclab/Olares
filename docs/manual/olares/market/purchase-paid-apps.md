---
outline: [2, 3]
description: Learn how to purchase paid applications on Olares Market and restore previously purchased applications.
---

# Manage paid applications

You can purchase paid applications directly from the Olares Market, or restore your previously purchased apps without additional payment.

## Purchase paid applications

You can purchase paid applications directly from the Olares Market. The payment process requires coordination between the Olares system, the LarePass mobile app, and a digital wallet such as MetaMask.

### Prerequisites

Before you begin, ensure that you have completed the following tasks:
- You have installed and logged into a browser wallet extension, such as MetaMask. To successfully complete the transaction, ensure that your wallet holds both of the following funds:
    - Application fee: Used to pay for the software license. The specific token required is determined by the app's `price.yaml` configuration, typically USDC or USDT.
    - Network gas fee: Used to pay for blockchain transaction costs. Note that gas fees must be paid in ETH.
- You have installed the LarePass app on your phone and logged into the same account as your Olares system.

### Step 1: Initiate purchase and identity verification

1.  Open the Olares Market.
2.  Search for the target paid app, click **Get**, and then click **Buy**.
3.  In the **Verify payment access** popup, click **Verify now**.
4.  Open your LarePass mobile app:
    
    a. Review the request popup titled **Authorize payment access**.
    
    b. Tap **Confirm**.

    c. If the popup does not appear, return to the app page in Olares Market and click **Buy** again to re-initiate the request.

    ![Authorize payment access](/images/manual/olares/authorize-pay.png#bordered)

### Step 2: Confirm order

1.  After successful verification, return to the app page in Olares Market and click **Buy** again.
2.  In the **Verify purchase request** popup, click **Verify now**.
3.  Check your LarePass mobile app again:
    
    a. Review the request popup titled **Authorize purchase**.
    
    b. Tap **Confirm**.

    ![Authorize purchase](/images/manual/olares/authorize-purchase.png#bordered)

### Step 3: Complete payment via wallet

1. Return to the app page in Olares Market and click **Pay**.
2. The browser will automatically open **MetaMask** or your connected wallet extension.
3. Review the amount in the wallet popup, and then confirm and approve the transaction.
4. Wait for the blockchain to confirm the transaction.
5. Once the payment is successfully completed, click **Install** to install the application.
6. After the payment is completed, navigate to **Settings** > **Safety** > **VC cards** to verify your purchase credential from your LarePass mobile app.

![Authorize purchase](/images/manual/olares/vc.png#bordered)

## Restore purchased applications

If you have previously purchased a paid application, you can restore and re-install it without additional payment, whether you simply uninstalled the app, reset your system, or migrated to a new Olares host.

### Re-install after uninstallation

If you are re-installing a paid app that was uninstalled from your current Olares system, the system preserves your license locally. 

You can find the app in the Market, click **Get**, and then click **Install** to re-install it.

### Restore after system reset or migration

If you have reset Olares OS or switched to new hardware, your local license data is cleared. You must verify your on-chain identity to retrieve your purchase rights, which will allow you to re-install the application.

1. Ensure that the current Olares OS is activated with the same **Olares ID** used for the original purchase.
2. Open the Olares Market.
3. Search for the target paid app, click **Get**, and then click **Buy**.

    :::info
    The system will detect your on-chain purchase record and will not charge you again.
    :::

4. In the **Authorize payment access** popup, click **Verify now**.
    
    :::info
    This step verifies your identity ownership and does not initiate a new transaction.
    :::

5. Once verified, the button will change to **Install**. Click it to start installation.

:::info Identity binding
Your purchase history binds to your Olares ID rather than specific hardware. If you have access to LarePass or a backup of your mnemonic phrase, you can activate the new system with the same Olares ID and restore your apps.
:::

:::warning Risk of data loss
If you lose your mnemonic phrase, you permanently lose your digital identity and all paid apps. In this case, you must create a new account and repurchase the application.
Note that license restoration also requires the developer's service to be online. If their service is unavailable, you might be unable to restore the application.
:::
