import { defineConfig, type DefaultTheme } from "vitepress";

export const oneSidebar: DefaultTheme.Sidebar = {
  "/zh/one/": [
    {
      text: "Olares One",
      link: "/zh/one/",
      items: [
        {
          text: "技术规格",
          link: "/zh/one/spec",
        },
        {
          text: "常见问题",
          link: "/zh/one/faq",
        },
        {
          text: "已知问题",
          link: "/zh/one/known-issues",
        },         
      ],
    },
    {
      text: "初始设置",
      items: [
        {
          text: "首次启动",
          link: "/zh/one/first-boot",
        },
        {
          text: "兑换会员",
          link: "/zh/one/redeem-membership",
        },
      ]
    },
    {
      text: "入门使用",
      items: [
        {
          text: "通过自然语言管理 Olares",
          link: "/zh/one/olares-onboarding",
        },
        {
          text: "通过 VPN 访问 Olares",
          link: "/zh/one/access-olares-via-vpn",
        },
        {
          text: "通过 .local 域名访问 Olares",
          link: "/zh/one/access-olares-via-local-domain",
        },
      ]
    },
    {
      text: "高级",
      items: [
        {
          text: "访问 Olares One 终端",
          collapsed: true,
          link: "/zh/one/access-overview",
          items:
            [
              {
                text: "通过 Control Hub 访问",
                link: "/zh/one/access-terminal-control-hub",
              },
              {
                text: "通过 SSH 访问",
                link: "/zh/one/access-terminal-ssh",
              },
              {
                text: "直接在设备上访问",
                link: "/zh/one/access-physical-console",
              },
            ]
        },        
        {
          text: "扩展存储",
          collapsed: true,
          items:
            [
              {
                text: "USB 驱动器",
                link: "/zh/one/expand-storage-usb-drive",
              },
              {
                text: "外置 SSD",
                link: "/zh/one/expand-storage-external-ssd",
              },
              {
                text: "NVMe SSD",
                link: "/zh/one/expand-storage-internal-ssd",
              },
            ]
        },
        {
          text: "连接两台 Olares One",
          link: "/zh/one/connect-two-olares-one",
          // items:
          //   [
          //     {
          //       text: "Manage accelerator resources",
          //       link: "/zh/one/two-one-gpu",
          //     },
          //     {
          //       text: "Run larger local LLMs",
          //       link: "/zh/one/two-one-llm",
          //     }
          //   ]
        },
        {
          text: "使用 eGPU 设置",
          link: "/zh/one/egpu",
        },        
        {
          text: "Olares OS 与 Windows 双启动",
          collapsed: true,
          items:
            [
              {
                text: "双硬盘设置（推荐）",
                link: "/zh/one/dual-boot-dual-drive",
              },
              {
                text: "单硬盘设置",
                link: "/zh/one/dual-boot-single-drive",
              }
              ,
              {
                text: "在 Windows 上安装驱动",
                link: "/zh/one/install-nvidia-driver",
              },
              {
                text: "故障排除",
                link: "/zh/one/dual-boot-windows-troubleshooting",
              }
            ]
        },
        /*{
          text: "Olares OS 与 Ubuntu 双启动",
          link: "/zh/one/dual-boot-ubuntu-dual-drive",
        },*/
        {
          text: "在 Olares One 上安装 Windows",
          link: "/zh/one/install-windows-primary-drive",
        },
        {
          text: "在 Olares One 上安装 Ubuntu",
          collapsed: true,
          items:
            [
              {
                text: "在 Olares One 上安装 Ubuntu Server",
                link: "/zh/one/install-ubuntu-server",
              },
              {
                text: "在 Olares One 上安装 Ubuntu Desktop",
                link: "/zh/one/install-ubuntu-desktop",
              }
            ]
        },                 
      ]
    },
    {
      text: "系统更新",
      items: [
        {
          text: "更新 OS",
          link: "/zh/one/update",
        },       
        {
          text: "恢复 Olares One",
          items: [
            {
              text: "恢复出厂设置",
              link: "/zh/one/factory-reset",
            },
            {
              text: "恢复 BIOS 默认设置",
              link: "/zh/one/factory-reset-in-bios",
            },
            {
              text: "重新安装 Olares OS",
              link: "/zh/one/create-drive",
            },
            {
              text: "创建可启动 USB 驱动器",
              link: "/zh/one/create-bootable-usb",
            },
          ],
        },
        {
          text: "管理 BIOS 和 EC",
          link: "/zh/one/update-firmware",
        },         
      ]
    },
  ],
}
