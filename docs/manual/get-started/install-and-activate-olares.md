---
noindex: true
---
## Finish installation and activate Olares
:::warning Same network required
To avoid activation failures, ensure that both your phone and the Olares device are connected to the same network.
:::

<!-- #region iso-activation-flow -->
1. Open LarePass. If you do not have an Olares ID, tap **Create an account** and follow the prompts.
2. On the activation page, tap **Discover nearby Olares**. LarePass lists the Olares instances detected on the same network.
3. Select your Olares instance and tap **Install now**.

   ![ISO Activate](/images/manual/larepass/iso-activate1.png#bordered)

4. When the installation completes, tap **Activate now**.
5. In the **Select a reverse proxy** dialog, select a node that is closer to your geographical location. The installer will then configure HTTPS certificate and DNS for Olares.

   :::tip Note
   - You can change this setting later on the [Change reverse proxy](../olares/settings/change-frp.md) page in Olares.
   - If your Olares device is connected to a public IP network, this step will be skipped automatically.
   :::
6. Select the language for Olares. Olares supports English, Simplified Chinese, German, Spanish, Italian, French, and Japanese.

   :::info
   This selection does not change the language of the LarePass app. The remaining activation steps stay in the current LarePass language. After activation, Olares Desktop uses the language selected here.
   :::
7. Follow the on-screen instructions to set the login password for Olares, then tap **Complete**.

   ![ISO Activate-2](/images/manual/larepass/iso-activate-4.png#bordered)

Once activation is complete, LarePass will display the desktop address of your Olares device, such as `https://desktop.marvin123.olares.com`.
<!-- #endregion iso-activation-flow -->
