---
search: false
---
## Finish installation and activate Olares
:::warning Same network required
To avoid activation failures, ensure that both your phone and the Olares device are connected to the same network.
:::

1. Open LarePass app on your phone.
2. On your account activation page, tap **Discover nearby Olares**. LarePass will list the detected Olares instances in the same network.
2. Select the target Olares instance from the list and tap **Install now**.
3. When the installation completes, click **Activate now**.
4. In the **Select a reverse proxy** dialog, select a node that is closer to your geographical location. The installer will then configure HTTPs certificate and DNS for Olares. 

   ![ISO Activate](/images/manual/larepass/iso-activate.png#bordered)
   :::tip Note
   - You can change this setting later on the [Change reverse proxy](../olares/settings/change-frp.md) page in Olares.   
   - If your Olares device is connected to a public IP network, this step will be skipped automatically.
   :::
5. Follow the on-screen instructions to set the login password for Olares, then tap **Complete**.


   ![ISO Activate-2](/images/manual/larepass/iso-activate-2.png#bordered)

Once activation is complete, LarePass will display the desktop address of your Olares device, such as `https://desktop.marvin123.olares.com`.
