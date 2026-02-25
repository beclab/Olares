import { defineConfig, type DefaultTheme } from "vitepress";

export const useCaseSidebar: DefaultTheme.Sidebar = {
    "/zh/use-cases/": [
        {
          text: "应用示例",
          link: "/zh/use-cases/",
          items: [
            {
              text: "Stable Diffusion",
              link: "/zh/use-cases/stable-diffusion",
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
            {
              text: "Open WebUI",
              link: "/zh/use-cases/openwebui",
            },
            {
              text: "Perplexica",
              link: "/zh/use-cases/perplexica",
            },
            {
              text: "Dify",
              link: "/zh/use-cases/dify",
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
            // {
            //   text: "Redroid",
            //   link: "/zh/use-cases/host-cloud-android",
            // },
          ],
        },
      ],
    }