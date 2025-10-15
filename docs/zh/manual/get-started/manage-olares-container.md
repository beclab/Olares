---
search: false
---
## 管理 Olares 容器

### 停止容器
要停止运行中的容器：
```bash
docker stop oic
```

### 重启容器
容器停止后，使用以下命令重启：
```bash
docker start oic
```
容器重启后，所有服务可能需要 6–7 分钟才能完全初始化。在此时间内请耐心等待。

### 卸载容器
要完全移除容器及其关联数据：
```bash
docker stop oic
docker rm oic
docker volume rm oic-data
```