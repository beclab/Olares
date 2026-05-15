# app-gateway 运维速查

完整说明（发版、版本、闭环）：  
[archdoc/方案/app-gateway/app-gateway-Olares发版与安装说明-2026-05-15.md](../../../../archdoc/方案/app-gateway/app-gateway-Olares发版与安装说明-2026-05-15.md)

---

## 发行包根目录（`--installer-dir`）怎么制作

须包含：

- `wizard/config/app-gateway-vendor/`（四个 chart + `VENDOR_VERSION.lock.yaml`）
- `wizard/config/apps/app-gateway/Chart.yaml`

**最小包（已装 Olares、仅补装 app-gateway）：**

```bash
export OLARES_SOURCE_ROOT=/path/to/Olares
export OLARES_INSTALLER_DIR=/path/to/.dist-agw
bash /path/to/devops/dev/platform-gateway/scripts/package-olares-installer-slice.sh
# 发行包根目录 = $OLARES_INSTALLER_DIR
```

**完整 Olares 发行包：**

```bash
cd /path/to/Olares
bash build/package.sh
# 发行包根目录 = .dist（或 $DIST_PATH）
```

---

## 已在运行的 Olares 集群上安装

```bash
# 1) 制作发行包（见上）
export INSTALLER=/path/to/Olares/.dist-agw

# 2) 使用含 install-app-gateway 的 olares-cli
export KUBECONFIG=/path/to/kubeconfig
olares-cli install-app-gateway \
  --installer-dir "${INSTALLER}" \
  --kubeconfig "${KUBECONFIG}"
```

验证：`kubectl -n linkerd get deploy`；`kubectl -n app-gateway get deploy,gatewayclass`

---

## 版本与代码

- 审批版本：Linkerd chart **2026.5.1**，Envoy Gateway **v1.8.0**
- `pkg/packaging/versions.go`、`.olares/config/app-gateway-vendor/VENDOR_VERSION.lock.yaml`
