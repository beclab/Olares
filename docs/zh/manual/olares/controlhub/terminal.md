---
description: 通过终端模块一键登入 Olares 宿主机终端，执行故障排查，日志查看和导出，以及修改系统配置等任务。
---
# 访问终端

控制面板左侧下方的**终端**模块提供一键登入 Olares 宿主机（Linux）终端的功能。这相当于通过 Root 身份以 SSH 连接到系统，但无需手动输入账号和密码。

![Terminal](/images/zh/manual/olares/controlhub-terminal.png#bordered)

你可以使用终端来：
- 在主机上直接进行调试和故障排查。
- 实时查看日志或回溯历史系统日志。
- 在操作系统层面修改系统配置。

:::warning 提示
这里执行的操作可能会直接影响宿主机环境，请谨慎操作。
:::

## 示例命令

```bash
# 查看 GPU 驱动和服务状态
olares-cli gpu status

# 收集所有日志（使用默认设置）
olares-cli logs

# 收集指定组件的日志
olares-cli logs --components k3s,redis,minio

# 检查 NVIDIA GPU 是否正常工作
nvidia-smi

# 检查系统是否识别到 NVIDIA GPU 硬件
lspci | grep -i vga | grep -i nvidia
```

更多命令请参考 [Olares CLI 参考文档](../../../developer/install/cli/olares-cli.md)。
