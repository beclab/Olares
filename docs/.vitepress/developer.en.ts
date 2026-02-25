import { defineConfig, type DefaultTheme } from "vitepress";

export const developerSidebar: DefaultTheme.Sidebar = {
 "/developer/": [
    {
      text: "Concepts",
      link: "/developer/concepts/",
      items: [
        { text: "Olares architecture", link: "/developer/concepts/system-architecture" },
        {
          text: "Olares ID",
          link: "/developer/concepts/olares-id",
          collapsed: true,
          items: [
            {
              text: "Decentralized ID",
              link: "/developer/concepts/did",
            },
            {
              text: "Blockchain Registry",
              link: "/developer/concepts/registry",
            },
            {
              text: "Verifiable Credential",
              link: "/developer/concepts/vc",
            },
            {
              text: "Autonomous Reputation",
              link: "/developer/concepts/reputation",
            },
            //  {
            //   text: "Self-Sovereign Network",
            //    link: "/developer/concepts/self-sovereign-network",
            //   },
            {
              text: "Identity Wallet",
              link: "/developer/concepts/wallet",
            },
          ],
        },
        { text: "Account", link: "/developer/concepts/account" },
        { text: "Application", link: "/developer/concepts/application" },
        { text: "Network", link: "/developer/concepts/network" },
        { text: "Data", link: "/developer/concepts/data" },
        { text: "Secrets", link: "/developer/concepts/secrets" },
      ],
    },
    {
      text: "Installation deep-dive",
      link: "/developer/install/",
      items: [
        {
          text: "Installation architecture",
          link: "/developer/install/installation-overview",
        },
        {
          text: "Installation process",
          link: "/developer/install/installation-process",
        },
        {
          text: "Olares Home",
          link: "/developer/install/olares-home",
        },
        {
          text: "Environment variables",
          link: "/developer/install/environment-variables",
        },
        {
          text: "Olares CLI",
          link: "/developer/install/cli/olares-cli",
          collapsed: true,
          items: [
            { text: "Access Olares terminal", link: "/developer/reference/access-olares-terminal" },
            {
              text: "backups",
              link: "/developer/install/cli/backups",
              collapsed: true,
              items: [
                { text: "backup", link: "/developer/install/cli/backups-backup" },
                { text: "download", link: "/developer/install/cli/backups-download" },
                { text: "region", link: "/developer/install/cli/backups-region" },
                { text: "restore", link: "/developer/install/cli/backups-restore" },
                { text: "snapshots", link: "/developer/install/cli/backups-snapshots" },
              ],
            }, 
            { text: "change-ip", link: "/developer/install/cli/change-ip" },
            { text: "disk", link: "/developer/install/cli/disk" },   
            { text: "download", link: "/developer/install/cli/download" },
            { text: "gpu", link: "/developer/install/cli/gpu" },
            { text: "info", link: "/developer/install/cli/info" },         
            { text: "install", link: "/developer/install/cli/install" },
            { text: "logs", link: "/developer/install/cli/logs" },
            { text: "node", link: "/developer/install/cli/node" },
            { text: "osinfo", link: "/developer/install/cli/osinfo" },
            { text: "precheck", link: "/developer/install/cli/precheck" },
            { text: "prepare", link: "/developer/install/cli/prepare" },
            { text: "release", link: "/developer/install/cli/release" },
            { text: "start", link: "/developer/install/cli/start" },
            { text: "stop", link: "/developer/install/cli/stop" },
            { text: "uninstall", link: "/developer/install/cli/uninstall" },
            { text: "upgrade", link: "/developer/install/cli/upgrade" },
            {
              text: "user",
              link: "/developer/install/cli/user",
              collapsed: true,
              items: [
                { text: "activate", link: "/developer/install/cli/user-activate" },
                { text: "create", link: "/developer/install/cli/user-create" },
                { text: "delete", link: "/developer/install/cli/user-delete" },
                { text: "get", link: "/developer/install/cli/user-get" },
                { text: "list", link: "/developer/install/cli/user-list" },
                { text: "reset-password", link: "/developer/install/cli/user-reset-password" },
              ],
            },
          ],
        },
        {
          text: "Olares versioning",
          link: "/developer/install/versioning",
        },
      ],
    },
    {
      text: "Develop Olares apps",
      link: "/developer/develop/",
      items: [
        {
          text: "Develop with Studio",
          collapsed: true,
          link: "/developer/develop/tutorial/",
          items: [
            {
              text: "Deploy an app",
              link: "/developer/develop/tutorial/deploy",
            },
            {
              text: "Develop in a dev container",
              link: "/developer/develop/tutorial/develop",
            },
            {
              text: "Package and upload",
              link: "/developer/develop/tutorial/package-upload",
            },
            {
              text: "Add app assets",
              link: "/developer/develop/tutorial/assets",
            },
          ],
        },
        {
          text: "Application package",
          collapsed: true,
          items: [
            {
              text: "Application chart",
              link: "/developer/develop/package/chart",
            },
            {
              text: "OlaresManifest",
              link: "/developer/develop/package/manifest",
            },
            /*{
              text: "Recommendation",
              link: "/developer/develop/package/recommend",
            },*/
            {
              text: "Helm extension",
              link: "/developer/develop/package/extension",
            },
          ],
        },
        //  {
        //   text: "Advanced",
        //   collapsed: true,
        //  items: [
        //    {
        //       text: "terminus-info",
        //      link: "/developer/develop/advanced/terminus-info",
        //     },
        //     {
        //       text: "Service provider",
        //      link: "/developer/develop/advanced/provider",
        //      },
        //     {
        //        text: "AI",
        //        link: "/developer/develop/advanced/ai",
        //       },
        //      { text: "Cookie", link: "/developer/develop/advanced/cookie" },
        //     { text: "Database", link: "/developer/develop/advanced/database" },
        //     {
        //       text: "Account",
        //       link: "/developer/develop/advanced/account",
        //      },
        //     {
        //       text: "Market",
        //       link: "/developer/develop/advanced/market",
        //      },
        //       {
        //       text: "Websocket",
        //       link: "/developer/develop/advanced/websocket",
        //     },
        //      {
        //       text: "File upload",
        //       link: "/developer/develop/advanced/file-upload",
        //      },
        //     {
        //       text: "Secret",
        //       link: "/developer/develop/advanced/secret",
        //      },
        //      {
        //        text: "Kubesphere",
        //      link: "/developer/develop/advanced/kubesphere",
        //      },
        //   ],
        //   },
        {
          text: "Middleware",
          link: "/developer/develop/mw-overview",
          collapsed: true,
          items: [
            {
              text: "Elasticsearch",
              collapsed: true,
              items :[
                {
                  text: "Integrate with Elasticsearch",
                  link: "/developer/develop/mw-integrate-with-es",                },
                {
                  text: "View Elasticsearch data",
                  link: "/developer/develop/mw-view-es-data",
                },
              ]
            },
            {
              text: "Grafana",
              link :"/developer/develop/mw-view-grafana-data",
            },
            {
              text: "MariaDB",
              collapsed: true,
              items :[
                {
                  text: "Integrate with MariaDB",
                  link: "/developer/develop/mw-integrate-with-mariadb",                },
                {
                  text: "View MariaDB data",
                  link: "/developer/develop/mw-view-mariadb-data",
                },
              ]
            },
            {
              text: "MinIO",
              collapsed: true,
              items :[
                {
                  text: "Integrate with MinIO",
                  link: "/developer/develop/mw-integrate-with-minio",                },
                {
                  text: "View MinIO data",
                  link: "/developer/develop/mw-view-minio-data",
                },
              ]
            },
            {
              text: "MongoDB",
              collapsed: true,
              items :[
                {
                  text: "Integrate with MongoDB",
                  link: "/developer/develop/mw-integrate-with-mongodb",                },
                {
                  text: "View MongoDB data",
                  link: "/developer/develop/mw-view-mongodb-data",
                },
              ]
            },
            {
              text: "MySQL",
              collapsed: true,
              items :[
                {
                  text: "Integrate with MySQL",
                  link: "/developer/develop/mw-integrate-with-mysql",                },
                {
                  text: "View MySQL data",
                  link: "/developer/develop/mw-view-mysql-data",
                },
              ]
            },  
            {
              text: "NATS",
              link :"/developer/develop/mw-view-nats-data",
            },
            {
              text: "OpenTelemetry",
              link :"/developer/develop/mw-view-otel-data",
            },
            {
              text: "PostgreSQL",
              collapsed: true,
              items :[
                {
                  text: "Integrate with PostgreSQL",
                  link: "/developer/develop/mw-integrate-with-pg",                },
                {
                  text: "View PostgreSQL data",
                  link: "/developer/develop/mw-view-pg-data",
                }
              ]
            },
            {
              text: "RabbitMQ",
              collapsed: true,
              items :[
                {
                  text: "Integrate with RabbitMQ",
                  link: "/developer/develop/mw-integrate-with-rabbitmq",                },
                {
                  text: "View RabbitMQ data",
                  link: "/developer/develop/mw-view-rabbitmq-data",
                }
              ]
            },
            {
              text: "Redis",
              collapsed: true,
              items :[
                {
                  text: "Integrate with Redis",
                  link: "/developer/develop/mw-integrate-with-redis",                },
                {
                  text: "View Redis data",
                  link: "/developer/develop/mw-view-redis-data",
                }
              ]
            },
          ]
        },
      ],
    },
    {
      text: "Distribute Olares apps",
      link: "/developer/develop/distribute-index",
      items: [ 
        {
          text: "Summit apps",
          link: "/developer/develop/submit-apps",
        },         
        {
          text: "Manage app lifecycle",
          link: "/developer/develop/manage-apps",
        },
        {
          text: "Promote your apps",
          link:"/developer/develop/promote-apps"
        },
        {
          text: "Publish paid apps",
          link: "/developer/develop/paid-apps",
        },
        
      ],
    },
    {
      text: "Contribute to Olares",
      items: [
        {
          text: "Develop system app",
          collapsed: true,
          items: [
            {
              text: "Overview",
              link: "/developer/contribute/system-app/overview",
            },
            {
              text: "Configure deployment",
              link: "/developer/contribute/system-app/deployment",
            },
            {
              text: "Configure permissions",
              link: "/developer/contribute/system-app/olares-manifest",
            },
            {
              text: "Install",
              link: "/developer/contribute/system-app/install",
            },
            {
              text: "Other",
              link: "/developer/contribute/system-app/other",
            },
          ],
        },
        {
          text: "Develop protocols",
          collapsed: true,
          items: [
            {
              text: "Contract",
              link: "/developer/contribute/olares-id/contract/contract",
              collapsed: true,
              items: [
                {
                  text: "Architecture",
                  link: "/developer/contribute/olares-id/contract/architecture",
                },
                {
                  text: "DID",
                  collapsed: true,
                  items: [
                    {
                      text: "Design",
                      link: "/developer/contribute/olares-id/contract/did/design",
                    },
                    {
                      text: "Official Taggers",
                      link: "/developer/contribute/olares-id/contract/did/official-taggers",
                    },
                    {
                      text: "Release History",
                      link: "/developer/contribute/olares-id/contract/did/release-history",
                    },
                    {
                      text: "FAQ",
                      link: "/developer/contribute/olares-id/contract/did/faq",
                    },
                  ],
                },
                {
                  text: "Reputation",
                  link: "/developer/contribute/olares-id/contract/contract-reputation",
                },
                {
                  text: "Manage",
                  collapsed: true,
                  items: [
                    {
                      text: "Contract",
                      link: "/developer/contribute/olares-id/contract/manage/contract",
                    },
                    {
                      text: "SDK",
                      link: "/developer/contribute/olares-id/contract/manage/sdk",
                    },
                    {
                      text: "Environment",
                      link: "/developer/contribute/olares-id/contract/manage/environment",
                    },
                  ],
                },
              ],
            },
            {
              text: "Verifiable Credential",
              link: "/developer/contribute/olares-id/verifiable-credential/overview",
              collapsed: true,
              items: [
                {
                  text: "Issuer",
                  link: "/developer/contribute/olares-id/verifiable-credential/issuer",
                },
                {
                  text: "Verifer",
                  link: "/developer/contribute/olares-id/verifiable-credential/verifer",
                },
                {
                  text: "Olares",
                  link: "/developer/contribute/olares-id/verifiable-credential/olares",
                },
              ],
            },
          ],
        },
      ],
    },
  ],
}