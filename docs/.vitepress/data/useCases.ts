// Single source of truth for the use-cases gallery on /use-cases/.
// Each entry maps an application to its category and a short description.
// The gallery (UseCaseGallery.vue) is locale-aware: on the Chinese site it
// uses titleZh/descriptionZh when present and falls back to the English fields.

export interface UseCase {
  title: string;
  titleZh?: string;
  link: string;
  category: string;
  description: string;
  descriptionZh?: string;
}

// Display order of the category filter pills (canonical English keys).
export const useCaseCategories: string[] = [
  "AI agent",
  "AI workbench",
  "Creative media",
  "Model services",
  "Embodied AI",
  "Virtual machine",
  "Entertainment",
  "Productivity",
  "Developer tools",
  "Utilities",
];

// Localized category labels, keyed by the canonical English category above.
export const categoryLabelsZh: Record<string, string> = {
  "AI agent": "AI Agent",
  "AI workbench": "AI 工作台",
  "Creative media": "创意媒体",
  "Model services": "模型服务",
  "Embodied AI": "具身智能",
  "Virtual machine": "虚拟机",
  Entertainment: "娱乐",
  Productivity: "生产力",
  "Developer tools": "开发者工具",
  Utilities: "实用工具",
};

export const useCases: UseCase[] = [
  { title: "OpenClaw", link: "/use-cases/openclaw", category: "AI agent", description: "Personal AI assistant that integrates with Discord and runs tasks", descriptionZh: "可集成 Discord 并执行任务的个人 AI 助手" },
  { title: "NemoClaw", link: "/use-cases/nemoclaw", category: "AI agent", description: "Always-on AI agent backed by a local LLM", descriptionZh: "由本地 LLM 驱动的常驻 AI 智能体" },
  { title: "Hermes Agent", link: "/use-cases/hermes", category: "AI agent", description: "Self-directed agent with Olares skills and a Gateway API", descriptionZh: "具备 Olares 技能和 Gateway API 的自主智能体" },
  { title: "OpenCode", link: "/use-cases/opencode", category: "AI agent", description: "AI coding agent driven by natural language", descriptionZh: "通过自然语言驱动的 AI 编程智能体" },
  { title: "Claude Code", link: "/use-cases/claude-code", category: "AI agent", description: "Write, test, and manage code through natural language", descriptionZh: "通过自然语言编写、测试和管理代码" },
  { title: "Pool CLI", link: "/use-cases/pool-cli", category: "AI agent", description: "Read code, run commands, and edit files via natural language", descriptionZh: "用自然语言读取代码、运行命令并编辑文件" },
  { title: "Paperclip", link: "/use-cases/paperclip", category: "AI agent", description: "Coordinate multiple AI agents on the same tasks", descriptionZh: "协调多个 AI 智能体协作处理同一任务" },
  { title: "DeerFlow 2.0", link: "/use-cases/deerflow2", category: "AI agent", description: "AI-powered deep research with a local model", descriptionZh: "基于本地模型的 AI 深度研究" },
  { title: "TradingAgents", link: "/use-cases/tradingagents", category: "AI agent", description: "Multi-agent financial market analysis", descriptionZh: "多智能体金融市场分析" },
  { title: "NOFX", link: "/use-cases/nofx", category: "AI agent", description: "Autonomous AI trading agent", descriptionZh: "自主 AI 交易智能体" },
  { title: "Open WebUI", link: "/use-cases/openwebui", category: "AI workbench", description: "Full-featured web chat for local LLMs", descriptionZh: "面向本地 LLM 的全功能 Web 聊天界面" },
  { title: "LobeHub (LobeChat)", link: "/use-cases/lobechat", category: "AI workbench", description: "Build local custom AI assistants", descriptionZh: "构建本地自定义 AI 助手" },
  { title: "AnythingLLM", link: "/use-cases/anythingllm", category: "AI workbench", description: "Local knowledge base with RAG over your documents", descriptionZh: "基于文档 RAG 的本地知识库" },
  { title: "Open Notebook", link: "/use-cases/open-notebook", category: "AI workbench", description: "Research notebook with AI insights and podcasts", descriptionZh: "支持 AI 洞察和播客的研究笔记本" },
  { title: "Vane (Perplexica)", link: "/use-cases/perplexica", category: "AI workbench", description: "Private, AI-powered search and Q&A", descriptionZh: "注重隐私的 AI 搜索与问答" },
  { title: "Dify", link: "/use-cases/dify-upgrade", category: "AI workbench", description: "Build AI apps and agents with a shared workspace", descriptionZh: "在共享工作空间中构建 AI 应用与智能体" },
  { title: "ComfyUI", link: "/use-cases/comfyui", category: "Creative media", description: "Node-based AI image generation", descriptionZh: "节点式 AI 图像生成" },
  { title: "ACE-Step 1.5", link: "/use-cases/ace-step-1.5", category: "Creative media", description: "Generate music from prompts or lyrics", descriptionZh: "根据提示词或歌词生成音乐" },
  { title: "Whisper-WebUI", link: "/use-cases/whisper-webui", category: "Creative media", description: "Speech-to-text, subtitles, and translation", descriptionZh: "语音转文字、字幕生成与翻译" },
  { title: "Speaches", link: "/use-cases/speaches", category: "Creative media", description: "Speech-to-text, text-to-speech, and voice chat", descriptionZh: "语音转文字、文字转语音与语音对话" },
  { title: "IndexTTS2", link: "/use-cases/indextts2", category: "Creative media", description: "Text-to-speech with zero-shot voice cloning", descriptionZh: "支持零样本声音克隆的文字转语音" },
  { title: "Ollama", link: "/use-cases/ollama", category: "Model services", description: "Download and run local AI models", descriptionZh: "下载并运行本地 AI 模型" },
  { title: "Bifrost", link: "/use-cases/bifrost", category: "Model services", description: "AI gateway that aggregates models behind one endpoint", descriptionZh: "将多个模型聚合到单一端点的 AI 网关" },
  { title: "LiteLLM", link: "/use-cases/litellm", category: "Model services", description: "Unify model providers behind one OpenAI-compatible API", descriptionZh: "用统一的 OpenAI 兼容 API 整合多家模型提供商" },
  { title: "TensorZero", link: "/use-cases/tensorzero", category: "Model services", description: "AI model gateway and observability platform", descriptionZh: "AI 模型网关与可观测性平台" },
  { title: "LLMFit", link: "/use-cases/llmfit", category: "Model services", description: "Benchmark your hardware to find the best models", descriptionZh: "对硬件进行基准测试以找到最合适的模型" },
  { title: "Engine Base apps", link: "/use-cases/llm-base-apps", category: "Model services", description: "Self-host local LLMs by cloning base apps for different inference engines", descriptionZh: "通过克隆不同推理引擎的基座应用来自托管本地大语言模型" },
  { title: "Xinference", link: "/use-cases/xinference", category: "Model services", description: "Deploy and serve models on Olares", descriptionZh: "在 Olares 上部署和提供模型服务" },
  { title: "Isaac Lab", link: "/use-cases/isaac-lab", category: "Embodied AI", description: "GPU-accelerated robot simulation training", descriptionZh: "GPU 加速的机器人仿真训练" },
  { title: "macOS", link: "/use-cases/macos", category: "Virtual machine", description: "Run a macOS VM with browser-based VNC", descriptionZh: "通过浏览器 VNC 运行 macOS 虚拟机" },
  { title: "Windows", link: "/use-cases/windows", category: "Virtual machine", description: "Run a Windows VM with VNC or Remote Desktop", descriptionZh: "通过 VNC 或远程桌面运行 Windows 虚拟机" },
  { title: "Immich", link: "/use-cases/immich", category: "Entertainment", description: "Back up and organize photos with AI search", descriptionZh: "备份整理照片，支持 AI 搜索" },
  { title: "Jellyfin", link: "/use-cases/jellyfin", category: "Entertainment", description: "Private media streaming server", descriptionZh: "私人流媒体服务器" },
  { title: "Steam", link: "/use-cases/steam-direct-play", category: "Entertainment", description: "Play Steam games locally or stream them", descriptionZh: "本地游玩 Steam 游戏或串流到其他设备" },
  { title: "Komga", link: "/use-cases/komga", category: "Entertainment", description: "Self-hosted comic and manga library", descriptionZh: "自托管的漫画书库" },
  { title: "Navidrome", link: "/use-cases/navidrome", category: "Entertainment", description: "Private music streaming server", descriptionZh: "私人音乐流媒体服务器" },
  { title: "TREK (NOMAD)", link: "/use-cases/trek", category: "Entertainment", description: "Plan trips collaboratively", descriptionZh: "协作规划行程" },
  { title: "*Arrs", link: "/use-cases/arrs", category: "Entertainment", description: "Automated media management with Sonarr, Radarr, and more", descriptionZh: "使用 Sonarr、Radarr 等实现媒体自动化管理" },
  { title: "Excalidraw", link: "/use-cases/excalidraw", category: "Productivity", description: "Hand-drawn style diagrams and whiteboarding", descriptionZh: "手绘风格的图表与白板" },
  { title: "Karakeep", link: "/use-cases/karakeep", category: "Productivity", description: "Save and organize bookmarks, notes, and PDFs", descriptionZh: "保存并整理书签、笔记与 PDF" },
  { title: "Obsidian LiveSync", link: "/use-cases/obsidian-livesync", category: "Productivity", description: "Sync Markdown notes across devices", descriptionZh: "跨设备同步 Markdown 笔记" },
  { title: "Penpot", link: "/use-cases/penpot", category: "Productivity", description: "Self-hosted design workspace, with Cursor via MCP", descriptionZh: "自托管设计工作空间，可通过 MCP 接入 Cursor" },
  { title: "Plane", link: "/use-cases/plane", category: "Productivity", description: "Self-hosted project management", descriptionZh: "自托管的项目管理" },
  { title: "RNA Sequencing", link: "/use-cases/rna-sequencing", category: "Productivity", description: "GPU-accelerated single-cell RNA analysis", descriptionZh: "GPU 加速的单细胞 RNA 测序分析" },
  { title: "OnlyOffice", link: "/use-cases/onlyoffice", category: "Productivity", description: "Self-hosted document editing and collaboration", descriptionZh: "自托管的文档编辑与协作" },
  { title: "JupyterHub", link: "/use-cases/jupyterhub", category: "Developer tools", description: "Multi-user Jupyter notebook environment", descriptionZh: "多用户 Jupyter Notebook 环境" },
  { title: "PDFMathTranslate", link: "/use-cases/pdfmathtranslate", category: "Utilities", description: "Translate scientific PDFs while preserving layout", descriptionZh: "翻译科学 PDF 并保留排版" },
  { title: "Context7", link: "/use-cases/context7", category: "Utilities", description: "Give AI coding assistants up-to-date docs via MCP", descriptionZh: "通过 MCP 为 AI 编程助手提供最新文档" },
  { title: "Falco", link: "/use-cases/falco", category: "Utilities", description: "Runtime security monitoring for hosts and containers", descriptionZh: "面向主机和容器的运行时安全监控" },
  { title: "Firecrawl", link: "/use-cases/firecrawl", category: "Utilities", description: "Web page loader and scraper for apps", descriptionZh: "面向应用的网页加载与抓取工具" },
  { title: "FlareSolverr", link: "/use-cases/flaresolverr", category: "Utilities", description: "Bypass Cloudflare protection in Prowlarr", descriptionZh: "在 Prowlarr 中绕过 Cloudflare 防护" },
  { title: "SearXNG", link: "/use-cases/searxng", category: "Utilities", description: "Privacy-focused meta search engine", descriptionZh: "注重隐私的元搜索引擎" },
  { title: "Home Assistant", link: "/use-cases/home-assistant", category: "Utilities", description: "Build your smart home hub with local device discovery and automation", descriptionZh: "搭建智能家居中枢，支持本地设备发现与自动化" },
];
