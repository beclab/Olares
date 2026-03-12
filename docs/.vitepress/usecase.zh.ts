import { defineConfig, type DefaultTheme } from "vitepress";

export const useCaseSidebar: DefaultTheme.Sidebar = {
    "/zh/use-cases/": [
        {
          text: "应用示例",
          link: "/zh/use-cases/",
        },
        {
          text: "AI",
          collapsed: false,
          items: [
            {
              text: "OpenClaw",
              link: "/zh/use-cases/openclaw",
              collapsed: true,
              items: [
                {
                  text: "Integrate with Discord",
                  link: "/zh/use-cases/openclaw-integration",
                },
                {
                  text: "Enable web search",
                  link: "/zh/use-cases/openclaw-web-access",
                },
                {
                  text: "Manage skills and plugins",
                  link: "/zh/use-cases/openclaw-skills",
                },
                {
                  text: "Upgrade OpenClaw",
                  link: "/zh/use-cases/openclaw-upgrade",
                },
              ]
            },
            {
              text: "Open WebUI",
              link: "/zh/use-cases/openwebui",
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
            // {
            //   text: "Dify",
            //   link: "/zh/use-cases/dify",
            // },
            {
              text: "LobeHub (LobeChat)",
              link: "/zh/use-cases/lobechat",
            },
            {
              text: "Perplexica",
              link: "/zh/use-cases/perplexica",
            },
            {
              text: "ACE-Step",
              link: "/zh/use-cases/ace-step",
            },
            {
              text: "Duix.Avatar",
              link: "/zh/use-cases/duix-avatar",
            },
            {
              text: "PDFMathTranslate",
              link: "/zh/use-cases/pdfmathtranslate",
            },
            // {
            //   text: "Stable Diffusion",
            //   link: "/zh/use-cases/stable-diffusion",
            // },
            {
              text: "DeerFlow",
              link: "/zh/use-cases/deerflow",
            },
          ]
        },
        {
          text: "虚拟机",
          collapsed: false,
          items: [
            // {
            //   text: "macOS",
            //   link: "/zh/use-cases/macOS",
            // },
            {
              text: "Windows",
              link: "/zh/use-cases/windows",
            },
            // {
            //   text: "Redroid",
            //   link: "/zh/use-cases/host-cloud-android",
            // },
          ]
        },
        {
          text: "娱乐",
          collapsed: false,
          items: [
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
          ]
        },
        {
          text: "效率",
          collapsed: false,
          items: [
            {
              text: "Stirling PDF",
              link: "/zh/use-cases/stirling-pdf",
            },
          ],
        },
      ],
    };