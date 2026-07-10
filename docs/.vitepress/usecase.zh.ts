import { defineConfig, type DefaultTheme } from "vitepress";

export const useCaseSidebar: DefaultTheme.Sidebar = {
    "/zh/use-cases/": [
        {
          text: "AI",
          collapsed: false,
          items: [
            {
              text: "AI Agent",
              collapsed: false,
              items: [
                {
                  text: "OpenClaw",
                  link: "/zh/use-cases/openclaw",
                  collapsed: true,
                  items: [
                    {
                      text: "集成聊天应用",
                      items: [
                        {
                          text: "集成 Discord",
                          link: "/zh/use-cases/openclaw-integration",
                        },
                        {
                          text: "集成 WhatsApp",
                          link: "/zh/use-cases/openclaw-channel-whatsapp",
                        },
                      ],
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
                      text: "使用 OpenClaw 管理 Olares",
                      link: "/zh/use-cases/openclaw-olares-skills",
                    },
                    //{
                    //  text: "启用本地文件访问",
                    //  link: "/zh/use-cases/openclaw-local-access",
                    //},
                    {
                      text: "启用沙盒",
                      link: "/zh/use-cases/openclaw-enable-sandbox",
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
                  text: "NemoClaw",
                  link: "/zh/use-cases/nemoclaw",
                  collapsed: true,
                  items: [
                    {
                      text: "集成 Google Workspace",
                      link: "/zh/use-cases/nemoclaw-google-workspace",
                    },
                    {
                      text: "使用 Olares CLI 管理 Olares",
                      link: "/zh/use-cases/nemoclaw-olares-cli",
                    },
                    {
                      text: "常见问题",
                      link: "/zh/use-cases/nemoclaw-common-issues",
                    },
                  ]
                },
                {
                  text: "Hermes Agent",
                  link: "/zh/use-cases/hermes",
                },
                {
                  text: "OpenCode",
                  link: "/zh/use-cases/opencode",
                  collapsed: true,
                  items: [
                    {
                      text: "使用 oh-my-openagent 编排多 Agent",
                      link: "/zh/use-cases/opencode-omo",
                    },
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
                  text: "Claude Code",
                  link: "/zh/use-cases/claude-code",
                },
                {
                  text: "Pool CLI",
                  link: "/zh/use-cases/pool-cli",
                },                
                {
                  text: "Paperclip",
                  link: "/zh/use-cases/paperclip",
                },
                {
                  text: "DeerFlow 2.0",
                  link: "/zh/use-cases/deerflow2",
                },
                // {
                //   text: "DeerFlow",
                //   link: "/zh/use-cases/deerflow",
                // },
                {
                  text: "TradingAgents",
                  link: "/zh/use-cases/tradingagents",
                },
                {
                  text: "NOFX",
                  link: "/zh/use-cases/nofx",
                },
              ],
            },
            {
              text: "AI 工作台",
              collapsed: true,
              items: [
                {
                  text: "Open WebUI",
                  collapsed: true,
                  link: "/zh/use-cases/openwebui",
                  items: [
                    {
                      text: "设置多用户访问",
                      link: "/zh/use-cases/openwebui-multiuser",
                    },
                    {
                      text: "配置语音交互",
                      link: "/zh/use-cases/openwebui-audio",
                    },
                    {
                      text: "启用网页搜索",
                      link: "/zh/use-cases/openwebui-search",
                    },
                    {
                      text: "在聊天中查询文档",
                      link: "/zh/use-cases/openwebui-knowledge",
                    },
                    {
                      text: "常见问题",
                      link: "/zh/use-cases/openwebui-common-issues",
                    },
                  ]
                },
                {
                  text: "LobeHub (LobeChat)",
                  link: "/zh/use-cases/lobechat",
                },
                {
                  text: "AnythingLLM",
                  link: "/zh/use-cases/anythingllm",
                },
                {
                  text: "Open Notebook",
                  link: "/zh/use-cases/open-notebook",
                },
                {
                  text: "Vane (Perplexica)",
                  link: "/zh/use-cases/perplexica",
                },
              ],
            },
            {
              text: "创意媒体",
              collapsed: true,
              items: [
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
                    {
                      text: "常见问题",
                      link: "/zh/use-cases/comfyui-common-issues",
                    },
                  ]
                },
                {
                  text: "ACE-Step 1.5",
                  link: "/zh/use-cases/ace-step-1.5",
                },
                /*{
                  text: "ACE-Step",
                  link: "/zh/use-cases/ace-step",
                },*/
                /*{
                  text: "Duix.Avatar",
                  link: "/zh/use-cases/duix-avatar",
                },*/
                // {
                //   text: "Stable Diffusion",
                //   link: "/zh/use-cases/stable-diffusion",
                // },
                {
                  text: "Whisper-WebUI",
                  link: "/zh/use-cases/whisper-webui",
                },
                {
                  text: "Speaches",
                  link: "/zh/use-cases/speaches",
                },
                {
                  text: "IndexTTS2",
                  link: "/zh/use-cases/indextts2",
                },
              ],
            },
            {
              text: "模型服务",
              collapsed: true,
              items: [
                {
                  text: "Ollama",
                  link: "/zh/use-cases/ollama",
                },
                {
                  text: "Bifrost",
                  link: "/zh/use-cases/bifrost",
                },
                {
                  text: "LiteLLM",
                  link: "/zh/use-cases/litellm",
                },
                {
                  text: "TensorZero",
                  link: "/zh/use-cases/tensorzero",
                },
                {
                  text: "LLMFit",
                  link: "/zh/use-cases/llmfit",
                },
                // {
                //   text: "Dify",
                //   link: "/zh/use-cases/dify",
                // },
              ],
            },
            {
              text: "具身智能",
              collapsed: true,
              items: [
                {
                  text: "Isaac Lab",
                  link: "/zh/use-cases/isaac-lab",
                },
              ],
            },
          ],
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
                  text: "Enable Intel integrated GPU passthrough",
                  link: "/zh/use-cases/windows-intel-gpu-passthrough",
                },
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
            {
              text: "Navidrome",
              link: "/zh/use-cases/navidrome",
            },
            {
              text: "TREK (NOMAD)",
              link: "/zh/use-cases/trek",
              collapsed: true,
              items: [
                {
                  text: "Advanced settings",
                  link: "/zh/use-cases/trek-advanced-settings",
                }
              ]
            },
            {
              text: "*Arrs",
              link: "/zh/use-cases/arrs",
              collapsed: true,
              items: [
                {
                  text: "升级 *Arrs",
                  link: "/zh/use-cases/arrs-upgrade",
                },
                {
                  text: "升级下载客户端",
                  link: "/zh/use-cases/download-clients-upgrade",
                },
              ]
            },
          ]
        },
        {
          text: "生产力",
          collapsed: false,
          items: [
            {
              text: "Plane",
              link: "/zh/use-cases/plane",
            },
            {
              text: "Excalidraw",
              link: "/zh/use-cases/excalidraw",
            },
            {
              text: "Karakeep",
              link: "/zh/use-cases/karakeep",
            },
            {
              text: "Obsidian LiveSync",
              link: "/zh/use-cases/obsidian-livesync",
            },
            {
              text: "Penpot",
              link: "/zh/use-cases/penpot",
            },
            {
              text: "RNA Sequencing",
              link: "/zh/use-cases/rna-sequencing",
            },
            // {
            //   text: "Stirling PDF",
            //   link: "/zh/use-cases/stirling-pdf",
            // },
          ],
        },
        {
          text: "开发者工具",
          collapsed: false,
          items: [
            {
              text: "JupyterHub",
              link: "/zh/use-cases/jupyterhub",
            },
          ],
        },
        {
          text: "实用工具",
          collapsed: false,
          items: [
            {
              text: "PDFMathTranslate",
              link: "/zh/use-cases/pdfmathtranslate",
            },
            {
              text: "Context7",
              link: "/zh/use-cases/context7",
            },
            {
              text: "Falco",
              link: "/zh/use-cases/falco",
            },
            {
              text: "Firecrawl",
              link: "/zh/use-cases/firecrawl",
            },
            {
              text: "FlareSolverr",
              link: "/zh/use-cases/flaresolverr",
            },
            {
              text: "Home Assistant",
              link: "/zh/use-cases/home-assistant",
            },
          ],
        },
      ],
    };
