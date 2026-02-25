import { defineConfig, type DefaultTheme } from "vitepress";

export const useCaseSidebar: DefaultTheme.Sidebar = {
    "/use-cases/": [
        {
          text: "Use cases",
          link: "/use-cases/",
          items: [
            {
              text: "Stable Diffusion",
              link: "/use-cases/stable-diffusion",
            },
            {
              text: "ComfyUI",
              link: "/use-cases/comfyui",
              collapsed: true,
              items: [
                {
                  text: "Manage ComfyUI",
                  link: "/use-cases/comfyui-launcher",
                },
                {
                  text: "Use ComfyUI for Krita",
                  link: "/use-cases/comfyui-for-krita",
                },
              ]
            },
            {
              text: "Ollama",
              link: "/use-cases/ollama",
            },
            {
              text: "Open WebUI",
              link: "/use-cases/openwebui",
            },
            {
              text: "Perplexica",
              link: "/use-cases/perplexica",
            },
            {
              text: "Dify",
              link: "/use-cases/dify",
            },
            {
              text: "Jellyfin",
              link: "/use-cases/stream-media",
            },
            {
              text: "Steam",
              collapsed: true,
              items: [
                {
                  text: "Play directly on Olares",
                  link: "/use-cases/play-games-directly",
                },
                {
                  text: "Stream to other devices",
                  link: "/use-cases/stream-game",
                }
              ]
            },
            //         {
            //           text: "Redroid",
            //           link: "/use-cases/host-cloud-android",
            //         },
            {
              text: "Windows",
              link: "/use-cases/windows",
            },
            {
              text: "DeerFlow",
              link: "/use-cases/deerflow",
            },
            {
              text: "Duix.Avatar",
              link: "/use-cases/duix-avatar",
            },
            {
              text: "ACE-Step",
              link: "/use-cases/ace-step",
            },
            {
              text: "Stirling PDF",
              link: "/use-cases/stirling-pdf",
            },
            {
              text: "PDFMathTranslate",
              link: "/use-cases/pdfmathtranslate",
            },
            {
              text: "LobeChat",
              link: "/use-cases/lobechat",
            },
            {
              text: "OpenClaw",
              link: "/use-cases/openclaw",
            },
          ],
        },
      ],
    }