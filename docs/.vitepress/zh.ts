import { defineConfig, type DefaultTheme } from "vitepress";
import { oneSidebar } from './one.zh.ts';
import { useCaseSidebar } from './usecase.zh.ts';
import { developerSidebar } from './developer.zh.ts';
const side = {
  "/zh/manual/": [
    {
      text: "概览",
      link: "/zh/manual/overview",
      items: [
        {
          text: "更新说明",
          collapsed: true,
          items: [
            {
              text: "Olares 1.12.6",
              link: "/zh/manual/update-guides/1.12.6",
            },
          ],
        },
        {
          text: "文档更新动态",
          link: "/zh/manual/release-notes",
        },
        {
          text: "常见问题",
          link: "/zh/manual/help/faqs",
          collapsed: true,
          items: [
            {
              text: "关于 Olares",
              link: "/zh/manual/help/olares",
            },
            {
              text: "安装与激活",
              link: "/zh/manual/help/installation",
            },
            {
              text: "使用 Olares",
              link: "/zh/manual/help/usage",
            },
          ],
        },
        {
          text: "获取支持",
          link: "/zh/manual/help/request-technical-support",
        },
      ],
    },
    {
      text: "快速开始",
      collapsed: false,
      link: "/zh/manual/get-started/",
      items: [
        {
          text: "下载 LarePass",
          link: "/zh/manual/larepass/",
        },
        {
          text: "创建 Olares ID",
          link: "/zh/manual/get-started/create-olares-id",
        },
        {
          text: "安装 Olares",
          link: "/zh/manual/get-started/install-olares",
          collapsed: true,
          items: [
            {
              text: "Linux",
              collapsed: true,
              items: [
                {
                  text: "使用 ISO 镜像（推荐）",
                  link: "/zh/manual/get-started/install-linux-iso",
                },
                {
                  text: "使用脚本",
                  link: "/zh/manual/get-started/install-linux-script",
                },
                {
                  text: "使用 Docker Compose",
                  link: "/zh/manual/get-started/install-linux-docker",
                },
              ],
            },
            {
              text: "DGX Spark",
              collapsed: true,
              items: [
                {
                  text: "使用脚本（推荐）",
                  link: "/zh/manual/get-started/install-spark-script",
                },
                {
                  text: "使用 ISO 镜像",
                  link: "/zh/manual/get-started/install-spark-iso",
                },
              ],
            },
            /* {
              text: "macOS",
              collapsed: true,
              items: [
                {
                  text: "使用脚本",
                  link: "/zh/manual/get-started/install-mac-script",
                },
                {
                  text: "使用 Docker 镜像",
                  link: "/zh/manual/get-started/install-mac-docker",
                },
              ],
            }, */
            /* {
              text: "Windows (WSL 2)",
              collapsed: true,
              items: [
                {
                  text: "使用脚本",
                  link: "/zh/manual/get-started/install-windows-script",
                },
              ],
            }, */
            /* {
              text: "PVE",
              collapsed: true,
              items: [
                {
                  text: "使用脚本",
                  link: "/zh/manual/get-started/install-pve-script",
                },
                {
                  text: "使用 ISO 镜像",
                  link: "/zh/manual/get-started/install-pve-iso",
                },
                { text: "LXC", link: "/zh/manual/get-started/install-lxc" },
              ],
            }, */
            /* {
              text: "树莓派",
              link: "/zh/manual/get-started/install-raspberry-pi",
            }, */
          ],
        },
        {
          text: "以成员身份加入 Olares",
          link: "/zh/manual/get-started/join-olares",
        },
        {
          text: "了解桌面",
          link: "/zh/manual/olares/desktop",
        },
        {
          text: "探索",
          link: "/zh/manual/get-started/next-steps",
        },
      ],
    },
    {
      text: "账户与访问",
      collapsed: false,
      items: [
        {
          text: "Olares ID",
          collapsed: true,
          items: [
            {
              text: "在 LarePass 中管理 Olares ID",
              link: "/zh/manual/larepass/manage-accounts",
            },
            {
              text: "修改密码并管理已登录设备",
              link: "/zh/manual/password-and-devices",
            },
            {
              text: "设置自定义域名 Olares ID",
              link: "/zh/manual/best-practices/set-custom-domain",
            },
            {
              text: "管理 Olares Space 账户与计费",
              link: "/zh/manual/space/manage-accounts",
            },
            {
              text: "备份助记词",
              link: "/zh/manual/larepass/back-up-mnemonics",
            },
          ],
        },
        {
          text: "团队",
          collapsed: true,
          items: [
            {
              text: "创建并管理成员",
              link: "/zh/manual/olares/settings/manage-team",
            },
            {
              text: "角色与权限",
              link: "/zh/manual/olares/settings/roles-permissions",
            },
          ],
        },
        {
          text: "访问 Olares",
          collapsed: true,
          items: [
            {
              text: "在本地网络访问",
              link: "/zh/manual/best-practices/local-access",
            },
            {
              text: "通过 LarePass VPN 远程访问",
              link: "/zh/manual/larepass/private-network",
            },
            {
              text: "通过 SSH 连接",
              link: "/zh/manual/access-olares-terminal",
            },
          ],
        },
        {
          text: "密码",
          collapsed: true,
          items: [
            {
              text: "管理 Vault 项目",
              link: "/zh/manual/olares/vault/vault-items",
            },
            {
              text: "安全共享 Vault 项目",
              link: "/zh/manual/olares/vault/share-vault-items",
            },
            {
              text: "自动填充密码",
              link: "/zh/manual/larepass/autofill",
            },
            {
              text: "生成 2FA 验证码",
              link: "/zh/manual/larepass/two-factor-verification",
            },
          ],
        },
        {
          text: "个性化设置",
          collapsed: true,
          items: [
            {
              text: "更改语言与外观",
              link: "/zh/manual/olares/settings/language-appearance",
            },
            {
              text: "创建 Profile",
              link: "/zh/manual/olares/profile",
            },
          ],
        },
      ],
    },
    {
      text: "应用",
      link: "/zh/manual/olares/",
      collapsed: false,
      items: [
        {
          text: "安装、更新和移除应用",
          link: "/zh/manual/olares/market/market",
        },
        {
          text: "手动更新应用镜像",
          link: "/zh/manual/update-app-image",
        },
        {
          text: "克隆应用",
          link: "/zh/manual/olares/market/clone-apps",
        },
        {
          text: "配置应用环境变量",
          link: "/zh/manual/olares/settings/manage-app-env",
        },
        {
          text: "设置系统级环境变量",
          link: "/zh/manual/manage-system-env",
        },
        {
          text: "为应用分配算力资源",
          link: "/zh/manual/olares/settings/gpu-resource",
        },
        {
          text: "更改应用访问地址",
          link: "/zh/manual/olares/settings/custom-app-domain",
        },
        {
          text: "让应用在局域网可用",
          link: "/zh/manual/olares/settings/overlay-gateway",
        },
        {
          text: "控制应用访问权限",
          link: "/zh/manual/olares/settings/manage-entrance",
        },
        {
          text: "连接 AI 应用与模型服务",
          link: "/zh/manual/best-practices/connect-ai-apps",
        },
        {
          text: "迁移旧版共享应用",
          link: "/zh/manual/migrate-shared-apps",
        },
        {
          text: "关于共享应用",
          link: "/zh/manual/olares/market/shared-apps",
        },
      ],
    },
    {
      text: "文件与存储",
      link: "/zh/manual/olares/files/",
      collapsed: false,
      items: [
        {
          text: "文件操作",
          collapsed: true,
          items: [
            {
              text: "上传、编辑和下载文件",
              link: "/zh/manual/olares/files/add-edit-download",
            },
            {
              text: "压缩与解压缩文件",
              link: "/zh/manual/olares/files/compress-extract-files",
            },
            {
              text: "分享文件",
              link: "/zh/manual/olares/files/share-files",
            },
            {
              text: "与电脑同步文件",
              link: "/zh/manual/olares/files/sync-files",
            },
            {
              text: "配置文件搜索",
              link: "/zh/manual/olares/settings/search",
            },
            {
              text: "管理共享 AI 模型文件",
              link: "/zh/manual/olares/files/files-common",
            },
          ],
        },
        {
          text: "连接外部存储",
          collapsed: true,
          items: [
            {
              text: "使用 U 盘",
              link: "/zh/manual/use-usb-drive",
            },
            {
              text: "挂载本地磁盘",
              link: "/zh/manual/mount-local-disk",
            },
            {
              text: "连接 SMB 共享",
              link: "/zh/manual/olares/files/mount-SMB",
            },
            {
              text: "连接 NFS 共享",
              link: "/zh/manual/olares/files/mount-nfs",
            },
            {
              text: "连接云存储",
              link: "/zh/manual/olares/files/mount-cloud-storage",
            },
          ],
        },
      ],
    },
    {
      text: "系统",
      collapsed: false,
      items: [
        {
          text: "查看系统和应用资源使用",
          link: "/zh/manual/olares/resources-usage",
        },
        {
          text: "查看 Olares 状态与 Olares Space 用量",
          link: "/zh/manual/space/manage-olares",
        },
        {
          text: "更新 Olares",
          link: "/zh/manual/olares/settings/update",
        },
        {
          text: "重新激活 Olares",
          link: "/zh/manual/larepass/activate-olares",
        },
        {
          text: "管理 Olares 设备",
          link: "/zh/manual/larepass/manage-olares",
        },
        {
          text: "备份与恢复 Olares",
          link: "/zh/manual/olares/settings/backup",
        },
        {
          text: "集群",
          collapsed: true,
          items: [
            {
              text: "安装多节点 Olares 集群",
              link: "/zh/manual/best-practices/install-olares-multi-node",
            },
          ],
        },
        {
          text: "存储",
          collapsed: true,
          items: [
            {
              text: "释放磁盘空间",
              link: "/zh/manual/free-up-disk-space",
            },
            {
              text: "扩展系统存储",
              link: "/zh/manual/best-practices/expand-storage-in-olares",
            },
          ],
        },
        {
          text: "高级管理",
          link: "/zh/manual/olares/controlhub/",
          collapsed: true,
          items: [
            {
              text: "管理工作负载",
              link: "/zh/manual/olares/controlhub/manage-workload",
            },
            {
              text: "查看和管理容器",
              link: "/zh/manual/olares/controlhub/manage-container",
            },
            {
              text: "管理资源配置",
              link: "/zh/manual/olares/controlhub/manage-resource",
            },
            {
              text: "管理中间件",
              link: "/zh/manual/olares/controlhub/manage-middleware",
            },
            {
              text: "在控制面板终端中执行命令",
              link: "/zh/manual/olares/controlhub/terminal",
            },
            {
              text: "配置 Olares 内主机名解析",
              link: "/zh/manual/olares/settings/set-up-hosts",
            },
            {
              text: "更改反向代理",
              link: "/zh/manual/olares/settings/change-frp",
            },
            {
              text: "配置视频播放",
              link: "/zh/manual/olares/settings/video",
            },
            {
              text: "管理仓库和镜像",
              link: "/zh/manual/olares/settings/developer",
            },
          ],
        },
      ],
    },
    {
      text: "高级设置",
      link: "/zh/manual/best-practices/",
      collapsed: true,
      items: [
        {
          text: "通过命令行安装并激活 Olares",
          link: "/zh/manual/best-practices/activate-olares-using-cli",
        },
        {
          text: "安装特定版本的 NVIDIA 驱动",
          link: "/zh/manual/best-practices/install-specific-cuda-version",
        },
      ],
    },
    {
      text: "帮助与故障排查",
      link: "/zh/manual/help/",
      collapsed: true,
      items: [
        {
          text: "收集诊断信息",
          link: "/zh/manual/collect-diagnostic-information",
        },
        {
          text: "故障排查",
          link: "/zh/manual/help/troubleshooting-guide",
          collapsed: true,
          items: [
            {
              text: "应用市场应用缺失",
              link: "/zh/manual/help/ts-missing-apps",
            },
            {
              text: "无法在应用独占模式移除已暂停应用",
              link: "/zh/manual/help/ts-cs-app-reappears",
            },
            {
              text: "控制面板启停应用后状态不一致",
              link: "/zh/manual/help/ts-inconsistent-app-status",
            },
            {
              text: "应用安装、更新期间或之后失败",
              link: "/zh/manual/help/ts-app-fails-after-update",
            },
            {
              text: "模型或引擎未就绪",
              link: "/zh/manual/help/ts-model-engine-not-ready",
            },
            {
              text: "内存不足或没有释放",
              link: "/zh/manual/help/ts-free-memory",
            },
            {
              text: "GPU 应用安装或恢复后处于暂停状态",
              link: "/zh/manual/help/ts-vram-shortage",
            },
            {
              text: "LarePass 专用网络无法使用",
              link: "/zh/manual/help/ts-larepass-vpn-not-working",
            },
            {
              text: "网络尚未准备好或 olares 连接错误",
              link: "/zh/manual/help/ts-network-not-ready",
            },
            {
              text: "自定义路由 ID 导致应用无法访问",
              link: "/zh/manual/help/ts-custom-route-domain",
            },
            {
              text: "Steam 串流卡顿或延迟",
              link: "/zh/manual/help/ts-steam-stream-lag",
            },
            {
              text: "磁盘空间已满",
              link: "/zh/manual/free-up-disk-space",
            },
            {
              text: "忘记桌面登录密码",
              link: "/zh/manual/help/ts-forget-login-password",
            },
            {
              text: "登录与激活错误信息",
              link: "/zh/manual/login-and-activation-errors",
            },
            {
              text: "LarePass 显示系统错误",
              link: "/zh/manual/help/ts-system-error",
            },
            {
              text: "Olares One 无法开机或没有画面",
              link: "/zh/one/ts-no-power-or-display",
            },
          ],
        },
        {
          text: "已知问题",
          link: "/zh/manual/help/known-issues",
        },
        { text: "术语表", link: "/zh/manual/glossary" },
      ],
    },
  ],
};

export const zh = defineConfig({
  lang: "zh",
  titleTemplate: ":title | Olares 中文文档",
  themeConfig: {
    editLink: {
      pattern: "https://github.com/beclab/Olares/edit/main/docs/:path",
      text: "在 GitHub 上编辑此页面",
    },
    socialLinks: [{ icon: "github", link: "https://github.com/beclab/olares" }],

    nav: [
      { text: "Olares OS", link: "zh/manual/overview" },
      { text: "Olares One", link: "/zh/one/" },
      { text: "应用示例", link: "/zh/use-cases/" },
      { text: "开发者文档", link: "/zh/developer/concepts/" },
    ],

    sidebar: {
      ...side,
      ...oneSidebar,
      ...useCaseSidebar,
      ...developerSidebar,
    },
  },
});
