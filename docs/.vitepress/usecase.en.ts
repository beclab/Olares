import { defineConfig, type DefaultTheme } from "vitepress";

export const useCaseSidebar: DefaultTheme.Sidebar = {
    "/use-cases/": [
        {
          text: "AI",
          collapsed: false,
          items: [
            {
              text: "OpenClaw",
              link: "/use-cases/openclaw",
              collapsed: true,
              items: [
                {
                  text: "Integrate with Discord",
                  link: "/use-cases/openclaw-integration",
                },
                {
                  text: "Enable web search",
                  link: "/use-cases/openclaw-web-access",
                },
                {
                  text: "Manage skills and plugins",
                  link: "/use-cases/openclaw-skills",
                },
                {
                  text: "Enable local file access",
                  link: "/use-cases/openclaw-local-access",
                },
                {
                  text: "Enable sandbox",
                  link: "/use-cases/openclaw-enable-sandbox",
                },                                 
                {
                  text: "Upgrade OpenClaw",
                  link: "/use-cases/openclaw-upgrade",
                },
              ]
            },
            {
              text: "Open WebUI",
              collapsed: true,
              items: [
                {
                  text: "Set up with model apps",
                  link: "/use-cases/openwebui",
                },
                {
                  text: "Set up with Ollama",
                  link: "/use-cases/openwebui-ollama",
                },
              ]
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
                  text: "Use ComfyUI in Krita",
                  link: "/use-cases/comfyui-for-krita",
                },
              ]
            },
            {
              text: "Context7",
              link: "/use-cases/context7",
            },
            {
              text: "Ollama",
              link: "/use-cases/ollama",
            },
            // {
            //   text: "Dify",
            //   link: "/use-cases/dify",
            // },
            {
              text: "LLMFit",
              link: "/use-cases/llmfit",
            },
            {  
              text: "OpenCode",
              link: "/use-cases/opencode",
              collapsed: true,
              items: [
                {
                  text: "Manage packages",
                  link: "/use-cases/opencode-packages",
                },
                {
                  text: "Skills and plugins",
                  link: "/use-cases/opencode-extensions",
                },
                {
                  text: "Common issues",
                  link: "/use-cases/opencode-issues",
                },
              ]
            },
            {
              text: "LobeHub (LobeChat)",
              link: "/use-cases/lobechat",
            },
            {
              text: "Vane (Perplexica)",
              link: "/use-cases/perplexica",
            },
            {
              text: "ACE-Step 1.5",
              link: "/use-cases/ace-step-1.5",
            },            
            /*{
              text: "ACE-Step",
              link: "/use-cases/ace-step",
            },*/
            {
              text: "Duix.Avatar",
              link: "/use-cases/duix-avatar",
            },
            {
              text: "PDFMathTranslate",
              link: "/use-cases/pdfmathtranslate",
            },
            // {
            //   text: "Stable Diffusion",
            //   link: "/use-cases/stable-diffusion",
            // },
            {
              text: "DeerFlow 2.0",
              link: "/use-cases/deerflow2",
            },
            {
              text: "DeerFlow",
              link: "/use-cases/deerflow",
            },
          ]
        },
        {
          text: "Virtual machine",
          collapsed: false,
          items: [
            {
              text: "macOS",
              link: "/use-cases/macos",
            },
            {
              text: "Windows",
              link: "/use-cases/windows",
              collapsed: true,
              items: [
                {
                  text: "Known issues",
                  link: "/use-cases/windows-issues",
                }
              ]
            },
            // {
            //   text: "Redroid",
            //   link: "/use-cases/host-cloud-android",
            // },
          ]
        },
        {
          text: "Entertainment",
          collapsed: false,
          items: [
            {
              text: "Immich",
              link: "/use-cases/immich",              
              collapsed: true,
              items: [
                {
                  text: "Import photos from NAS",
                  link: "/use-cases/immich-import-from-nas",
                },
              ]
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
            {
              text: "Komga",
              link: "/use-cases/komga",
            },
          ]
        },
        {
          text: "Productivity",
          collapsed: false,
          items: [
            {
              text: "Excalidraw",
              link: "/use-cases/excalidraw",
            },
            {
              text: "Stirling PDF",
              link: "/use-cases/stirling-pdf",
            },
          ]
        },
        {
          text: "Utilities",
          collapsed: false,
          items: [
            {
              text: "FlareSolverr",
              link: "/use-cases/flaresolverr",
            },
          ]
        },
      ],
    }
