---
description: 通过 Wise 过滤视图示例，按阅读状态、日期、作者和标签查找订阅内容及已保存的内容。
head:
  - - meta
    - name: keywords
      content: Olares, Wise, 过滤视图示例, RSS 筛选, 标签筛选, 内容管理
---

# 过滤视图示例

以下查询示例可以帮你查找订阅源中的内容，以及你选择保存的内容。

## 查看特定订阅源的未读内容

想集中查看某个订阅源的未读内容，可以使用：

```
feed_id:12345 AND seen:false
```

将 `12345` 替换为实际订阅源 ID。要查找 ID，请打开**设置** > **RSS 订阅源**并编辑该订阅源。

## 查看某位作者的内容

按作者查找内容：

```
author:"John Doe"
```

此查询显示作者为“John Doe”的内容。

## 查看最近发布的内容

按发布日期筛选：

```
published_at__gt:2025-04-01
```

此查询显示 2025 年 4 月 1 日之后发布的内容。

## 查看带笔记的已保存内容

查找同时带有特定标签和笔记的已保存内容：

```
tag:AI AND has:note AND islibrary:true
```

## 查看稍后阅读的特定主题

查找带有“Work”标签且已加入**稍后阅读**的内容：

```
tag:Work AND location:readlater
```

## 查看订阅源中的未读视频

查找订阅源中的未读视频：

```
file_type:video AND seen:false AND isfeed:true
```

## 找回很久没有打开的内容

查找上次打开时间早于指定日期的已保存内容：

```
last_opened__lt:2023-12-31 AND tag:Productivity AND islibrary:true
```

此查询可用于找回你之前保存、带有“Productivity”标签的内容。
