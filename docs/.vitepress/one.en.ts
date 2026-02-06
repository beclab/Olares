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
          text: "Access Olares securely",
          link: "/one/access-olares-via-vpn",
        },
        {
          text: "Redeem Olares Space membership",
          link: "/one/redeem-membership",
        },
      ]
    },
    {
      text: "Onboarding",
      items: [
        {
          text: "Open WebUI with Ollama",
          link: "/one/open-webui",
        },
        {
          text: "Generate images with ComfyUI",
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
        {
          text: "Download YouTube videos",
          link: "/one/wise-download",
        },
        {
          text: "Deploy an app",
          link: "/one/deploy",
        },

      ]
    },
//     {
//       text: "Manage",
//       items: [
//         {
//           text: "Set up app entrances",
//           link: "/one/app-entrances",
//         },
//         {
//           text: "Create users",
//           link: "/one/users",
//         },
//       ]
//     },
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
          text: "Play games with Steam Headless",
          link: "/one/steam",
        },
        {
          text: "Access Windows in Olares",
          link: "/one/windows",
        },
        {
          text: "Generate music with Ace-Step",
          link: "/one/ace-step",
        },

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
          text: "SSH into Olares One",
          link: "/one/access-terminal-ssh",
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
          items:
            [
              {
                text: "Manage GPU",
                link: "/one/two-one-gpu",
              },
              {
                text: "Run larger local LLMs",
                link: "/one/two-one-llm",
              }
            ]
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
                text: "Install NVIDIA driver",
                link: "/one/install-nvidia-driver",
              }
            ]
        },
        {
          text: "Create a bootable USB",
          link: "/one/create-drive",
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
          text: "Back up & restore",
          link: "/one/backup-resotre",
        },
        {
          text: "Factory reset",
          link: "/one/factory-reset",
        },
      ]
    },
  ],
}