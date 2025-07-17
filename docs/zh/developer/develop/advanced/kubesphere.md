# Kubesphere

Olares 集成了 Kubesphere 的许多高级功能，如多用户系统和集群数据监控。要从 Kubesphere** 安装官方 console 工具，请从 Olares 代码存储库下载并安装它。

```sh
curl -LO https://github.com/Above-Os/terminus-os/raw/main/third-party/ks-console/ks-console-v3.3.0.tgz

# username 为 Olares 的登录用户
sudo helm install console ./ks-console-v3.3.0.tgz \
    -n user-space-<username> \
    --set username=<username>
```

安装后，刷新桌面。即可在 Olares 中看到 Console 的图标。 打开 Console，可用 Olares ID 和密码登录。
