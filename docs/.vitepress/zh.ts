import { defineConfig, type DefaultTheme } from "vitepress";

const side = {
  "/zh/manual/": [
    {
      text: "Olares 是什么？",
      link: "/zh/manual/overview",
      items: [
        // { text: "比较 Olares 和 NAS", link: "/zh/manual/olares-vs-nas" },
        {
          text: "常见问题",
          //              link: "/zh/manual/help/faqs",
          collapsed: true,
          items: [
            {
              text: "产品",
              link: "/zh/manual/help/olares",
            },
            {
              text: "安装激活",
              link: "/zh/manual/help/installation",
            },
            {
              text: "使用",
              link: "/zh/manual/help/usage",
            },
            //     {
            //       text: "技术支持",
            //      link: "/zh/manual/help/request-technical-support",
            //       },
            //     {
            //       text: "Troubleshooting Guide",
            //       link: "/zh/manual/help/troubleshooting-guide",
            //     },
          ],
        },
      ],
    },
    {
      text: "快速开始",
      collapsed: false,
      link: "/zh/manual/get-started/",
      items: [
        // { text: "Quick start", link: "/zh/manual/get-started/quick-start" },
        {
          text: "创建 Olares ID",
          link: "/zh/manual/get-started/create-olares-id",
        },
        {
          text: "安装激活",
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
            },
            {
              text: "Windows (WSL 2)",
              collapsed: true,
              items: [
                {
                  text: "使用脚本",
                  link: "/zh/manual/get-started/install-windows-script",
                },
                /*{
                  text: "使用 Docker 镜像",
                  link: "/zh/manual/get-started/install-windows-docker",
                },*/
              ],
            },
            {
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
            },
            {
              text: "树莓派",
              link: "/zh/manual/get-started/install-raspberry-pi",
            },
          ],
        },
        {
          text: "备份助记词",
          link: "/zh/manual/larepass/back-up-mnemonics",
        },
        {
          text: "安全访问 Olares",
          link: "zh/manual/get-started/local-access",
        },
        {
          text: "探索",
          link: "/zh/manual/get-started/next-steps",
        },
      ],
    },
    {
      text: "LarePass",
      link: "/zh/manual/larepass/",
      collapsed: true,
      items: [
        {
          text: "管理账户",
          collapsed: true,
          items: [
            { text: "创建账户", link: "/zh/manual/larepass/create-account" },
            { text: "备份助记词", link: "/zh/manual/larepass/back-up-mnemonics" },
            { text: "管理集成", link: "/zh/manual/larepass/integrations" },
          ],
        },
        { text: "使用专用网络", link: "/zh/manual/larepass/private-network" },
        {
          text: "管理设备",
          collapsed: true,
          items: [
            { text: "激活 Olares", link: "/zh/manual/larepass/activate-olares" },
            { text: "管理 Olares", link: "/zh/manual/larepass/manage-olares" },
          ],
        },
        { text: "管理文件", link: "/zh/manual/larepass/manage-files" },
        //    collapsed: true,
        //    items: [
        //    {text: "常用文件操作", link:"/zh/manual/larepass/manage-files"},
        //    {text: "同步与共享", link:"/zh/manual/larepass/sync-share"}
        //  ]
        // },
        {
          text: "管理密码",
          collapsed: true,
          items: [
            { text: "自动填充", link: "/zh/manual/larepass/autofill" },
            { text: "双重验证", link: "/zh/manual/larepass/two-factor-verification" },
          ],
        },
        /*{
          text: "管理内容",
          link: "/zh/manual/larepass/manage-knowledge",
        },*/
      ],
    },
    {
      "text": "Olares 应用",
      "collapsed": true,
      "link": "/zh/manual/olares/",
      "items": [
        { "text": "桌面", "link": "/zh/manual/olares/desktop" },
        {
          "text": "应用市场",
          "collapsed": true,
          "items": [
            {
              "text": "基本操作",
              "link": "/zh/manual/olares/market/market"
            },
            {
              "text": "管理付费应用",
              "link": "/zh/manual/olares/market/purchase-paid-apps"
            },
            {
              "text": "克隆应用",
              "link": "/zh/manual/olares/market/clone-apps"
            },
          ],
        },
        {
          "text": "文件管理器",
          "collapsed": true,
          "link": "/zh/manual/olares/files/",
          "items": [
            {
              "text": "基本文件操作",
              "link": "/zh/manual/olares/files/add-edit-download"
            },
            //{
            //   "text": "同步与共享",
            //  "link": "/zh/manual/larepass/sync-share"
            //  },
            {
              "text": "分享文件",
              "link": "/zh/manual/olares/files/share-files"
            },
            {
              "text": "同步文件至本地",
              "link": "/zh/manual/olares/files/sync-files"
            },
            {
              "text": "挂载 SMB",
              "link": "/zh/manual/olares/files/mount-SMB"
            },
            {
              "text": "挂载云存储",
              "link": "/zh/manual/olares/files/mount-cloud-storage"
            },
          ],
        },
        {
          "text": "Vault",
          "collapsed": true,
          "link": "/zh/manual/olares/vault/",
          "items": [
            {
              "text": "管理 Vault 项目",
              "link": "/zh/manual/olares/vault/vault-items"
            },
            {
              "text": "管理共享 Vault",
              "link": "/zh/manual/olares/vault/share-vault-items"
            },
            {
              "text": "自动填充",
              "link": "/zh/manual/larepass/autofill"
            },
            {
              "text": "双因素验证",
              "link": "/zh/manual/larepass/two-factor-verification"
            }
          ]
        },
        {
          "text": "Wise",
          "collapsed": true,
          "link": "/zh/manual/olares/wise/",
          "items": [
            {
              "text": "基本操作",
              "link": "/zh/manual/olares/wise/basics"
            },
            /*{
              "text": "获取推荐引擎",
              "link": "/zh/manual/olares/wise/recommend"
            },*/
            {
              "text": "管理订阅",
              "link": "/zh/manual/olares/wise/subscribe"
            },
            {
              text: "管理 Cookie",
              link: "/zh/manual/olares/wise/manage-cookies",
            },
            {
              "text": "管理知识",
              "link": "/zh/manual/olares/wise/filter",
              collapsed: true,
              items: [
                {
                  "text": "过滤语法参考",
                  "link": "/zh/manual/olares/wise/filter-syntax-guide"
                },
                {
                  "text": "过滤视图示例",
                  "link": "/zh/manual/olares/wise/filter-examples"
                }
              ],
            },
          ],
        },
        {
          "text": "控制面板",
          "collapsed": true,
          "link": "/zh/manual/olares/controlhub/",
          "items": [
            {
              "text": "管理工作负载",
              "link": "/zh/manual/olares/controlhub/manage-workload"
            },
            {
              "text": "管理容器",
              "link": "/zh/manual/olares/controlhub/manage-container"
            },
            {
              "text": "管理资源配置",
              "link": "/zh/manual/olares/controlhub/manage-resource"
            },
            {
              "text": "管理中间件",
              "link": "/zh/manual/olares/controlhub/manage-middleware"
            },
            {
              text: "访问终端",
              link: "/zh/manual/olares/controlhub/terminal",
            },
          ],
        },
        {
          "text": "设置",
          "collapsed": true,
          "link": "/zh/manual/olares/settings/",
          "items": [
            { "text": "我的 Olares", link: "/zh/manual/olares/settings/my-olares" },
            {
              "text": "管理用户",
              "collapsed": true,
              "items": [
                {
                  "text": "角色与权限",
                  "link": "/zh/manual/olares/settings/roles-permissions",
                },
                {
                  "text": "创建管理用户",
                  "link": "/zh/manual/olares/settings/manage-team",
                }
              ],
            },
            {
              "text": "管理应用",
              "collapsed": true,
              "items": [
                {
                  "text": "管理应用入口",
                  "link": "/zh/manual/olares/settings/manage-entrance",
                },
                {
                  "text": "自定义应用域名",
                  "link": "/zh/manual/olares/settings/custom-app-domain",
                },
                {
                  text: "管理应用环境变量",
                  link: "/zh/manual/olares/settings/manage-app-env",
                },
              ],
            },
            {
              "text": "管理集成",
              "link": "/zh/manual/olares/settings/integrations",
            },
            {
              "text": "自定义外观",
              "link": "/zh/manual/olares/settings/language-appearance",
            },
            { text: "管理 VPN", link: "/zh/manual/olares/settings/remote-access", },
            {
              "text": "配置网络",
              "collapsed": true,
              "items": [
                {
                  "text": "更改反向代理",
                  "link": "/zh/manual/olares/settings/change-frp",
                },
                {
                  "text": "设置 hosts 文件",
                  "link": "/zh/manual/olares/settings/set-up-hosts",
                },
              ],
            },
            { text: "管理 GPU", link: "/zh/manual/olares/settings/gpu-resource" },
            { text: "视频设置", link: "/zh/manual/olares/settings/video" },
            { text: "文件搜索", link: "/zh/manual/olares/settings/search" },
            {
              "text": "备份与恢复",
              "collapsed": true,
              "items": [
                { text: "备份", link: "/zh/manual/olares/settings/backup" },
                { text: "恢复", link: "/zh/manual/olares/settings/restore" },
              ],
            },
            { text: "开发者资源", link: "/zh/manual/olares/settings/developer" },
          ]
        },
        { "text": "仪表盘", "link": "/zh/manual/olares/resources-usage" },
        { "text": "Profile", "link": "/zh/manual/olares/profile" }
      ]
    },
    {
      text: "教程",
      collapsed: true,
      link: "/zh/manual/best-practices/",
      items: [
        {
          text: "设置自定义域名",
          link: "/zh/manual/best-practices/set-custom-domain",
        },
        {
          text: "使用 Wise 管理知识",
          link: "/zh/manual/best-practices/organize-content",
        },
        {
          text: "安装多节点",
          link: "/zh/manual/best-practices/install-olares-multi-node",
        },
        {
          text: "设置 SMTP",
          link: "/zh/manual/best-practices/set-up-SMTP-service",
        },
        {
          text: "在启用显卡直通的 PVE 上安装 Olares",
          link: "/zh/manual/best-practices/install-olares-gpu-passthrough",
        },
        {
          text: "在 Olares 中扩展存储空间",
          link: "/zh/manual/best-practices/expand-storage-in-olares",
        },
        {
          text: "本地访问 Olares",
          link: "/manual/best-practices/local-access",
        },
      ],
    },
    { text: "术语", link: "/zh/manual/glossary" },
  ],
  "/zh/space/": [
    {
      text: "Olares Space",
      link: "/zh/space/",
      collapsed: true,
      items: [
        {
          text: "管理账号",
          link: "/zh/space/manage-accounts",
        },
        {
          text: "托管 Olares",
          collapsed: true,
          items: [
            {
              text: "创建 Olares",
              link: "/zh/space/create-olares",
            },
            {
              text: "管理 Olares",
              link: "/zh/space/manage-olares",
            },
          ],
        },
        {
          text: "托管域名",
          collapsed: true,
          items: [
            {
              text: "设置自定义域名",
              link: "/zh/space/host-domain",
            },
            {
              text: "管理域名",
              link: "/zh/space/manage-domain",
            },
          ],
        },
        {
          text: "备份与恢复",
          link: "/zh/space/backup-restore",
        },
        { text: "计费", link: "/zh/space/billing" },
      ],
    },
  ],
  "/zh/use-cases/": [
    {
      text: "应用示例",
      link: "/zh/use-cases/",
      items: [
        {
          text: "Stable Diffusion",
          link: "/zh/use-cases/stable-diffusion",
        },
        {
          text: "ComfyUI",
          link: "/zh/use-cases/comfyui",
          collapsed: true,
          items: [
            {
              text: "Manage ComfyUI",
              link: "/zh/use-cases/comfyui-launcher",
            },
            {
              text: "Use ComfyUI for Krita",
              link: "/zh/use-cases/comfyui-for-krita",
            },
          ]
        },
        {
          text: "Ollama",
          link: "/zh/use-cases/ollama",
        },
        {
          text: "Open WebUI",
          link: "/zh/use-cases/openwebui",
        },
        {
          text: "Perplexica",
          link: "/zh/use-cases/perplexica",
        },
        {
          text: "Dify",
          link: "/zh/use-cases/dify",
        },
        {
          text: "Jellyfin",
          link: "/zh/use-cases/stream-media",
        },
        {
          text: "Steam",
          collapsed: true,
          items: [
            {
              text: "在 Olares 本机游玩",
              link: "/zh/use-cases/play-games-directly",
            },
            {
              text: "串流到其他设备",
              link: "/zh/use-cases/stream-game",
            }
          ]
        },
        // {
        //   text: "Redroid",
        //   link: "/zh/use-cases/host-cloud-android",
        // },
      ],
    },
  ],
  "/zh/developer/": [
    {
      text: "概念",
      link: "/zh/developer/concepts/",
      items: [
        { text: "系统架构", link: "/zh/developer/concepts/system-architecture" },
        {
          text: "Olares ID",
          link: "/zh/developer/concepts/olares-id",
          collapsed: true,
          items: [
            {
              text: "去中心化标识符",
              link: "/zh/developer/concepts/did",
            },
            {
              text: "DID Registry",
              link: "/zh/developer/concepts/registry",
            },
            {
              text: "可验证凭证",
              link: "/zh/developer/concepts/vc",
            },
            {
              text: "自治声誉",
              link: "/zh/developer/concepts/reputation",
            },
            //      {
            //     text: "主权网络",
            //   link: "/zh/developer/concepts/self-sovereign-network",
            //  },
            {
              text: "身份钱包",
              link: "/zh/developer/concepts/wallet",
            },
          ],
        },
        { text: "账户", link: "/zh/developer/concepts/account" },
        { text: "应用", link: "/zh/developer/concepts/application" },
        { text: "网络", link: "/zh/developer/concepts/network" },
        { text: "数据", link: "/zh/developer/concepts/data" },
        { text: "密钥", link: "/zh/developer/concepts/secrets" },
      ],
    },
    {
      text: "Olares 安装详解",
      link: "/zh/developer/install/",
      items: [
        {
          text: "安装概述",
          link: "/zh/developer/install/installation-overview",
        },
        {
          text: "安装流程",
          link: "/zh/developer/install/installation-process",
        },
        {
          text: "Olares Home",
          link: "/zh/developer/install/olares-home",
        },
        {
          text: "环境变量",
          link: "/zh/developer/install/environment-variables",
        },
        {
          text: "Olares CLI",
          link: "/zh/developer/install/cli/olares-cli",
          collapsed: true,
          items: [
            {
              text: "backups",
              link: "/zh/developer/install/cli/backups",
              collapsed: true,
              items: [
                { text: "backup", link: "/zh/developer/install/cli/backups-backup" },
                { text: "download", link: "/zh/developer/install/cli/backups-download" },
                { text: "region", link: "/zh/developer/install/cli/backups-region" },
                { text: "restore", link: "/zh/developer/install/cli/backups-restore" },
                { text: "snapshots", link: "/zh/developer/install/cli/backups-snapshots" },
              ],
            }, 
            { text: "change-ip", link: "/zh/developer/install/cli/change-ip" },
            { text: "disk", link: "/zh/developer/install/cli/disk" },   
            { text: "download", link: "/zh/developer/install/cli/download" },
            { text: "gpu", link: "/zh/developer/install/cli/gpu" },
            { text: "info", link: "/zh/developer/install/cli/info" },         
            { text: "install", link: "/zh/developer/install/cli/install" },
            { text: "logs", link: "/zh/developer/install/cli/logs" },
            { text: "node", link: "/zh/developer/install/cli/node" },
            { text: "osinfo", link: "/zh/developer/install/cli/osinfo" },
            { text: "precheck", link: "/zh/developer/install/cli/precheck" },
            { text: "prepare", link: "/zh/developer/install/cli/prepare" },
            { text: "release", link: "/zh/developer/install/cli/release" },
            { text: "start", link: "/zh/developer/install/cli/start" },
            { text: "stop", link: "/zh/developer/install/cli/stop" },
            { text: "uninstall", link: "/zh/developer/install/cli/uninstall" },
            { text: "upgrade", link: "/zh/developer/install/cli/upgrade" },
            {
              text: "user",
              link: "/zh/developer/install/cli/user",
              collapsed: true,
              items: [
                { text: "activate", link: "/zh/developer/install/cli/user-activate" },
                { text: "create", link: "/zh/developer/install/cli/user-create" },
                { text: "delete", link: "/zh/developer/install/cli/user-delete" },
                { text: "get", link: "/zh/developer/install/cli/user-get" },
                { text: "list", link: "/zh/developer/install/cli/user-list" },
                { text: "reset-password", link: "/zh/developer/install/cli/user-reset-password" },
              ],
            },
          ],
        },
        {
          text: "版本说明",
          link: "/zh/developer/install/versioning",
        },
      ],
    },
    {
      text: "开发 Olares 应用",
      link: "/zh/developer/develop/",
      items: [
        {
          text: "使用 Studio 开发",
          collapsed: true,
          link: "/zh/developer/develop/tutorial/",
          items: [
            {
              text: "部署应用",
              link: "/zh/developer/develop/tutorial/deploy",
            },
            {
              text: "使用开发容器",
              link: "/zh/developer/develop/tutorial/develop",
            },
            {
              text: "打包与上传",
              link: "/zh/developer/develop/tutorial/package-upload",
            },
            {
              text: "添加应用素材",
              link: "/zh/developer/develop/tutorial/assets",
            },
          ],
        },
        {
          text: "应用包管理",
          collapsed: true,
          items: [
            {
              text: "应用 Chart 包",
              link: "/zh/developer/develop/package/chart",
            },
            {
              text: "OlaresManifest",
              link: "/zh/developer/develop/package/manifest",
            },
            /*{
              text: "推荐算法",
              link: "/zh/developer/develop/package/recommend",
            },*/
            {
              text: "Helm 扩展",
              link: "/zh/developer/develop/package/extension",
            },
          ],
        },
        //  {
        //      text: "进阶",
        //     collapsed: true,
        //    items: [
        //      {
        //        text: "terminus-info",
        //       link: "/zh/developer/develop/advanced/terminus-info",
        //      },
        //      {
        //       text: "Service Provider",
        //       link: "/zh/developer/develop/advanced/provider",
        //      },
        //     {
        //       text: "AI",
        //       link: "/zh/developer/develop/advanced/ai",
        //     },
        //      { text: "Cookie", link: "/zh/developer/develop/advanced/cookie" },
        //      { text: "数据库", link: "/zh/developer/develop/advanced/database" },
        //     {
        //       text: "账户",
        //       link: "/zh/developer/develop/advanced/account",
        //      },
        //      {
        //       text: "应用市场",
        //       link: "/zh/developer/develop/advanced/market",
        //      },
        // {
        //   text: "Analytic",
        //   link: "/zh/developer/develop/advanced/analytic",
        // },
        //       {
        //        text: "Websocket",
        //        link: "/zh/developer/develop/advanced/websocket",
        //      },
        //      {
        //        text: "文件上传",
        //       link: "/zh/developer/develop/advanced/file-upload",
        //      },
        // {
        //   text: "Rss",
        //   link: "/zh/developer/develop/advanced/rss",
        // },
        //      {
        //         text: "密钥",
        //         link: "/zh/developer/develop/advanced/secret",
        //      },
        // {
        //   text: "Notification",
        //   link: "/zh/developer/develop/advanced/notification",
        // },
        // {
        //   text: "Frontend",
        //   link: "/zh/developer/develop/advanced/frontend",
        // },
        //        {
        //          text: "Kubesphere",
        //          link: "/zh/developer/develop/advanced/kubesphere",
        //         },
        //       ],
        //      },
{
          text: "中间件",
          link: "/zh/developer/develop/mw-overview",
          collapsed: true,
          items: [
            {
              text: "Elasticsearch",
              collapsed: true,
              items :[
                {
                  text: "集成 Elasticsearch",
                  link: "zh/developer/develop/mw-integrate-with-es",                },
                {
                  text: "查看 Elasticsearch 数据",
                  link: "zh/developer/develop/mw-view-es-data",
                },
              ]
            },
            {
              text: "Grafana",
              link: "zh/developer/develop/mw-view-grafana-data",
            },
            {
              text: "MariaDB",
              collapsed: true,
              items :[
                {
                  text: "集成 MariaDB",
                  link: "zh/developer/develop/mw-integrate-with-mariadb",                },
                {
                  text: "查看 MariaDB 数据",
                  link: "zh/developer/develop/mw-view-mariadb-data",
                },
              ]
            },
            {
              text: "MinIO",
              collapsed: true,
              items :[
                {
                  text: "集成 MinIO",
                  link: "zh/developer/develop/mw-integrate-with-minio",                },
                {
                  text: "查看 MinIO 数据",
                  link: "zh/developer/develop/mw-view-minio-data",
                },
              ]
            },
            {
              text: "MongoDB",
              collapsed: true,
              items :[
                {
                  text: "集成 MongoDB",
                  link: "zh/developer/develop/mw-integrate-with-mongodb",                },
                {
                  text: "查看 MongoDB 数据",
                  link: "zh/developer/develop/mw-view-mongodb-data",
                },
              ]
            }, 
            {
              text: "MySQL",
              collapsed: true,
              items :[
                {
                  text: "集成 MySQL",
                  link: "zh/developer/develop/mw-integrate-with-mysql",                },
                {
                  text: "查看 MySQL 数据",
                  link: "zh/developer/develop/mw-view-mysql-data",
                },
              ]
            },           
            {
              text: "NATS",
              link :"zh/developer/develop/mw-view-nats-data",
            },
            {
              text: "OpenTelemetry",
              link :"zh/developer/develop/mw-view-otel-data",
            },
            {
              text: "PostgreSQL",
              collapsed: true,
              items :[
                {
                  text: "集成 PostgreSQL",
                  link: "zh/developer/develop/mw-integrate-with-pg",                },
                {
                  text: "查看 PostgreSQL 数据",
                  link: "zh/developer/develop/mw-view-pg-data",
                }
              ]
            },
            {
              text: "RabbitMQ",
              collapsed: true,
              items :[
                {
                  text: "集成 RabbitMQ",
                  link: "zh/developer/develop/mw-integrate-with-rabbitmq",                },
                {
                  text: "查看 RabbitMQ 数据",
                  link: "zh/developer/develop/mw-view-rabbitmq-data",
                }
              ]
            },
            {
              text: "Redis",
              collapsed: true,
              items :[
                {
                  text: "集成 Redis",
                  link: "zh/developer/develop/mw-integrate-with-redis",                },
                {
                  text: "查看 Redis 数据",
                  link: "zh/developer/develop/mw-view-redis-data",
                },
              ]
            },
          ]
        },
        
      ],
    },
    {
      text: "分发 Olares 应用",
      link: "/zh/developer/develop/distribute-index",
      items: [ 
        {
          text: "提交应用",
          link: "/zh/developer/develop/submit-apps",
        },         
        {
          text: "管理应用",
          link: "/zh/developer/develop/manage-apps",
        },
        {
          text: "推广应用",
          link:"/zh/developer/develop/promote-apps"
        },
        {
          text: "发布付费应用",
          link: "/zh/developer/develop/paid-apps",
        },
        
      ],
    },
    {
      text: "参与贡献",
      items: [
        {
          text: "开发系统应用",
          collapsed: true,
          items: [
            {
              text: "概述",
              link: "/zh/developer/contribute/system-app/overview",
            },
            {
              text: "应用部署配置",
              link: "/zh/developer/contribute/system-app/deployment",
            },
            {
              text: "Olares 权限配置",
              link: "/zh/developer/contribute/system-app/olares-manifest",
            },
            {
              text: "安装",
              link: "/zh/developer/contribute/system-app/install",
            },
            {
              text: "其他",
              link: "/zh/developer/contribute/system-app/other",
            },
          ],
        },
        {
          text: "开发协议",
          collapsed: true,
          items: [
            {
              text: "合约",
              link: "/zh/developer/contribute/olares-id/contract/contract",
              collapsed: true,
              items: [
                {
                  text: "架构",
                  link: "/zh/developer/contribute/olares-id/contract/architecture",
                },
                {
                  text: "DID",
                  collapsed: true,
                  items: [
                    {
                      text: "设计",
                      link: "/zh/developer/contribute/olares-id/contract/did/design",
                    },
                    {
                      text: "官方 Tagger",
                      link: "/zh/developer/contribute/olares-id/contract/did/official-taggers",
                    },
                    {
                      text: "发布历史",
                      link: "/zh/developer/contribute/olares-id/contract/did/release-history",
                    },
                    {
                      text: "FAQ",
                      link: "/zh/developer/contribute/olares-id/contract/did/faq",
                    },
                  ],
                },
                {
                  text: "声誉",
                  link: "/zh/developer/contribute/olares-id/contract/contract-reputation",
                },
                {
                  text: "管理",
                  collapsed: true,
                  items: [
                    {
                      text: "合约",
                      link: "/zh/developer/contribute/olares-id/contract/manage/contract",
                    },
                    {
                      text: "SDK",
                      link: "/zh/developer/contribute/olares-id/contract/manage/sdk",
                    },
                    {
                      text: "环境",
                      link: "/zh/developer/contribute/olares-id/contract/manage/environment",
                    },
                  ],
                },
              ],
            },
            {
              text: "可验证凭证（VC）",
              link: "/zh/developer/contribute/olares-id/verifiable-credential/overview",
              collapsed: true,
              items: [
                {
                  text: "发行方",
                  link: "/zh/developer/contribute/olares-id/verifiable-credential/issuer",
                },
                {
                  text: "验证方",
                  link: "/zh/developer/contribute/olares-id/verifiable-credential/verifer",
                },
                {
                  text: "Olares 案例",
                  link: "/zh/developer/contribute/olares-id/verifiable-credential/olares",
                },
              ],
            },
          ],
        },
      ],
    },
  ],
};

export const zh = defineConfig({
  lang: "zh",
  themeConfig: {
    //logo: "/icon.png",
    socialLinks: [{ icon: "github", link: "https://github.com/beclab/olares" }],

    nav: [
      { text: "Olares", link: "zh/manual/overview" },
      { text: "Olares Space", link: "/zh/space/" },
      { text: "应用示例", link: "/zh/use-cases/" },
      { text: "开发者文档", link: "/zh/developer/concepts/" },
    ],

    sidebar: side,
  },
});