import { defineConfig, type DefaultTheme } from "vitepress";

export const oneSidebar: DefaultTheme.Sidebar = {
  "/one/": [
    {
      text: "Olares One",
      link: "/one/",
      items: [
        {
          text: "Technical spec",
          link: "/one/spec",
        },
        {
          text: "FAQs",
          link: "/one/faq",
        },
        {
          text: "Known issues",
          link: "/one/known-issues",
        },        
      ],
    },
    {
      text: "Initial setup",
      items: [
        {
          text: "First boot",
          link: "/one/first-boot",
        },
        {
          text: "Access Olares via VPN",
          link: "/one/access-olares-via-vpn",
        },
        {
          text: "Access Olares via .local domain",
          link: "/one/access-olares-via-local-domain",
        },
        {
          text: "Redeem membership",
          link: "/one/redeem-membership",
        },
      ]
    },
    {
      text: "Onboarding",
      items: [
        {
          text: "Chat with local LLMs",
          link: "/one/open-webui",
        },
        {
          text: "Generate images and videos with ComfyUI",
          link: "/one/comfyui",
        },
        {
          text: "Switch GPU mode",
          link: "/one/gpu",
        },
      ]
    },
    {
      text: "Use",
      items: [
        {
          text: "Customize Olares",
          link: "/one/customize",
        },
        {
          text: "Manage files",
          link: "/one/files",
        },
        {
          text: "Install & update apps",
          link: "/one/market",
        },
        {
          text: "Secure passwords",
          link: "/one/vault",
        },
        /*{
          text: "Download YouTube videos",
          link: "/one/wise-download",
        },*/
        {
          text: "Deploy an app",
          link: "/one/deploy",
        },

      ]
    },
    {
       text: "Manage",
       items: [
        {
          text: "Add and manage users",
          link: "/one/create-users",
        },
         {
           text: "Configure application access",
           link: "/one/config-app-access",
         },
       ]
     },
    {
      text: "Monitor",
      items: [
        {
          text: "System resources",
          link: "/one/dashboard",
        },
        {
          text: "Traffic",
          link: "/one/space",
        },
      ]
    },
    {
      text: "Explore",
      items: [
        {
          text: "Play Steam games",
          collapsed: true,
          items: [
            {
              text: "Streaming",
              link: "/one/steam-stream"
            },
            {
              text: "Direct play",
              link: "/one/steam-direct-play",
            }]
        },
        {
          text: "Access Windows in Olares",
          link: "/one/windows",
        },
        /*{
          text: "Generate music with ACE-Step",
          link: "/one/ace-step",
        },*/

        {
          text: "Deep research with DeerFlow",
          link: "/one/deerflow",
        },
      ]
    },
    {
      text: "Advanced",
      items: [
        {
          text: "Access Olares One terminal",
          collapsed: true,
          link: "/one/access-overview",
          items:
            [
              {
                text: "Access via Control Hub",
                link: "/one/access-terminal-control-hub",
              },
              {
                text: "Access via SSH",
                link: "/one/access-terminal-ssh",
              },
              {
                text: "Access via physical console",
                link: "/one/access-local-login",
              },
            ]
        },        
        {
          text: "Expand storage",
          collapsed: true,
          items:
            [
              {
                text: "USB drive",
                link: "/one/expand-storage-usb-drive",
              },
              {
                text: "External SSD",
                link: "/one/expand-storage-external-ssd",
              },
              {
                text: "NVMe SSD",
                link: "/one/expand-storage-internal-ssd",
              },
            ]
        },
        {
          text: "Connect two Olares One",
          link: "/one/connect-two-olares-one"
          // items:
          //   [
          //     {
          //       text: "Manage GPU",
          //       link: "/one/two-one-gpu",
          //     },
          //     {
          //       text: "Run larger local LLMs",
          //       link: "/one/two-one-llm",
          //     }
          //   ]
        },
        {
          text: "Set up with eGPU",
          link: "/one/egpu",
        },
        {
          text: "Dual-boot Olares OS with Windows",
          collapsed: true,
          items:
            [
              {
                text: "Dual-drive setup (Recommended)",
                link: "/one/dual-boot-dual-drive",
              },
              {
                text: "Single-drive setup",
                link: "/one/dual-boot-single-drive",
              }
              ,
              {
                text: "Install drivers on Windows",
                link: "/one/install-nvidia-driver",
              }
            ]
        },
      ]
    },
    {
      text: "System update",
      items: [
        {
          text: "Update OS",
          link: "/one/update",
        },
        {
          text: "Update firmware",
          link: "/one/update-firmware",
        },        
        {
          text: "Back up & restore data",
          link: "/one/backup-resotre",
        },
        {
          text: "Restore Olares One",
          items: [
            {
              text: "Factory reset",
              link: "/one/factory-reset",
            },
            {
              text: "Restore BIOS defaults",
              link: "/one/factory-reset-in-bios",
            },
            {
              text: "Reinstall Olares OS",
              link: "/one/create-drive",
            },
          ],
        },
      ]
    },
  ],
}