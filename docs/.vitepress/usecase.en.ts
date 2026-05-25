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
                {
                  text: "Common issues",
                  link: "/use-cases/openclaw-common-issues",
                },
              ]
            },
            {
              text: "Hermes Agent",
              link: "/use-cases/hermes",  
            },
            {
              text: "Claude Code",
              link: "/use-cases/claude-code",
            },                         
            {
              text: "OpenCode",
              link: "/use-cases/opencode",
              collapsed: true,
              items: [
                {
                  text: "Orchestrate with oh-my-openagent",
                  link: "/use-cases/opencode-omo",
                },
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
              text: "Open WebUI",
              link: "/use-cases/openwebui",
              collapsed: true,
              items: [
                {
                  text: "Set up multi-user access",
                  link: "/use-cases/openwebui-multiuser",
                },
                {
                  text: "Configure voice interactions",
                  link: "/use-cases/openwebui-audio",
                },
                {
                  text: "Enable web search",
                  link: "/use-cases/openwebui-search",
                },
                {
                  text: "Query documents in chat",
                  link: "/use-cases/openwebui-knowledge",
                },
                {
                  text: "Common issues",
                  link: "/use-cases/openwebui-common-issues",
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
                {
                  text: "Common issues",
                  link: "/use-cases/comfyui-common-issues",
                },
              ]
            },
            {
              text: "NemoClaw",
              link: "/use-cases/nemoclaw",
              collapsed: true,
              items: [
                {
                  text: "Integrate with Google Workspace",
                  link: "/use-cases/nemoclaw-google-workspace",
                },
                {
                  text: "Manage Olares with Olares CLI",
                  link: "/use-cases/nemoclaw-olares-cli",
                },
                {
                  text: "Common issues",
                  link: "/use-cases/nemoclaw-common-issues",
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
            {
              text: "Open Notebook",
              link: "/use-cases/open-notebook",
            },
            // {
            //   text: "Dify",
            //   link: "/use-cases/dify",
            // },
            {
              text: "LiteLLM",
              link: "/use-cases/litellm",
            },
            {
              text: "LLMFit",
              link: "/use-cases/llmfit",
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
              text: "AnythingLLM",
              link: "/use-cases/anythingllm",
            },
            {
              text: "Bifrost",
              link: "/use-cases/bifrost",
            },
            /*{
              text: "Duix.Avatar",
              link: "/use-cases/duix-avatar",
            },*/
            {
              text: "Paperclip",
              link: "/use-cases/paperclip",
            },
            {
              text: "PDFMathTranslate",
              link: "/use-cases/pdfmathtranslate",
            },
            {
              text: "Whisper-WebUI",
              link: "/use-cases/whisper-webui",
            },
            {
              text: "Speaches",
              link: "/use-cases/speaches",
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
              text: "IndexTTS2",
              link: "/use-cases/indextts2",
            },
            {
              text: "NOFX",
              link: "/use-cases/nofx",
            },
            {
              text: "TensorZero",
              link: "/use-cases/tensorzero",
            },           
            // {
            //   text: "DeerFlow",
            //   link: "/use-cases/deerflow",
            // },

          ]
        },
        {
          text: "Embodied AI",
          collapsed: false,
          items: [
            {
              text: "Isaac Lab",
              link: "/use-cases/isaac-lab",
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
                  text: "Enable Intel integrated GPU passthrough",
                  link: "/use-cases/windows-intel-gpu-passthrough",
                },
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
            {
              text: "TREK (NOMAD)",
              link: "/use-cases/trek",
              collapsed: true,
              items: [
                {
                  text: "Advanced settings",
                  link: "/use-cases/trek-advanced-settings",
                }
              ]
            },
            {
              text: "*Arrs",
              link: "/use-cases/arrs",
            },                        
          ]
        },
        {
          text: "Productivity",
          collapsed: false,
          items: [
            {
              text: "Plane",
              link: "/use-cases/plane",
            },
            {
              text: "Excalidraw",
              link: "/use-cases/excalidraw",
            },
            {
              text: "Karakeep",
              link: "/use-cases/karakeep",
            },
            {
              text: "Penpot",
              link: "/use-cases/penpot",
            },
            {
              text: "RNA Sequencing",
              link: "/use-cases/rna-sequencing",
            },
            // {
            //   text: "Stirling PDF",
            //   link: "/use-cases/stirling-pdf",
            // },
          ]
        },
        {
          text: "Developer tools",
          collapsed: false,
          items: [
            {
              text: "JupyterHub",
              link: "/use-cases/jupyterhub",
            },
          ]
        },
        {
          text: "Utilities",
          collapsed: false,
          items: [
            {
              text: "Falco",
              link: "/use-cases/falco",
            },
            {
              text: "FlareSolverr",
              link: "/use-cases/flaresolverr",
            },
          ]
        },
      ],
    }
