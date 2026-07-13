import { defineConfig, type DefaultTheme } from "vitepress";

export const useCaseSidebar: DefaultTheme.Sidebar = {
    "/use-cases/": [
        {
          text: "AI",
          collapsed: false,
          items: [
            {
              text: "AI agent",
              collapsed: false,
              items: [
                {
                  text: "OpenClaw",
                  link: "/use-cases/openclaw",
                  collapsed: true,
                  items: [
                    {
                      text: "Integrate with channels",
                      items: [
                        {
                          text: "Integrate with Discord",
                          link: "/use-cases/openclaw-integration",
                        },
                        {
                          text: "Integrate with WhatsApp",
                          link: "/use-cases/openclaw-channel-whatsapp",
                        },
                      ],
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
                      text: "Manage Olares with OpenClaw",
                      link: "/use-cases/openclaw-olares-skills",
                    },
                    //{
                    //  text: "Enable local file access",
                    //  link: "/use-cases/openclaw-local-access",
                    //},
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
                  text: "Hermes Agent",
                  link: "/use-cases/hermes",
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
                  text: "Claude Code",
                  link: "/use-cases/claude-code",
                },
                {
                  text: "Pool CLI",
                  link: "/use-cases/pool-cli",
                },                
                {
                  text: "Paperclip",
                  link: "/use-cases/paperclip",
                },
                {
                  text: "DeerFlow 2.0",
                  link: "/use-cases/deerflow2",
                },
                // {
                //   text: "DeerFlow",
                //   link: "/use-cases/deerflow",
                // },
                {
                  text: "TradingAgents",
                  link: "/use-cases/tradingagents",
                },
                {
                  text: "NOFX",
                  link: "/use-cases/nofx",
                },
              ],
            },
            {
              text: "AI workbench",
              collapsed: true,
              items: [
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
                  text: "LobeHub (LobeChat)",
                  link: "/use-cases/lobechat",
                },
                {
                  text: "AnythingLLM",
                  link: "/use-cases/anythingllm",
                },
                {
                  text: "Open Notebook",
                  link: "/use-cases/open-notebook",
                },
                {
                  text: "Vane (Perplexica)",
                  link: "/use-cases/perplexica",
                },
                {
                  text: "Dify",
                  link: "/use-cases/dify-upgrade",
                },
              ],
            },
            {
              text: "Creative media",
              collapsed: true,
              items: [
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
                  text: "ACE-Step 1.5",
                  link: "/use-cases/ace-step-1.5",
                },
                /*{
                  text: "ACE-Step",
                  link: "/use-cases/ace-step",
                },*/
                /*{
                  text: "Duix.Avatar",
                  link: "/use-cases/duix-avatar",
                },*/
                // {
                //   text: "Stable Diffusion",
                //   link: "/use-cases/stable-diffusion",
                // },
                {
                  text: "Whisper-WebUI",
                  link: "/use-cases/whisper-webui",
                },
                {
                  text: "Speaches",
                  link: "/use-cases/speaches",
                },
                {
                  text: "IndexTTS2",
                  link: "/use-cases/indextts2",
                },
              ],
            },
            {
              text: "Model services",
              collapsed: true,
              items: [
                {
                  text: "Engine Base apps",
                  link: "/use-cases/llm-base-apps",
                },
                {
                  text: "Ollama",
                  link: "/use-cases/ollama",
                },
                {
                  text: "Bifrost",
                  link: "/use-cases/bifrost",
                },
                {
                  text: "LiteLLM",
                  link: "/use-cases/litellm",
                },
                {
                  text: "TensorZero",
                  link: "/use-cases/tensorzero",
                },
                {
                  text: "LLMFit",
                  link: "/use-cases/llmfit",
                },
                {
                  text: "Xinference",
                  link: "/use-cases/xinference",
                },
                // {
                //   text: "Dify",
                //   link: "/use-cases/dify",
                // },
              ],
            },
            {
              text: "Embodied AI",
              collapsed: true,
              items: [
                {
                  text: "Isaac Lab",
                  link: "/use-cases/isaac-lab",
                },
              ],
            },
          ],
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
              text: "Navidrome",
              link: "/use-cases/navidrome",
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
              collapsed: true,
              items: [
                {
                  text: "Upgrade *Arrs",
                  link: "/use-cases/arrs-upgrade",
                },
                {
                  text: "Upgrade download clients",
                  link: "/use-cases/download-clients-upgrade",
                },
              ]
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
            {
              text: "OnlyOffice",
              link: "/use-cases/onlyoffice",
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
              text: "PDFMathTranslate",
              link: "/use-cases/pdfmathtranslate",
            },
            {
              text: "Context7",
              link: "/use-cases/context7",
            },
            {
              text: "Falco",
              link: "/use-cases/falco",
            },
            {
              text: "Firecrawl",
              link: "/use-cases/firecrawl",
            },
            {
              text: "FlareSolverr",
              link: "/use-cases/flaresolverr",
            },
            {
              text: "SearXNG",
              link: "/use-cases/searxng",
            },
            {
              text: "Home Assistant",
              link: "/use-cases/home-assistant",
            },
          ]
        },
      ],
    }
