# appid / entranceid 实施记录

日期：2026-10-29。

## 代码范围

| 仓库 | 分支 | 改动 |
| --- | --- | --- |
| terminus | refactor/appid，基于 refactor/user-domain-model / ada65a709 | 生命周期分配与持久化、Settings 全局查重、BFL/Helm/OIDC/网关/L4/Provider 读取持久化 ID、CLI 一次性迁移任务、测试及构建脚本 |
| api | refactor/appid，基于原 main | Entrance.id、Terminus.localName、CRD、按名称保留 ID 的分配接口、全局占用检查、共同写锁、生成客户端与测试 |
| proxylistener | refactor/appid，基于原 main | 新开发入口直接保存用户前缀 + 端口 ID，Pod IP 刷新保留 ID，乐观锁与共同写锁、角色模板与测试 |

入口生成只发生在新建和升级回填中。既有入口按名称保留 ID；正常读取/更新遇到缺失 ID 会报错并提示升级。sharedEntrance 的历史算法保持不变，YAML 默认别名未修改。

Terminus 生成客户端原有资源路径是 `terminus`，与 CRD 的 `terminuses` 不一致；已修正代码生成配置并重新生成。对应 getter 现在为 `Terminuses()`，调用方已更新。

## 一次性升级

入口为 `cli/pkg/upgrade/breakchange_entrance_ids.go` 的 `EntranceIDsBreakingChangeTasks()`。版本号待发布时绑定，当前未注册到任意版本。

- 固定计划和完整原配置备份：`storage.OlaresRootDir/upgrade/entrance-ids-v1.json`，0600，不覆盖；旁边的 `.progress/` 保存逐应用收据。
- 集群状态：`os-system/entrance-ids-v1` ConfigMap，保存计划摘要、原始用户、Terminus UID 和 planned/complete 状态。
- 已完成直接跳过，不扫描 Application；中断后读取原计划，不重新计算冲突组。备份丢失/摘要变化、身份变化、CR 被替换、Spec 漂移或 ID 被 CRD 裁剪均报错。
- 使用当前 resourceVersion 写入；只清除重复 Settings 路由的 `third_level_domain`，保留第三方域名、证书和其他字段。Owner 优先，仅非 Owner 重复则全部清除，清除后可重新占用。
- Settings、控制器、开发入口生产者与升级使用同一可续租 Lease；直接写 CR 或旧服务不受这把锁保护，发布维护窗口仍需停止这些写入。

## 本地验证

可复现命令（需要 sibling api / proxylistener checkout）：

```sh
ENTRANCE_RUN_TESTS=1 ENTRANCE_BUILD_OUTPUT=/private/tmp/appid-bin \
  GOCACHE=/private/tmp/appid-go-cache sh build/entrance-ids.sh
```

已通过：

- API 新增的 ID 分配、重排/冲突、开发端口、shared 历史算法、序列化和全局占用测试；Lease 并发串行化测试。
- app-service 的 appcfg、gateway、apiserver、appinstaller 包测试，以及本次控制器身份/持久化测试。
- BFL 受影响 URL 包、L4 provider/translator、system-server provider registry 包测试。
- CLI 新增迁移包和完整 upgrade 包测试（LC_ALL=C，避免 macOS 缺少 C.UTF-8 导致 shell 输出断言失败）。
- proxylistener 全部测试，包含非原始用户开发 ID、Pod IP 更新保号和缺失 ID 拒绝回填。
- 六个本地二进制构建：app-service、BFL、L4、system-server、CLI、proxylistener。产物目录 `/private/tmp/appid-bin`；未部署。

API 完整 `api/app.bytetrade.io/v1alpha1` 测试仍有两项既有失败：

1. `TestThirdLevelCusDomainPrefixes/empty_zone_returns_bare_prefixes`（nil / 空 slice 差异）。
2. `TestEntrancesWithZone/multiple_entrances_with_defaultThirdLevelDomainConfig_override`（旧别名 fixture 预期不一致）。

两项均用 HEAD 原始 `helper.go` 的 Go overlay 复现；未为消除这些既有失败而修改旧接口行为。完整 API suite 不计为通过。

## 发布前剩余工作

1. 发布配套 API，更新各组件 go.mod / go.sum，脱离临时 workspace 后独立构建。
2. 选定升级版本并绑定 breaking change task，按计划安排 CRD/兼容组件/停写/迁移/恢复顺序。
3. 发布 CRD 与实际 chart RBAC；proxylistener 的角色示例尚需按 service account 添加绑定。
4. 在真实集群验证目标 Pod、共享认证、同 zone TLS/SNI、域名切换及回滚。

本次未提交、推送、发布 API 或修改运行集群。
