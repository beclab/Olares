# monitoring-server 拆除执行计划

## 1. 目标与基线

- 目标：将 `monitoring-server` (`beclab/ks-apiserver-proxy`) 对 Desktop、Settings、Dashboard、ControlHub 和 `olares-cli` 提供的能力迁移到每用户 BFL，完成入口切换后删除旧 workload、Service 和 provider RBAC。
- 后端、CLI 和 Helm 仓库：`/Users/hys/code/beclab/terminus`，分支 `refactor/monitoring-server`，基线 `origin/main` @ `a1b7d236e60ef383593c66b9850a5e4719914eb5` (2026-10-12)。
- system-frontend 权威仓库：`/Users/hys/code/beclab/TermiPass`，分支 `refactor/monitoring-server`，基线 `origin/main` @ `debb4580ebca4c69044c1b213debf40eb92f6a7b` (2026-10-10)；`terminus/apps/docker/system-frontend` 中的随仓副本同步保留相同的路由切换，避免后续从 terminus 构建时回退到 monitoring-server。
- 路径兼容目标：对外仍保持 `/capi/*`、`/kapis/*`、`/api/v1/*` 和 `/apis/*`，不要求前端或 CLI 改 API URL。
- 安全底线：在普通用户的请求和响应过滤与 `owner`/`admin` 角色映射完成且验证之前，不得把入口直接指向 `ks-apiserver`。

## 2. 仓库边界与分支策略

`beclab/system-frontend` 的权威源码和发布流程已确认在 `TermiPass`：

- nginx 源码位于 `docker/system-frontend/nginx/`，镜像入口是 `docker/system-frontend/dockerfile`。
- `.github/workflows/update-system-frontend-jobs.yml` 负责自动构建发布，`.github/workflows/build-system-frontend-manual.yml` 可指定 tag 手动构建 `beclab/system-frontend`。
- `terminus/apps/docker/system-frontend` 是 2025-12 导入的随仓副本，本仓没有对应镜像发布 workflow；它不作为当前候选镜像的发布源，但本任务同步修改其 nginx 路由和 Dockerfile 镜像源，保证两份构建入口行为一致。

两个仓库是独立 Git 历史，因此都使用同名分支 `refactor/monitoring-server`，但分别从各自的 `origin/main` 创建：

1. `terminus/refactor/monitoring-server`：BFL、CLI、Helm 资源、BFL 和 system-frontend image tag。
2. `TermiPass/refactor/monitoring-server`：权威 nginx 入口配置和 system-frontend 镜像构建。

TermiPass 原本的本地 `main` 比 `origin/main` 跟踪引用落后 941 个提交；已避免在该旧基线上修改，并从最新 `origin/main` 创建了 `refactor/monitoring-server`。

### 当前执行状态（2026-10-14）

- terminus Phase 1～3 已完成并提交：共享身份校验、5 个 `/capi` 兼容接口、KubeSphere HTTP/WebSocket 安全代理均已落在 `refactor/monitoring-server`。
- TermiPass Phase 4 源码切流已完成并提交；terminus 中的 system-frontend 随仓副本也已同步路由切换。全树消费者扫描确认前端源码没有 `/current_user` 调用后，两个 system-frontend 配置中的 `/current_user` location 及其专用 `SettingsServer -> monitoring:28080` upstream 已删除；`Middleware -> middleware:28080` 保持原样。
- Phase 5 的 CLI 所有权文案已更新，远端定向 Go 测试通过。
- terminus 已在远端完成新增/修改包的 test、race、vet 和 build；全量测试只有本文记录的 certmanager 外部 401 基线例外。
- TermiPass 已在远端顺序完成 icon subset、Desktop、Settings、Dashboard 和 ControlHub 构建；修改后的三份 nginx 配置已用节点缓存的 `nginx:stable-alpine3.17-slim` 镜像执行真实 `nginx -t` 并通过。全仓 lint 仍有 83 个与本次 nginx-only diff 无关的既有错误。
- 197 已用 nerdctl 从完整 `Dockerfile.api` 构建并运行 `beclab/bfl:monitoring-server-removal-937c2b419`（manifest `sha256:687bca80faf6c3e19783f391c5df88a4c930082adeab0710a01f7e1aaac7584c`），BFL 临时 hostPath 二进制挂载已移除。Dockerfile 的构建与运行基础镜像均改走 `mirrors.olares.cn`，Go 模块代理为 `https://goproxy.cn,direct`。真实 `/capi/app/detail` 响应已确认只保留 `user`、`runtime`、`systemNamespaces`，不再包含 `clusterRole`、`localeManifest`、`gitopsEngine`；返回前会以结构化 `appDetail response` 日志记录用户名和完整响应对象，197 已触发请求并确认日志输出。`/capi/namespaces/group` 已改由 client-go 读取 Namespace/User，使用无 token 的 `X-BFL-USER: olaresid` 请求验证通过；展平分组后的 16 个 Namespace 对象与切换前 KubeSphere API 基线逐字段一致，规范化 JSON 的 SHA-256 均为 `45235dab8e9daf5bd78e92002236a16381a768a073fa497fb859f4d851b6bc8f`。
- 197 已按正式 workflow 的应用清单串行构建十个前端产物；删除 `/current_user` 后重新用 nerdctl 构建并运行 `beclab/system-frontend:monitoring-server-removal-d40efeef9`（manifest `sha256:b51b65201fa8ef81902b4a37c733065e645d8917367a670ec39dff55fc317b56`）。system-frontend 的两个临时 hostPath 挂载已移除；Dockerfile 基础镜像改为 `mirrors.olares.cn/library/nginx:stable-alpine`。
- terminus 随仓 system-frontend 副本已同步到 197；三份 nginx 文件与本地 SHA-256 一致，并在运行中的同命名空间 Nginx 容器内通过独立临时配置执行 `nginx -t`。当前线上候选镜像仍由 TermiPass 权威源构建，无需因等价副本补丁重复替换。
- 197 测试集群已完成镜像化流量切换。运行中镜像内十个应用的 `index.html`、实际 nginx 配置、`nginx -t` 与四个无 token 鉴权探针均验证通过并确认请求进入新 BFL。更新后的 `nginx -T` 已确认不含 `/current_user`、`SettingsServer` 或 `monitoring:28080`；请求 `/current_user` 会落入 Control Hub 的 SPA fallback 并返回 `text/html`，不再代理到 monitoring。`/middleware` 仍保持原 upstream。
- `/Users/hys/code/beclab` 工作区扫描未发现 `/current_user` 调用方，但在 `oac`、`app-service`、`bfl`、`l4-bfl-proxy`、`tapr`、`system-apps` 等独立仓库中仍发现旧 Nginx 配置副本；它们不属于本次两个指定分支、未进入当前 197 镜像，需由各仓库后续独立清理，不能据此误判为仍有消费者。
- 两个候选镜像目前只存在于 197 的 containerd，尚未推送远程 registry，也未更新 Helm chart tag。必须先完成镜像发布和声明式 tag 更新，才能把当前验证结果视为可跨节点重建的正式候选版本，并开始稳定观察窗口和 Phase 6 删除。

## 3. 交付顺序

### Phase 0：现场盘点与兼容基线

交付物：可重放的旧服务契约样例、路由清单和安全清单。

- 在测试集群导出 `monitoring-server-deployment`、`monitoring-server` Service、每用户 `monitoring` ExternalName Service、5 组 ClusterRole/ClusterRoleBinding 和 BFL ServiceAccount 的实际 YAML。
- 分别用 `owner`、`admin`、`normal` 账号录制 5 个 `/capi` 接口的状态码、header 和 JSON 样例。监控接口同时覆盖 instant query、range query、分母为 0、时间戳错位和 kube-state-metrics 多 `instance` 序列。
- 录制代理路由的典型响应：namespace、pod、PVC、networkpolicy、cronjob、job、monitoring namespaces GET/POST、user metrics、shared app namespace。
- 录制 WebSocket 场景：`kapis/terminal`、`api/v1/watch`、`apis/apps/v1/watch`、pod exec，包括正常关闭、token 过期和上游中断。
- 用 `rg` 重新生成消费者清单。确认前端源码没有 `/current_user` 调用后直接删除 system-frontend 中的该入口及专用 monitoring upstream；不在 BFL 增加兼容实现。

验收：契约 fixture 能被 Go 测试读取；已有每条路由的调用者、身份来源、上游和响应改写规则对照表。

### Phase 1：抽出 BFL 可复用的身份与上游客户端

预计落点：

- `framework/bfl/pkg/apiserver/filters.go`
- `framework/bfl/pkg/apiserver/apiserver.go`
- `framework/bfl/pkg/apiserver/runtime/`
- 新增 `framework/bfl/pkg/proxy/` 或等价的内部包

任务：

- 将现有 `authenticate` 的 token/用户解析抽成可被 go-restful WebService 和原生 `http.Handler` 共用的函数，保持 `X-BFL-USER` 与 `X-Authorization` 兼容。
- 为 WebSocket 增加 `auth_token` cookie 取 token 的兼容路径，并在转发前改成 `Authorization: Bearer ...`。不将信任边界上的客户端 `Authorization`、`X-BFL-USER` 或 hop-by-hop headers 原样传给上游。
- 对每用户 BFL 做身份一致性校验：解析出的 username 必须与 `constants.Username` 一致；如现网路由确有代理身份例外，必须以显式 allowlist 和测试表达。
- 角色判定复用 `apis.Base.IsAdminUser`，明确 `owner` 和 `admin` 都是管理员，`normal` 不是。避免使用进程级无过期用户缓存。
- 建立到 `ks-apiserver.kubesphere-system` 的统一 transport，设置 dial/TLS response-header/idle 超时，但不用全局 write timeout 误杀 watch/exec 长连接。

测试：表驱动覆盖缺 token、坏 token、`X-BFL-USER`、cookie token、用户不匹配、owner/admin/normal 六类分支。

### Phase 2：BFL 接管 5 个 `/capi` 接口

预计落点：新增 `framework/bfl/pkg/apis/capi/`，在 `framework/bfl/pkg/apiserver/apiserver.go` 注册 root path 为 `/capi` 的 WebService。不使用会强制加 `/bfl/<module>/<version>` 的 `runtime.NewWebService`。

按以下顺序实现：

1. `GET /capi/app/myapps`：复用 app-process 的现有 handler/service，保持旧接口的非 BFL envelope 响应。当前树内没找到直接消费者，但在拆除前保留兼容。
2. `GET /capi/system/ifs`：从 BFL 直连 olaresd `/system/ifs`，透传 query，将 `TERMINUSD_HOST` 作为显式 BFL 配置注入，并设置超时。
3. `GET /capi/app/detail`：返回 `user`、`runtime`、`systemNamespaces`；`globalrole` 继续把 owner/admin 映射为 `platform-admin`。消费者确认不再需要后，响应中删除 `clusterRole`、`localeManifest`、`gitopsEngine`；`clusterRole` 仍仅作为 BFL 内部过滤 `globalRules` 的依据，不再对外暴露。
4. `GET /capi/monitoring.kubesphere.io/v1alpha3/:type`：复用 monitor 模块的 KubeSphere client，移植 metric rename、split-series collapse 和 `*_utilisation` 计算。range query 必须按时间戳对齐，缺失/零分母返回 0，不按数组下标相除。
5. `GET /capi/namespaces/group`：通过 client-go 直接读取 Namespace 和 User CR，不再依赖用户 token 调 KubeSphere HTTP API。Namespace 保留原 `kubesphere.io/workspace!=kubesphere.io/devopsproject` selector，并在 BFL 内按创建时间降序、同时间名称降序复刻 `sortBy=createTime`；用 owner-role 找主账号，按 `-<owner>` 后缀分组，识别 label 或 `-shared` 的 shared namespace。只有管理员读取完整 User 列表并可看 Shared/System 组；普通用户不读取其他 User，只返回自己的 namespace。

测试：

- 每个 handler 都用 Phase 0 的 fixture 做 golden contract test，比较状态码和 JSON 语义，不依赖字段顺序。
- `app/detail`、`monitoring/:type`、`namespaces/group` 必须分别覆盖 owner、admin、normal。
- 上游超时、4xx、5xx、坏 JSON 和空列表都有确定响应，不 panic。

### Phase 3：BFL 接管 KubeSphere HTTP 代理和 WebSocket

预计落点：`framework/bfl/pkg/proxy/`、`framework/bfl/pkg/apiserver/apiserver.go`。

任务：

- 以原生 `http.Handler`/`httputil.ReverseProxy` 挂载 `/kapis/`、`/api/v1/` 和 `/apis/`，明确拒绝不在允许集内的根路径。原生 handler 必须显式调用 Phase 1 的鉴权，不能假设 go-restful container filter 会生效。
- 沿用客户 token 访问 KubeSphere，不改成 BFL ServiceAccount 的集群管理权限。
- 移植普通用户的请求改写：monitoring namespaces GET/POST 的 namespace 参数注入；POST JSON 最大 10 MiB，坏 JSON 返回 400，超限返回 413。
- 移植普通用户的响应过滤：namespace、pod、PVC、networkpolicy、cronjob、job、monitoring namespace result 和 user metric split-series collapse。所有 ownership 判定使用 `-<owner>` 后缀，shared app 同时支持 label 与 `-shared`。
- 只缓冲需要改写的有限 JSON 响应；正确处理 gzip、`Content-Length`、`Transfer-Encoding`、上游错误和客户取消。对大响应设上限，超限时失败关闭，不得回退成将未过滤数据返给普通用户。
- 使用同一 proxy transport 支持 WebSocket upgrade、双向流、half-close 和客户取消。限制每用户同时 watch/exec 数，超限返回 429；连接建立超时与 idle/ping 策略通过配置管理。
- 如果需要保留原 provider RBAC 的“按前端域名限制路径”能力，根据经过可信代理重写的 Host/route identity 实现显式策略；不直接信任外部传入的 `X-Forwarded-Host`。

安全测试：

- normal 不能读到其他用户、system 或未允许 shared namespace；owner/admin 保持旧行为。
- gzip/非 gzip、chunked/已知长度、坏 JSON、大 body、上游中断都不能绕过过滤。
- WebSocket 不带 token、用户不匹配、token 过期必须在 upgrade 前失败。
- 使用 `go test -race ./...` 验证连接计数器、取消和缓存无竞争。

### Phase 4：在 TermiPass 切换入口，但保留旧服务作回滚点

前置：Phase 2/3 的契约、安全和 WebSocket 测试全部通过；TermiPass 已从最新 `origin/main` 创建同名 `refactor/monitoring-server` 分支。

任务：

- `TermiPass/docker/system-frontend/nginx/desktop.conf`：删除 `DesktopMonitoringServer`，仅将 `/kapis` 改为 `proxy_pass http://bfl`，不动指向 user-service 的 `/api`。
- `TermiPass/docker/system-frontend/nginx/settings.conf`：删除 `SettingsServer_Monitoring`，仅将 `/kapis` 改为 `proxy_pass http://bfl`，不动现有 `/api`、`/apis/backup`、files 等路由。
- `TermiPass/docker/system-frontend/nginx/dashboard-control-hub.conf`：两个 server block 内的 `/kapis`、`/capi`、KubeSphere `/api` 以及 terminal/watch/exec regex location 均改为 `proxy_pass http://bfl`。保留且回归 `/api/gpu/list`、`/api/profile/init`、`/user-service`、`/middleware` 等更精确的本地路由；删除无消费者的 `/current_user` location 及仅供它使用的 `SettingsServer -> monitoring:28080` upstream。
- 将以上三份 nginx 路由修改同步到 `terminus/apps/docker/system-frontend/nginx/`；同时将该副本 Dockerfile 的 nginx 基础镜像改为 `mirrors.olares.cn/library/nginx:stable-alpine`。该副本不是当前发布源，但必须通过相同的配置检查，防止备用构建路径重新引入 monitoring-server 流量。
- WebSocket regex 必须覆盖当前 TermiPass `origin/main` 中的四类路径：`kapis/terminal`、`api/v1/watch`、`apis/apps/v1/watch`、`api/v1/namespaces/.../pods/.../exec`；保留 HTTP/1.1、Upgrade、Connection 和 Host headers。
- `/current_user` 不迁移到 BFL；消费者扫描无调用后直接删除。验收时通过 `nginx -T` 确认该 location 和 monitoring upstream 不存在；由于 Control Hub 有 SPA fallback，直接请求该路径预期返回 `index.html`，而不是以是否返回 404 判断。
- `TermiPass/packages/app/config/{dashboard,control-hub}.js`、`Desktop/proxy.js` 和 `Settings/proxy.js` 的开发代理已经指向对应公开入口域名，路径不变时无需修改。
- 不修改已下架 Studio 的 `TermiPass/docker/studio/studio.conf` 和旧的 `TermiPass/docker/settings/nginx.conf`；它们不参与本次 system-frontend 切流，在最终全树扫描中作为已知历史引用记录。
- 在 TermiPass 使用 `.github/workflows/build-system-frontend-manual.yml` 生成一个明确的候选 tag，验证后再走 `.github/workflows/update-system-frontend-jobs.yml` 的正式发布节奏。
- 先发布新 BFL 镜像并确认 Ready，再发布 TermiPass system-frontend 镜像，最后在 `terminus/apps/.olares/config/user/helm-charts/system-apps/templates/olares-app.yaml` 更新 system-frontend image tag 完成切流。
- 本阶段不删 `monitoring-server`、provider RBAC 或 ExternalName Service，以便通过回滚 system-frontend 镜像快速恢复。

验收：用 owner/admin/normal 账号跑 Desktop、Settings、Dashboard、ControlHub 回归；通过 access log/metric 确认新请求落到 BFL，旧 monitoring-server 在观察窗口内无新流量。

### Phase 5：CLI 和文档兼容复核

- 保持 `cli/pkg/dashboard`、`cli/cmd/ctl/cluster/application`、`cli/pkg/clusterctx` 对原路径和响应的解包假设，运行已有 CLI 测试。
- 更新 `cli/pkg/olares/id.go`、`cli/skills/olares-dashboard/` 和 `cli/skills/olares-cluster/` 中将 BFF 描述为 ks-apiserver-proxy/monitoring-server 的文字，不改公共命令行行为。
- 运行 CLI 的 dashboard overview、cluster application list、cluster context 和 pod exec 集成测试。

### Phase 6：删除 monitoring-server 与过渡资源

前置：Phase 4 切流后完成至少一个约定的稳定观察窗口，无旧服务流量，无多租户数据越权，无 WebSocket 容量告警。

任务：

- 删除 `framework/monitor/.olares/config/cluster/deploy/ks_server_deploy.yaml` 及对应 README/模块装配；如 `framework/monitor` 无其他责任，删除整个目录。
- 从 `cli/pkg/systemcomponents/registry.go` 移除 `monitoring-server-deployment` 就绪检查。
- 删除 `apps/.olares/config/user/helm-charts/system-apps/templates/monitoring-provider.yaml` 中 5 组 ClusterRole、5 组 ClusterRoleBinding 和 `monitoring` ExternalName Service；若文件不再有内容，删除文件。
- 用 `helm template` 与升级测试确认旧的 cluster-scoped RBAC 会被正确删除，不留孤儿资源。
- 全树 `rg 'monitoring-server|monitoring:28080|provider-service-ref: monitoring-server'` 除历史/迁移文档和明确标注的非 system-frontend 旧配置外，不得有其他运行时引用。
- 单独的运营任务：停止 `beclab/ks-apiserver-proxy` 新镜像发布，但不立即删除已发布镜像，留足回滚窗口。

## 4. 测试与验收命令

所有代码测试和集成验证统一在 `olares@192.168.31.197` 上执行，不以本机测试结果作为验收证据。每轮验证前，从 `/Users/hys/code/beclab` 执行以下同步：

```bash
rsync -zarv -e "ssh -p 22 " \
  --exclude=.svn \
  --exclude=.cvs \
  --exclude=.idea \
  --exclude=.DS_Store \
  --exclude=.git \
  --exclude=.hg \
  --exclude='*.hprof' \
  --exclude='*.pyc' \
  terminus/ olares@192.168.31.197:/code/terminus/
```

同步规则：

- 本地 `terminus/refactor/monitoring-server` 工作树是测试代码的唯一源；每次同步前记录本地 commit 和 `git status --short`。
- 使用上述 exclude 保留远端 `/code/terminus/.git`，不同步本地 Git metadata。不加 `--delete`，避免删除远端文件。
- rsync 完成后在远端记录同步时间和本地 commit，再通过 SSH 执行命令。远端 `/code/terminus` 自身的 Git 分支/索引不作为代码版本判据。
- 测试产物、缓存和临时文件只留在远端，不反向 rsync 回本地工作树。

实现期间每次同步后固定执行：

```bash
ssh -p 22 olares@192.168.31.197 '
  cd /code/terminus/framework/bfl &&
  KUBECONFIG=/etc/rancher/k3s/k3s.yaml go test ./... &&
  KUBECONFIG=/etc/rancher/k3s/k3s.yaml go test -race ./... &&
  KUBECONFIG=/etc/rancher/k3s/k3s.yaml go vet ./... &&
  KUBECONFIG=/etc/rancher/k3s/k3s.yaml go build ./cmd/apiserver
'
```

当前远端基线的 `pkg/utils/certmanager/TestCertManager` 会请求真实证书 API；未提供 JWS 时固定返回 401。该项作为既有环境性例外单独记录，不可掩盖其他测试失败；本任务新增/修改包的定向测试、race、vet 和构建仍须全部通过。

```bash
ssh -p 22 olares@192.168.31.197 '
  cd /code/terminus/cli &&
  go test ./pkg/dashboard/... ./pkg/clusterctx/... ./pkg/clusterclient/... ./pkg/cmdutil/... ./cmd/ctl/cluster/...
'
```

根据实际 chart 依赖补充并固定远端 `helm template`/upgrade 测试命令。TermiPass 的权威源不在 `terminus/` rsync 范围内；Phase 4 开始前，使用同等 exclude 规则将 `/Users/hys/code/beclab/TermiPass/` 同步到远端 `/code/TermiPass/`，然后在远端执行：

```bash
ssh -p 22 olares@192.168.31.197 '
  cd /code/TermiPass &&
  npm run lint
'

ssh -p 22 olares@192.168.31.197 '
  cd /code/TermiPass &&
  (cd packages/app && npm run icons:subset) &&
  npm run desktop:build &&
  npm run settings:build &&
  npm run dashboard:build &&
  npm run control-hub:build
'
```

lint 单独运行，避免仓库现存的非本任务 lint 基线阻断 nginx-only 变更的构建验证；必须记录错误总数并确认本次 diff 未新增对应源文件错误。四个前端构建必须串行执行，避免共享 Quasar 构建目录互相覆盖。

system-frontend 多架构镜像继续由 TermiPass GitHub Actions workflow 构建。在 `192.168.31.197` 拉取候选镜像并启动容器，对实际装载的 Desktop、Settings、Dashboard/ControlHub nginx 配置执行 `nginx -t`，再进行集群切流。

集群级验收矩阵：

| 类别 | owner | admin | normal |
| --- | --- | --- | --- |
| 5 个 `/capi` 状态码与 JSON 契约 | 必测 | 必测 | 必测 |
| namespace/pod/workload 可见性 | 全局 | 全局 | 仅本用户允许范围 |
| shared/system namespace 规则 | 必测 | 必测 | 必测隔离 |
| instant/range monitoring | 必测 | 必测 | 必测 |
| terminal/watch/exec WebSocket | 必测 | 必测 | 必测授权边界 |
| CLI dashboard/cluster | 必测 | 必测 | 必测 |

`namespaces/group` 的 client-go 切换验收必须在部署前保存 KubeSphere namespace API 基线，部署后仅携带 `X-BFL-USER` 请求新接口，将各分组的 `.data[]` 展平并按 `.metadata.name` 排序后逐字段比较。197 实测 16 个 Namespace 的 `diff` 无输出，切换前后 SHA-256 均为 `45235dab8e9daf5bd78e92002236a16381a768a073fa497fb859f4d851b6bc8f`；同时确认 owner 分组顺序为 `user-system-olaresid`、`user-space-olaresid`，System 组保持创建时间降序及同时间名称降序。

最终完成标准：

- 前端和 CLI 无需修改 API URL，主要响应与旧服务契约一致。
- normal 用户无法通过查询参数、POST body、gzip/坏 JSON 或 WebSocket 越权。
- owner 和 admin 两种 owner-role 都被识别为 `platform-admin`。
- BFL 的 watch/exec 长连接不导致该用户的普通 API 失去响应，容量与错误有可观测指标。
- Helm 升级后集群中无 monitoring-server Deployment/Service、无 monitoring provider RBAC、无 `monitoring:28080` ExternalName Service。
- `olares-cli` 就绪检查不再等待 monitoring-server。

## 5. 提交拆分与回滚

建议保持以下可独立评审的提交边界。两个仓库都使用 `refactor/monitoring-server` 分支：

1. terminus：`test(bfl): capture monitoring-server compatibility contracts`
2. terminus：`refactor(bfl): add shared request identity and upstream transport`
3. terminus：`feat(bfl): serve monitoring compatibility capi endpoints`
4. terminus：`feat(bfl): proxy kubesphere http and websocket traffic safely`
5. TermiPass：`refactor(system-frontend): route kubesphere traffic to bfl`
6. terminus：`chore(system-apps): bump system-frontend for bfl cutover`
7. terminus：`docs(cli): update monitoring bff ownership`
8. terminus：`chore(helm): remove monitoring-server resources`

回滚顺序：

1. 在 Phase 6 前，只需回滚 system-frontend 镜像/tag，路由即可回到仍运行的 monitoring-server。
2. Phase 6 后如需回滚，同时恢复 monitoring-server Helm 资源、provider RBAC/ExternalName Service 和旧 system-frontend tag。
3. BFL 新增的兼容路由可在回滚期间保留，因为它们对旧入口无影响；等事故结束后再决定是否回退 BFL 镜像。
