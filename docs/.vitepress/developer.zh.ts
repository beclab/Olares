import { defineConfig, type DefaultTheme } from "vitepress";

export const developerSidebar: DefaultTheme.Sidebar = {
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
    }