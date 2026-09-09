---
description: 检查磁盘用量，删除未使用的 AI 模型和应用镜像，并确认释放出的空间。
---
# 释放磁盘空间

如果出现「Authentication failed, disk space is full」等错误，或应用安装因存储空间耗尽而失败，请检查磁盘用量并清理不再需要的内容。

## 前提条件

- 拥有 Olares 的管理员访问权限。
- 拥有 Olares 的终端访问权限，用于移除未使用的镜像。请参考[通过 SSH 连接 Olares](/zh/manual/access-olares-terminal)。

## 检查磁盘用量

1. 打开仪表盘，查看系统和各应用的磁盘用量。详细信息请参考[查看系统和应用资源使用情况](olares/resources-usage.md)。
2. 在「文件」中检查可以迁移到外部存储或删除的大文件和文件夹。

## 删除未使用的 AI 模型

AI 模型通常是占用磁盘空间最多的内容。

1. 在「文件」中打开「Home」目录。
2. 删除「Huggingface」和「Ollama」文件夹中未使用的模型。这些文件夹按对应应用命名，方便识别。

## 移除未使用的应用镜像

应用和软件包镜像会随使用时间不断累积。在 Olares 1.12.6 及以上版本，可以用 `olares-cli` 列出未使用的镜像并删除：

```bash
olares-cli doctor images --unused --no-headers | awk '{print $1}' | xargs -r sudo crictl rmi
```

:::warning
不要使用 `sudo crictl rmi --prune`。该命令会移除所有已停止应用的镜像，可能拖慢应用恢复速度，并导致界面显示问题。
:::

## 确认释放出的空间

返回仪表盘，确认磁盘用量已下降。如果系统仍提示磁盘空间不足，请重复以上步骤，或扩展系统存储。更多信息请参考[扩展系统存储](best-practices/expand-storage-in-olares.md)。
