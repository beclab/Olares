import { defineConfig, type DefaultTheme } from "vitepress";
import { oneSidebar } from './one.zh.ts';
import { useCaseSidebar } from './usecase.zh.ts';
import { developerSidebar } from './developer.zh.ts';
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
              text: "安装配置与访问",
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
          ],
        },
        {
          text: "故障排查",
          collapsed: true,
          items: [
            {
              text: "内存不足或没有释放",
              link: "/zh/manual/help/ts-free-memory",
            },
            {
              text: "应用市场应用缺失",
              link: "/zh/manual/help/ts-missing-apps",
            }
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
      "text": "Olares 应用",
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
            { text: "高级设置", link: "/zh/manual/olares/settings/developer" },
          ]
        },
        { "text": "仪表盘", "link": "/zh/manual/olares/resources-usage" },
        { "text": "Profile", "link": "/zh/manual/olares/profile" }
      ]
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
        text: "Olares Space",
        link: "/zh/manual/space/index",
        collapsed: true,
        items: [
          {
            text: "管理账号",
            link: "/zh/manual/space/manage-accounts",
          },
          {
            text: "托管 Olares",
            collapsed: true,
            items: [
              {
                text: "创建 Olares",
                link: "/zh/manual/space/create-olares",
              },
              {
                text: "管理 Olares",
                link: "/zh/manual/space/manage-olares",
              },
            ],
          },
          {
            text: "托管域名",
            collapsed: true,
            items: [
              {
                text: "设置自定义域名",
                link: "/zh/manual/space/host-domain",
              },
              {
                text: "管理域名",
                link: "/zh/manual/space/manage-domain",
              },
            ],
          },
          {
            text: "备份与恢复",
            link: "/zh/manual/space/backup-restore",
          },
          { text: "计费", link: "/zh/manual/space/billing" },
        ],
      },
    {
      text: "教程",      
      link: "/zh/manual/best-practices/",
      collapsed: true,
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
          link: "/zh/manual/best-practices/local-access",
        },
      ],
    },
    { text: "术语", link: "/zh/manual/glossary" },
  ],
};

export const zh = defineConfig({
  lang: "zh",
  themeConfig: {
    //logo: "/icon.png",
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