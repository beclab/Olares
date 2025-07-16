# Kubesphere

Olares has integrated many advanced features of Kubesphere like the multi-user system and cluster data monitoring. To install the official console tool from **Kubesphere**, download and install it from the **Olares** code repository.

```sh
curl -LO https://github.com/Above-Os/terminus-os/raw/main/third-party/ks-console/ks-console-v3.3.0.tgz

# The username is your Olares ID
sudo helm install console ./ks-console-v3.3.0.tgz \
    -n user-space-<username> \
    --set username=<username>
```

After you install it, refresh your desktop. You'll then see the Console icon in Olares. Open Console and log in with your Olares ID and password.
