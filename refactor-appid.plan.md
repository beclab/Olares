# appid / entranceid 重构执行计划

## 当前基线与执行范围

- 2026-10-29 已开始并完成本地实现；实施记录见 [refactor-appid.build.md](refactor-appid.build.md)。
- terminus 当前分支为 `refactor/appid`，从 `refactor/user-domain-model` 的 `ada65a709` 新建。
- 配套 `/Users/hys/code/beclab/api` 和 `/Users/hys/code/beclab/proxylistener` 同样已新建 `refactor/appid`，各自基于原 main；未提交、推送或部署。
- 需求来源：`/Users/hys/Downloads/用户重构-appid-entranceid 规则整理.md`，以及本对话中确认的补充规则。
- 实际升级目录是 `cli/pkg/upgrade/`；已提供 `EntranceIDsBreakingChangeTasks()`，具体发布版本尚待绑定，不虚构版本号。
- API 尚未发布，各组件使用临时 Go workspace 联合构建；发布 API 后仍需更新正式依赖并验证独立构建。

## 已确定的业务规则

定义 `H(x) = hex(MD5(x))[:8]`，入口索引从 0 开始。

| 对象 | appid | 普通 entranceid | sharedEntranceid |
| --- | --- | --- | --- |
| 域名原始用户 / 旧 Owner | 保持不变 | 单入口为 appid；多入口为 appid + 索引；升级时保存到 CR | 保持现有算法 |
| 共用域名的 admin / normal | 保持不变 | 前缀为 H(appid + ":" + user)；单入口直接使用前缀，多入口追加索引 | 保持现有算法 |
| shared 应用的普通入口 | 保持不变 | 统一使用安装者的持久化 ID，不随访问用户重新生成 | 与普通入口分开处理，保持原规则 |

开发入口 ID 单独按端口生成，不使用普通入口的位置索引：

| 开发入口所属用户 | 持久化 ID / 域名前缀 |
| --- | --- |
| Owner / 域名原始用户 | `<appid>-<port>` |
| 其他用户 | `H(appid + ":" + user)-<port>` |

开发入口的主访问域名和原有按端口访问路径统一读取该持久化 ID；其他用户不继续保留会冲突的 `<appid>-<port>` 旧别名。Owner 仍按上文的域名原始用户身份判断。

补充约定：

1. 域名原始用户按 `Terminus.spec.localName == user.Name` 判断，不使用当前 Owner/admin 角色替代。
2. **所有入口 ID 都必须持久化**：普通入口保存至 `Application.spec.entrances[].id`，共享入口保存至 `Application.spec.sharedEntrances[].id`；系统应用及用户新增入口同样纳入。用户新增入口若存于 Settings/UserSettings 的 addedEntrances，则在对应持久化入口对象中保存 ID。已有自定义路由 ID 继续保存在相应配置字段中，不混淆基础入口 ID 和别名。具体类型、JSON/YAML 标签及 CRD schema 一起交付。
3. 所有入口按名称保持身份。应用更新、入口重排和重复 reconcile 保留已有 ID；新增普通入口使用新生成规则并立即持久化，新增共享入口使用原 sharedEntranceid 算法并立即持久化。遇到冲突报错，不修改既有入口的 URL。
4. Settings 自定义路由 ID 保证当前全局唯一，覆盖所有用户和 shared 应用的用户覆盖设置。
5. 升级时 Owner 与其他用户重复：保留 Owner，清除其他用户的冲突值；只有非 Owner 重复：全部清除。
6. 清除后允许空闲 ID 被重新占用，不增加永久禁用记录。
7. **YAML 默认路由及其冲突由人工控制，代码层面不新增处理**。不自动追加用户名、不为此新增升级冲突扫描或升级阻断，也不自动清理 YAML 配置；保留现有别名行为。这不取消 Settings 自定义路由的全局唯一性要求。
8. 保留第三方域名、证书及其他无关配置。清理三级路由 ID 时不能删除整个 customDomain 对象。
9. **存量入口 ID 的回填/迁移代码只在系统升级时执行一次**。成功后写入完成标记，后续升级重试、服务启动、reconcile 和应用更新不再执行整批回填。中途失败允许恢复未完成部分，不重复迁移已完成项。
10. **升级完成后，正常安装和新增入口直接使用新生成规则并持久化**，不保留以迁移开关长期选择旧/新生成算法的运行时分支；读取已有入口直接使用已保存 ID。
11. **sharedEntranceid 只改变存储方式，不改变生成算法**：按当前应用实际使用的历史版本规则生成并保存，单入口/多入口及版本差异均保持兼容，不在本次统一或重新编号。
12. 开发入口也持久化；原始用户使用 `<appid>-<port>`，其他用户使用 `H(appid + ":" + user)-<port>`。同名入口的 Pod IP 等后端属性更新保留 ID；端口变化按对应新开发入口生成和保存 ID。

## 阶段一：重新核验身份模型与公共 API

- [x] 核验当前用户域名模型中的 Terminus 读取方式、实际 CR schema、localName 的写入来源和稳定性。
- [x] 复用现有统一域名读取接口，不重新引入按用户拼接 zone 的旧逻辑；检查 app-service 的 `pkg/cluster/config.go`、BFL 的 local URL 逻辑及 L4 的 machine_network 逻辑。
- [ ] 各模块的 API 版本目前不一致：app-service 使用 `5bf33ba22b5b` 对应版本，CLI 使用 `d281261cdff4` 对应版本，BFL/L4/system-server 为 v0.0.24。确定统一兼容的发布版本。
- [x] 检查 API 的 Entrance、ApplicationSpec、EffectiveEntrances 与当前 ID helper，为普通入口、sharedEntrances 和持久化的用户新增入口统一设计 ID 字段。
- [x] 若 API 的 Terminus 类型尚无 LocalName，按现有身份模型补齐类型；不要重复定义与当前模型冲突的字段或读取路径。
- [x] 实现普通入口的新 ID 分配、开发入口的用户前缀 + 端口分配、共享入口的原算法分配、持久化读取、按入口名合并及冲突检查接口；生成函数只用于创建或升级回填，不用于日常重算已有 ID。
- [x] 明确解析层次：持久化基础 ID、YAML 默认别名、Settings 自定义域名分开处理，保留已有别名覆盖行为。
- [x] 更新 Application CRD schema 和必要生成代码；新标量字段的 deepcopy 应核实是否已由整体复制覆盖。
- [x] 测试序列化往返、原始用户/其他用户、单/多入口、重排、新增、重复名称、缺失身份及 ID 冲突。

交付：所有入口的持久化类型、CRD、公共接口及通过的相关测试。sharedEntrance 的不同历史版本生成算法保持不变。

## 阶段二：Application 生命周期中的持久化

主要位置：`framework/app-service/controllers/application_controller.go`。

- [x] 升级后新安装应用在创建 CR 时，为所有普通入口按新规则生成并写入 ID，为所有共享入口按原规则生成并写入 ID；开发入口由 proxylistener 按已确认的用户/端口规则创建并保存 ID。忽略 chart 提供的 ID，避免覆盖平台分配。
- [x] 更新普通入口和共享入口列表前按名称合并旧 ID，保留新 chart 的服务、端口和显示属性；用户新增入口的持久化容器同样保留已有 ID。
- [x] 覆盖入口删除、重排、单入口变多入口、重复 reconcile；保持已有 ID 稳定。
- [x] shared 应用依据安装者生成一次 ID，用户覆盖层不能改变平台分配的基础 ID。
- [x] 对新 ID 检查其他应用默认 ID、自定义路由及同应用其他入口的占用；不把本次改造扩大为 YAML 默认别名冲突治理。
- [x] proxylistener 重建 addedEntrances 时按名称保留已保存 ID，创建新开发入口时直接写入正确 ID；主域名和端口别名使用同一持久化前缀。
- [x] 将存量回填实现为升级专用任务，与正常安装/更新逻辑分开；控制器和服务启动过程不触发全量扫描回填。
- [x] 正常安装和新增入口只调用升级后的生成接口，不通过长期运行时开关回退旧分配；迁移完成标记只用于跳过已完成升级任务。
- [x] 升级后发现存量入口缺失 ID 时明确报错并交由升级恢复流程处理，不在读取、启动或 reconcile 中悄悄补算。

验收：普通入口、共享入口及用户新增入口的持久化对象中均可查看 ID；chart 更新和控制器重启不会丢失或重新分配，启动/reconcile 不执行升级回填。

## 阶段三：全链路 URL 与路由接入

| 模块 | 需要核验和修改的位置 | 目标 |
| --- | --- | --- |
| app-service URL | `pkg/appcfg/helpers.go` | 使用持久化 ID 和当前统一域名模型 |
| 网关 | `pkg/gateway/srr.go` | 主 Host 和备用 Host 均不回退到其他用户会冲突的旧 ID |
| BFL | `pkg/apis/base.go`、`pkg/app_service/v1/internal.go`、`tools.go`、`local_urls.go` | 保留 CR 中的 ID，入口 URL 与顶层启动 URL 一致，兼容本地域名模式 |
| Helm | `framework/app-service/pkg/appinstaller/helm_utils.go` | 新安装和升级时 `.Values.domain` 与实际入口 ID 一致 |
| OIDC | Application 控制器及相关注册路径 | 回调地址跟随实际入口 ID，不继续拼 appid + 索引 |
| Settings | `pkg/apiserver/handler_settings.go` | 域名变更后的 Helm values 和默认回退使用相同入口规则 |
| L4 provider/translator | `framework/l4-bfl-proxy/internal/provider/`、`internal/translator/` | 路由、公开入口白名单和来源限制使用相同 ID；开发入口主路由及端口访问路径读取已保存的用户前缀 + 端口 ID |
| 开发入口生产者 | 独立 proxylistener 仓库的 `internal/kube/listener.go` | 新入口分配并保存 ID，后端刷新保留 ID，不继续为其他用户生成旧的无用户维度别名 |
| Provider registry | `framework/system-server/pkg/providerregistry/v2alpha1/alias.go` | 根据持久化 ID / 全局唯一别名识别安装者，不能再假定域名第二段就是用户名 |

- [x] 全仓搜索普通/共享入口 ID 重算、appid + 索引拼接以及前端自行拼 URL 的路径；所有消费端读取持久化结果，sharedEntrance 的读取路径也不再每次重新计算。
- [x] 验证旧入口、本地域名、临时用户和 sharedEntrance 的兼容性。
- [ ] 按 zone 检查重复 SNI/filter chain；合并同 zone 的路由，而不是简单丢弃其他用户的 chain。
- [ ] 合并后保留每个路由的认证服务、认证用户和来源限制，防止复用某一个用户的全局认证配置。
- [x] 明确共用系统域名的归属，缺少原始用户信息时拒绝歧义配置。

## 阶段四：Settings 全局唯一性

- [x] 扫描 Spec.Settings 和 Spec.UserSettings，按实际生效/覆盖语义去重；坏配置应报错，不能被忽略后允许重复占用。
- [x] 同时保留系统前缀、持久化默认 ID 与 YAML 别名的占用，统一大小写及前后空白处理。
- [x] 用 CR 标识、用户、入口识别自身，允许原值重保存；不能只用 Spec.Name 区分安装实例。
- [x] 防止通用 Settings 更新接口绕过专用域名接口的校验。
- [x] 通过集群级协调机制串行化检查后写入；可使用可续租 Lease，并在失去租约时取消写入上下文。
- [x] Settings、控制器分配 ID、迁移写入使用同一个协调范围；明确直接写 CR 和旧版本服务不受该机制保护。
- [x] 使用 resourceVersion 防止覆盖并发修改；覆盖设置、修改、清空、删除、卸载和失败后的恢复。
- [x] 保留第三方域名原有的全局唯一性和证书验证。

## 阶段五：仅升级执行一次的存量持久化任务

**回填代码仅属于系统升级流程**。在 `cli/pkg/upgrade/` 下新增 breaking change 迁移任务，修改存量应用的 entranceid 并持久化所有入口 ID，复用现有 breaking change 升级组织方式。目标版本号暂不确定，先完成任务、完成标记和测试，后续版本发布时绑定具体升级入口；不因为缺少版本号暂停实现，也不虚构版本号。该逻辑不作为日常运行时回填机制。

1. 预检统一域名模型、真实原始用户、CRD schema 和组件兼容版本；在维护窗口暂停安装、卸载与 chart 升级。
2. 读取持久化的目标迁移版本和完成状态。已成功完成则直接跳过，不扫描重算或重新分配入口 ID。
3. 首次执行读取完整 Application 及用户新增入口清单，保存原配置备份和不可变迁移计划，包含普通/共享/开发入口旧新 ID及 Settings 自定义路由冲突清理列表。
4. 备份文件以 0600 创建且不覆盖已有文件；域名配置中的私钥等数据按备份管理。
5. 普通入口按 Owner 兼容/admin-normal 新规则计算并写入；sharedEntrance 按其原有实际生成规则计算并写入，保持共享 URL 不变；开发入口按原始用户 `<appid>-<port>` / 其他用户 `H(appid + ":" + user)-<port>` 计算并写入。其他用户新增入口也纳入一次性持久化。
6. 基于同一快照一次性计算自定义路由冲突分组，按 Owner 优先/非 Owner 全清规则清理，仅修改三级路由字段。
7. YAML 默认路由冲突由人工在升级前后控制，迁移代码不新增检测、自动修复或阻断升级；只执行已确定的入口 ID 持久化和 Settings 自定义路由清理。
8. 写入前检查集群身份、原始用户、CR UID、入口清单及业务配置未漂移；使用当前 resourceVersion 写入并校验 ID 没有被 CRD 裁剪。
9. 逐项保存进度。失败后从原计划恢复未完成项，跳过已完成项；不能重新计算剩余冲突组，也不能重算已保存 ID。
10. 所有入口及清理项写入并验证完成后，才持久化本次升级任务的全局完成标记。若最后写标记前中断，重试核验已完成结果后补写标记，不重复改号。
11. 升级成功后正常安装/新增入口直接使用新规则生成并保存；日常读取、服务启动、reconcile、应用更新均不调用存量回填任务。
12. 验证路由重建，并输出入口 ID 和被清理配置报告；完整 URL 变化需结合真实 zone 映射生成。

发布顺序：CRD 与身份信息就绪 → 兼容读取组件就绪并暂停相关写入 → 升级任务一次性回填全部入口 → 写完成标记 → 恢复使用新生成规则的正常运行 → 配合身份模型切换共用域名。兼容读取代码如为滚动发布临时存在，升级后不能继续代替持久化或在运行时回填。

回滚演练：停写，先恢复不会冲突的域名条件，按备份及当前 resourceVersion 恢复原配置与兼容组件，并协调回滚对应迁移版本状态。不能在共用域名状态直接恢复旧的重复入口算法；正常运行不保留旧/新算法切换开关。

## 阶段六：验证与验收

| 验证场景 | 通过标准 |
| --- | --- |
| Owner 存量应用 | 默认入口 ID 保留 |
| 两个其他用户安装同一应用 | appid 相同，普通入口 ID 不同，访问各自正确后端 |
| 应用升级、入口重排、控制器重启 | 原有入口 ID 不变且仍在 CR 中 |
| shared 应用 | 普通入口统一使用安装者 ID；共享入口 ID 已持久化且与升级前一致 |
| 所有入口持久化 | 普通、系统、共享和用户新增入口均有已保存 ID，消费端不重算 |
| 一次性升级任务 | 成功后重试或再次升级直接跳过，启动和 reconcile 不调用回填 |
| 升级后新安装/新增入口 | 普通入口按新规则生成并保存；共享入口按原规则生成并保存 |
| 开发入口 | 原始用户为 `<appid>-<port>`，其他用户为 `H(appid + ":" + user)-<port>`，ID 已保存；Pod IP 更新不改号，无冲突旧别名残留 |
| YAML 默认路由 | 现有值和覆盖行为不变，不新增自动治理或升级阻断 |
| 并发设置相同自定义 ID | 当前最多一个用户成功占用 |
| 重复配置迁移 | Owner 优先；仅非 Owner 重复则全部清空；无关字段保留 |
| 迁移中断和恢复 | 原报告可继续执行，不漏清、不覆盖新配置 |
| 展示与消费链路 | 桌面顶层 URL、入口 URL、Helm、OIDC、网关及 Provider 一致 |
| 同 zone 的 TLS 路由 | Envoy 接受配置，无重复 SNI，认证与来源限制不跨用户泄漏 |
| 本地与公网模式 | 沿用当前 user-domain-model 的统一域名行为 |

- [x] API、app-service、BFL、L4、system-server、CLI、proxylistener 的受影响测试通过，记录命令和结果。
- [x] 对完整测试的既有失败先在同一基线复现；历史对话中的失败记录仅作线索，不替代重新核验。
- [ ] 发布 API 并更新各模块 go.mod / go.sum。临时 workspace 只用于开发，不用本地 replace 作为正式交付。
- [ ] 各模块脱离开发 workspace 后独立构建通过，特别检查旧 genproto 主模块和拆分子模块的依赖冲突。
- [ ] 更新实际部署的 RBAC 和 CRD 来源；读 Terminus 的组件具备相应 get/list/watch 权限。
- [ ] 测试集群完成 HTTP 到目标 Pod、shared 应用认证、TLS 轮换、身份域名切换及回滚演练。

## 建议提交顺序

1. 配套 API：字段、公共接口、CRD、测试。
2. terminus：发布依赖接入、所有入口的正常生成与持久化，以及按名称保留 ID。
3. terminus 与配套 proxylistener：BFL、Helm、OIDC、网关、L4、Provider 等消费端迁移，以及开发入口 ID 创建/更新。
4. terminus：全局查重、写入协调及并发测试。
5. terminus：`cli/pkg/upgrade/` 下的 breaking change 任务、备份报告、断点恢复及一次性完成标记；版本号待发布时绑定。
6. terminus：独立构建、集群联调、发布与回滚说明。

完成标准：所有入口 ID 均持久化；升级回填成功后只执行一次且可断点恢复；升级后普通入口采用新生成规则，sharedEntranceid 原算法不变；生命周期稳定、用户隔离正确，且 API 发布依赖与实际部署验证全部完成。

## 实施后保留的发布事项

- API 发布与各模块正式版本引用：本地验证由 `build/entrance-ids.sh` 创建临时 workspace，不提交本地 replace。
- 为目标发布版本追加 `EntranceIDsBreakingChangeTasks()`；安装 CRD 后、恢复写入前运行。任务没有接入启动/reconcile，也没有虚构版本注册。
- 应用入口 CRD、app-service RBAC 与 proxylistener 的角色/绑定需要在发布 chart 中生效。proxylistener 提供 `config/rbac/entrance-ids.yaml`，具体 service account 绑定由安装 chart 完成。
- 真实集群的 TLS/SNI、访问目标 Pod、共享认证与回滚演练仍待发布环境验收；现有机器级 listener 合并实现保持复用，本地 translator 用户上下文测试已通过。
- API 完整测试有两项已在原始 helper 源码复现的既有失败，详见实施记录；本次新增 API 测试通过。CLI 完整 upgrade 测试在 `LC_ALL=C` 下通过。
