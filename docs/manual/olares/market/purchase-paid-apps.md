---
outline: [2, 3]
description: Learn how to purchase paid applications on Olares. This guide details the complete process, from configuring your crypto wallet to verifying identity and confirming payments via the LarePass mobile app.
---

# Purchase paid applications

You can purchase paid applications directly from the Olares Market. The payment process requires coordination between the Olares system, the LarePass mobile app, and a digital wallet such as MetaMask.

## Prerequisites

Before you begin, ensure that you have completed the following tasks:
- You have installed and logged into a browser wallet extension, such as MetaMask. To successfully complete the transaction, ensure that your wallet holds both of the following funds:
    - Application fee: Used to pay for the software license. The specific token required is determined by the app's `price.yaml` configuration, typically USDC or USDT.
    - Network gas fee: Used to pay for blockchain transaction costs. Note that gas fees must be paid in ETH.
- You have installed the LarePass app on your phone and logged into the same account as your Olares system.

## Step 1: Initiate purchase and identity verification

1.  Open the Olares Market.
2.  Search for the target paid app, click **Get**, and then click **Buy**.
3.  In the **Verify payment access** popup, click **Verify now**.
4.  Open your LarePass mobile app:
    
    a. Review the request popup titled **Authorize payment access**.
    
    b. Tap **Confirm** .

    c. If the popup does not appear, return to the app page in Olares Market and click **Buy** again to re-initiate the request.

    ![Authorize payment access](/images/manual/olares/authorize-pay.png#bordered)

## Step 2: Confirm order

1.  After successful verification, return to the app page in Olares Market and click **Buy** again.
2.  In the **Verify purchase request** popup, click **Verify now**.
3.  Check your LarePass mobile app again:
    
    a. Review the request popup titled **Authorize purchase**.
    
    b. Tap **Confirm**.

    ![Authorize purchase](/images/manual/olares/authorize-purchase.png#bordered)

## Step 3: Complete payment via wallet

1. Return to the app page in Olares Market and click **Pay**.
2. The browser will automatically open **MetaMask** or your connected wallet extension.
3. Review the amount in the wallet popup, and then confirm and approve the transaction.
4. Wait for the blockchain to confirm the transaction.
5. Once the payment is successfully completed, you can proceed to install the application by clicking **Install**.
6. After the payment is completed, you can verify your purchase credential from your LarePass mobile app by navigating to **Settings** > **Safety** > **VC cards**.

![Authorize purchase](/images/manual/olares/vc.png#bordered)
