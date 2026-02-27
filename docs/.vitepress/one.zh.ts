import { defineConfig, type DefaultTheme } from "vitepress";

export const oneSidebar: DefaultTheme.Sidebar = {
  "/zh/one/": [
    {
      text: "Olares One",
      link: "/zh/one/",
      items: [
        {
          text: "Technical spec",
          link: "/zh/one/spec",
        },
        {
          text: "FAQs",
          link: "/zh/one/faq",
        },
      ],
    },
    {
      text: "Initial setup",
      items: [
        {
          text: "First boot",
          link: "/zh/one/first-boot",
        },
        {
          text: "Access Olares securely",
          link: "/zh/one/access-olares-via-vpn",
        },
        {
          text: "Redeem Olares Space membership",
          link: "/zh/one/redeem-membership",
        },
      ]
    },
    {
      text: "Onboarding",
      items: [
        {
          text: "Open WebUI with Ollama",
          link: "/zh/one/open-webui",
        },
        {
          text: "Generate images with ComfyUI",
          link: "/zh/one/comfyui",
        },
        {
          text: "Switch GPU mode",
          link: "/zh/one/gpu",
        },
      ]
    },
    {
      text: "Use",
      items: [
        {
          text: "Customize Olares",
          link: "/zh/one/customize",
        },
        {
          text: "Manage files",
          link: "/zh/one/files",
        },
        {
          text: "Install & update apps",
          link: "/zh/one/market",
        },
        {
          text: "Secure passwords",
          link: "/zh/one/vault",
        },
        {
          text: "Download YouTube videos",
          link: "/zh/one/wise-download",
        },
        {
          text: "Deploy an app",
          link: "/zh/one/deploy",
        },

      ]
    },
//     {
//       text: "Manage",
//       items: [
//         {
//           text: "Set up app entrances",
//           link: "/zh/one/app-entrances",
//         },
//         {
//           text: "Create users",
//           link: "/zh/one/users",
//         },
//       ]
//     },
    {
      text: "Monitor",
      items: [
        {
          text: "System resources",
          link: "/zh/one/dashboard",
        },
        {
          text: "Traffic",
          link: "/zh/one/space",
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
              link: "/zh/one/steam-stream"
            },
            {
              text: "Direct play",
              link: "/zh/one/steam-direct-play",
            }]
        },
        {
          text: "Access Windows in Olares",
          link: "/zh/one/windows",
        },
        {
          text: "Generate music with Ace-Step",
          link: "/zh/one/ace-step",
        },

        {
          text: "Deep research with DeerFlow",
          link: "/zh/one/deerflow",
        },
      ]
    },
    {
      text: "Advanced",
      items: [
        {
          text: "SSH into Olares One",
          link: "/zh/one/access-terminal-ssh",
        },
        {
          text: "Expand storage",
          collapsed: true,
          items:
            [
              {
                text: "USB drive",
                link: "/zh/one/expand-storage-usb-drive",
              },
              {
                text: "External SSD",
                link: "/zh/one/expand-storage-external-ssd",
              },
              {
                text: "NVMe SSD",
                link: "/zh/one/expand-storage-internal-ssd",
              },
            ]
        },
        {
          text: "Connect two Olares One",
          link: "/zh/one/connect-two-olares-one"
          // items:
          //   [
          //     {
          //       text: "Manage GPU",
          //       link: "/zh/one/two-one-gpu",
          //     },
          //     {
          //       text: "Run larger local LLMs",
          //       link: "/zh/one/two-one-llm",
          //     }
          //   ]
        },
        {
          text: "Set up with eGPU",
          link: "/zh/one/egpu",
        },
        {
          text: "Dual-boot Olares OS with Windows",
          collapsed: true,
          items:
            [
              {
                text: "Dual-drive setup (Recommended)",
                link: "/zh/one/dual-boot-dual-drive",
              },
              {
                text: "Dual-drive setup",
                link: "/zh/one/dual-boot-single-drive",
              }
              ,
              {
                text: "Install NVIDIA driver",
                link: "/zh/one/install-nvidia-driver",
              }
            ]
        },
        {
          text: "Create a bootable USB",
          link: "/zh/one/create-drive",
        },
      ]
    },
    {
      text: "System update",
      items: [
        {
          text: "Update OS",
          link: "/zh/one/update",
        },
        {
          text: "Back up & restore",
          link: "/zh/one/backup-resotre",
        },
        {
          text: "Factory reset",
          link: "/zh/one/factory-reset",
        },
      ]
    },
  ],
}