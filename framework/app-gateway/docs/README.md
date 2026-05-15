# app-gateway 运维速查

完整说明：  
[archdoc/方案/app-gateway/app-gateway-Olares发版与安装说明-2026-05-15.md](../../../../archdoc/方案/app-gateway/app-gateway-Olares发版与安装说明-2026-05-15.md)

---

## kubeconfig（Olares 节点补装）

`install-app-gateway` 与 `kubectl` 使用同一 kubeconfig 加载规则。

| 场景 | 路径 |
|------|------|
| Olares **控制面本机**（推荐） | `~/.kube/config`（可省略 `--kubeconfig`） |
| K3s 真源 | `/etc/rancher/k3s/k3s.yaml` |
| 远程连 API | kubeconfig 中 `server` 须为 **控制面 IP:6443**，非 `127.0.0.1` |

```bash
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"
kubectl get nodes
olares-cli install-app-gateway --installer-dir /path/to/.dist-agw
```

---

## 发行包根目录（`--installer-dir`）

须含 `wizard/config/app-gateway-vendor/` 与 `wizard/config/apps/app-gateway/Chart.yaml`。

```bash
export OLARES_SOURCE_ROOT=/path/to/Olares
export OLARES_INSTALLER_DIR=/path/to/.dist-agw
bash /path/to/devops/dev/platform-gateway/scripts/package-olares-installer-slice.sh
```

---

## 版本

Linkerd chart **2026.5.1**，Envoy Gateway **v1.8.0** — `VENDOR_VERSION.lock.yaml`、`pkg/packaging/versions.go`
