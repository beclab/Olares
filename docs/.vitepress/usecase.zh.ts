import { defineConfig, type DefaultTheme } from "vitepress";

export const useCaseSidebar: DefaultTheme.Sidebar = {
    "/zh/use-cases/": [
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
                  text: "集成 Discord",
                  link: "/zh/use-cases/openclaw-integration",
                },
                {
                  text: "启用网页搜索",
                  link: "/zh/use-cases/openclaw-web-access",
                },
                {
                  text: "管理技能和插件",
                  link: "/zh/use-cases/openclaw-skills",
                },
                {
                  text: "启用本地文件访问",
                  link: "/zh/use-cases/openclaw-local-access",
                },
                {
                  text: "启用沙盒",
                  link: "/use-cases/openclaw-enable-sandbox",
                },                                 
                {
                  text: "升级 OpenClaw",
                  link: "/zh/use-cases/openclaw-upgrade",
                },
                {
                  text: "常见问题",
                  link: "/zh/use-cases/openclaw-common-issues",
                }                
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
                  text: "管理 ComfyUI",
                  link: "/zh/use-cases/comfyui-launcher",
                },
                {
                  text: "在 Krita 中使用 ComfyUI",
                  link: "/zh/use-cases/comfyui-for-krita",
                },
              ]
            },
            {
              text: "Context7",
              link: "/zh/use-cases/context7",
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
              text: "LiteLLM",
              link: "/zh/use-cases/litellm",
            },
            {
              text: "LLMFit",
              link: "/zh/use-cases/llmfit",
            },
            {  
              text: "OpenCode",
              link: "/zh/use-cases/opencode",
              collapsed: true,
              items: [
                {
                  text: "包管理",
                  link: "/zh/use-cases/opencode-packages",
                },
                {
                  text: "技能与插件",
                  link: "/zh/use-cases/opencode-extensions",
                },
                {
                  text: "常见问题",
                  link: "/zh/use-cases/opencode-issues",
                },
              ]
            },
            {
              text: "LobeHub (LobeChat)",
              link: "/zh/use-cases/lobechat",
            },
            {
              text: "Vane (Perplexica)",
              link: "/zh/use-cases/perplexica",
            },
            {
              text: "ACE-Step 1.5",
              link: "/zh/use-cases/ace-step-1.5",
            },            
            /*{
              text: "ACE-Step",
              link: "/zh/use-cases/ace-step",
            },*/
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
              text: "DeerFlow 2.0",
              link: "/zh/use-cases/deerflow2",
            },
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
            {
              text: "macOS",
              link: "/zh/use-cases/macos",
            },
            {
              text: "Windows",
              link: "/zh/use-cases/windows",
              collapsed: true,
              items:[
                {
                  text: "Known issues",
                  link: "/zh/use-cases/windows-issues",
                },
              ]
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
              text: "Immich",
              link: "/zh/use-cases/immich",              
              collapsed: true,
              items: [
                {
                  text: "Import photos from NAS",
                  link: "/zh/use-cases/immich-import-from-nas",
                },
              ]
            },
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
            {
              text: "Komga",
              link: "/zh/use-cases/komga",
            },
          ]
        },
        {
          text: "效率",
          collapsed: false,
          items: [
            {
              text: "Plane",
              link: "/zh/use-cases/plane",
              text: "Excalidraw",
              link: "/zh/use-cases/excalidraw",
            },
            {
              text: "Stirling PDF",
              link: "/zh/use-cases/stirling-pdf",
            },
          ],
        },
        {
          text: "工具",
          collapsed: false,
          items: [
            {
              text: "FlareSolverr",
              link: "/zh/use-cases/flaresolverr",
            },
          ],
        },
      ],
    };