import { defineConfig, type DefaultTheme } from "vitepress";

const side = {
  "/manual/": [
    {
      text: "What is Olares",
      link: "/manual/overview",
      items: [
        { text: "Compare Olares and NAS", link: "/manual/olares-vs-nas" },
        { text: "Help and support", link: "/manual/help/request-technical-support"}
          //  collapsed: true,
          //   items: [
            // { text: "FAQs", link: "/manual/help/faqs" },
            //    {
            //     text: "Request support",
          //     link: "/manual/help/request-technical-support",
          //  },
            //{
            //  text: "Troubleshooting Guide",
            //  link: "/manual/help/troubleshooting-guide",
            // },
      ],
    },
    {
      text: "Get started",
      collapsed: false,
      link: "/manual/get-started/",
      items: [
        // { text: "Quick start", link: "/manual/get-started/quick-start" },
        {
          text: "Create an Olares ID",
          link: "/manual/get-started/create-olares-id",
        },
        {
          text: "Install Olares",
          link: "/manual/get-started/install-olares",
          collapsed: true,
          items: [
            {
              text: "Linux",
              collapsed: true,
              items: [
                {
                  text: "Using ISO image (Recommended)",
                  link: "/manual/get-started/install-linux-iso",
                },
                {
                  text: "Using script",
                  link: "/manual/get-started/install-linux-script",
                },
                {
                  text: "Using Docker Compose",
                  link: "/manual/get-started/install-linux-docker",
                },
              ],
            },
            {
              text: "macOS",
              collapsed: true,
              items: [
                {
                  text: "Using script",
                  link: "/manual/get-started/install-mac-script",
                },
                {
                  text: "Using Docker image",
                  link: "/manual/get-started/install-mac-docker",
                },
              ],
            },
            {
              text: "Windows (WSL 2)",
              collapsed: true,
              items: [
                {
                  text: "Using script",
                  link: "/manual/get-started/install-windows-script",
                },
                /* {
                  text: "Using Docker image",
                  link: "/manual/get-started/install-windows-docker",
                },*/
              ],
            },
            {
              text: "PVE",
              collapsed: true,
              items: [
                {
                  text: "Using script",
                  link: "/manual/get-started/install-pve-script",
                },
                {
                  text: "Using ISO image",
                  link: "/manual/get-started/install-pve-iso",
                },
                { text: "LXC on PVE", link: "/manual/get-started/install-lxc" },
              ],
            },

            {
              text: "Raspberry Pi",
              link: "/manual/get-started/install-raspberry-pi",
            },
          ],
        },
        {
          text: "Back up mnemonics",
          link: "/manual/larepass/back-up-mnemonics"
        },
        {
          text: "Access Olares locally",
          link: "/manual/get-started/local-access",
        },
        {
          text: "What's next",
          link: "/manual/get-started/next-steps",
        },
      ],
    },
    {
      text: "LarePass",
      link: "/manual/larepass/",
      collapsed: true,
      items: [
        {
          text: "Manage accounts",
          collapsed: true,
          items: [
            {text: "Create accounts", link:"/manual/larepass/create-account"},
            {text: "Back up mnemonics", link: "/manual/larepass/back-up-mnemonics"},
            {text: "Manage integrations", link:"/manual/larepass/integrations"},
          ],
        },
        {text: "Use VPN", link:"/manual/larepass/private-network"},
        {
          text: "Manage device",
          collapsed: true,
          items: [
            {text: "Activate Olares", link:"/manual/larepass/activate-olares"},
            {text: "Manage Olares", link:"/manual/larepass/manage-olares"},
          ],
        },
        {text: "Manage files", link:"/manual/larepass/manage-files"},
         // collapsed: true,
          //items: [
          //  {text: "Common file operations", link:"/manual/larepass/manage-files"},
           // {text: "Sync and share", link:"/manual/larepass/sync-share"}
         // ]
       // },
        {
          text: "Manage passwords",
          collapsed: true,
          items: [
            {text: "Autofill passwords", link: "/manual/larepass/autofill"},
            {text: "Generate 2FA codes", link: "/manual/larepass/two-factor-verification"},
          ],
        },
        {
          text: "Manage knowledge",
          link: "/manual/larepass/manage-knowledge",
        },
      ],
    },
    {
      text: "Olares applications",
      collapsed: true,
      link: "/manual/olares/",
      items: [
        { text: "Desktop", link: "/manual/olares/desktop", },
        { text: "Market", link: "/manual/olares/market", },
        {
          text: "Files",
          collapsed: true,
          link: "/manual/olares/files/",
          items: [
            {
              text: "Basic file operations",
              link: "/manual/olares/files/add-edit-download",
            },
         //   {
        //      text: "Sync and share",
         //     link: "/manual/larepass/sync-share",
         //   },
            {
              text: "Mount SMB",
              link: "/manual/olares/files/mount-SMB",
            },
           {
             text: "Mount cloud storage",
             link: "/manual/olares/files/mount-cloud-storage",
           },
          ],
        },
        {
          text: "Vault",
          collapsed: true,
          link: "/manual/olares/vault/",
          items: [
            {
              text: "Manage vault items",
              link: "/manual/olares/vault/vault-items",
            },
            {
              text: "Manage shared vault",
              link: "/manual/olares/vault/share-vault-items",
            },
            {
              text: "Autofill",
              link: "/manual/larepass/autofill",
            },
            {
              text: "Generate 2FA codes",
              link: "/manual/larepass/two-factor-verification",
            },
          ],
        },
        {
          text: "Wise",
          collapsed: true,
          link: "/manual/olares/wise/",
          items: [
            {
              text: "Basic operations",
              link: "/manual/olares/wise/basics",
            },
            {
              text: "Get recommendation engine",
              link: "/manual/olares/wise/recommend",
            },
            {
              text: "Manage your feeds",
              link: "/manual/olares/wise/subscribe",
            },
            {
              text: "Organize your knowledge",
              link: "/manual/olares/wise/filter",
            },
          ],
        },
        {
          text: "Control Hub",
          collapsed: true,
          link: "/manual/olares/controlhub/",
          items: [
            {
              text: "Manage workloads",
              link: "/manual/olares/controlhub/manage-workload",
            },
            {
              text: "Manage containers",
              link: "/manual/olares/controlhub/manage-container",
            },
            {
              text: "Manage resource config",
              link: "/manual/olares/controlhub/manage-resource",
            },
            {
              text: "Manage middleware",
              link: "/manual/olares/controlhub/manage-middleware",
            },
            {
              text: "Access Terminal",
              link: "/manual/olares/controlhub/terminal",
            },
          ],
        },
        {
          text: "Settings",
          collapsed: true,
          link: "/manual/olares/settings/",
          items: [
            {
              text: "My Olares",
              link: "/manual/olares/settings/my-olares"
            },
            {
              text: "Manage users",
              collapsed: true,
              items: [
                {
                  text: "Roles and permissions",
                  link: "/manual/olares/settings/roles-permissions",
                },
                {
                  text: "Create and manage users",
                  link: "/manual/olares/settings/manage-team",
                }
              ],
            },
            {
              text: "Manage applications",
              collapsed: true,
              items: [
                {
                  text: "Manage app entrance",
                  link: "/manual/olares/settings/manage-entrance",
                },
                {
                  text: "Customize app domains",
                  link: "/manual/olares/settings/custom-app-domain",
                },
                {
                  text: "Manage app variables",
                  link: "/manual/olares/settings/manage-app-env",
                },
              ],
              },
            {
              text: "Manage integrations",
              link:"/manual/olares/settings/integrations",
             },
             {
              text: "Customize appearance",
              link:"/manual/olares/settings/language-appearance",
             },
            {text: "Manage VPN", link: "/manual/olares/settings/remote-access",},
            {
              text: "Configure network",
              collapsed: true,
              items: [
                {
                  text: "Change revere proxy",
                  link: "/manual/olares/settings/change-frp",
                },
                {
                  text: "Set up hosts file",
                  link:"/manual/olares/settings/set-up-hosts",
                },
              ],
             },
            {text: "Manage GPU", link: "/manual/olares/settings/gpu-resource"},
            {text: "Set video playback", link: "/manual/olares/settings/video"},
            {
              text: "Backup and restore",
              collapsed: true,
              items: [
                {text: "Backup", link: "/manual/olares/settings/backup"},
                {text: "Restore", link: "/manual/olares/settings/restore"},
              ],
            },
            {text: "Developer resources", link: "/manual/olares/settings/developer"},
            ]
          },
          {text: "Dashboard", link: "/manual/olares/resources-usage"},
          {text: "Profile", link: "/manual/olares/profile"},
        ],
    },
      {
        text: "Best practices",
        link: "/manual/best-practices/",
        collapsed: true,
        items: [
          {
            text: "Set up custom domain",
            link: "/manual/best-practices/set-custom-domain",
          },
          {
            text: "Manage knowledge with Wise",
            link: "/manual/best-practices/organize-content",
          },
          {
            text: "Install a multi-node Olares cluster",
            link: "/manual/best-practices/install-olares-multi-node",
          },
          {
            text: "Install Olares on PVE with GPU Passthrough",
            link: "/manual/best-practices/install-olares-gpu-passthrough",
          },
          {
            text: "Expand storage in Olares",
          link: "/manual/best-practices/expand-storage-in-olares",
        },
        ],
      },
    { text: "Glossary", link: "/manual/glossary" },
  ],
  "/space/": [
    {
      text: "Olares Space",
      link: "/space/",
      items: [
        {
          text: "Manage accounts",
          link: "/space/manage-accounts",
        },
        {
          text: "Host Olares",
          collapsed: true,
          items: [
            {
              text: "Create Olares",
              link: "/space/create-olares",
            },
            {
              text: "Manage Olares",
              link: "/space/manage-olares",
            },
          ],
        },
        {
          text: "Host domains",
          collapsed: true,
          items: [
            {
              text: "Set up a custom domain",
              link: "/space/host-domain",
            },
            {
              text: "Manage a domain",
              link: "/space/manage-domain",
            },
          ],
        },
        {
          text: "Back up and restore",
          link: "/space/backup-restore",
        },
        { text: "Billing", link: "/space/billing" },
      ],
    },
  ],
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
        link: "/use-cases/stream-game",
      },
      {
        text: "Redroid",
        link: "/use-cases/host-cloud-android",
      },
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
        text: "LobeChat",
        link: "/use-cases/lobechat",
      },   
    ],
  },
 ],
  "/developer/": [
    {
      text: "Concepts",
      link: "/developer/concepts/",
      items: [
        { text: "Olares architecture", link: "/developer/concepts/system-architecture" },
        { text: "Olares ID",
          link: "/developer/concepts/olares-id",
          collapsed: true,
          items: [
            {
              text: "Decentralized ID",
              link: "/developer/concepts/did",
            },
            {
              text: "Blockchain Registry",
              link: "/developer/concepts/registry",
            },
            {
              text: "Verifiable Credential",
              link: "/developer/concepts/vc",
            },
            {
              text: "Autonomous Reputation",
              link: "/developer/concepts/reputation",
            },
            //  {
            //   text: "Self-Sovereign Network",
            //    link: "/developer/concepts/self-sovereign-network",
            //   },
            {
              text: "Identity Wallet",
              link: "/developer/concepts/wallet",
            },
          ],
        },
        { text: "Account", link: "/developer/concepts/account" },
        { text: "Application", link: "/developer/concepts/application" },
        { text: "Network", link: "/developer/concepts/network" },
        { text: "Data", link: "/developer/concepts/data" },
        { text: "Secrets", link: "/developer/concepts/secrets" },
      ],
    },
      {
      text: "Installation deep-dive",
      link: "/developer/install/",
      items: [
        {
          text: "Installation architecture",
          link: "/developer/install/installation-overview",
        },
        {
          text: "Installation process",
          link: "/developer/install/installation-process",
        },
        {
          text: "Olares Home",
          link: "/developer/install/olares-home",
        },
        {
          text: "Environment variables",
          link: "/developer/install/environment-variables",
        },
        {
          text: "Olares CLI",
          link: "/developer/install/cli/olares-cli",
          collapsed: true,
          items: [
            { text: "gpu", link: "/developer/install/cli/gpu" },
            { text: "osinfo", link: "/developer/install/cli/osinfo" },
            { text: "node", link: "/developer/install/cli/node" },
            {
              text: "backups",
              link: "/developer/install/cli/backups",
              collapsed: true,
              items: [
                  {text: "download", link: "/developer/install/cli/backups-download"},
                  {text: "region", link: "/developer/install/cli/backups-region"},
                  {text: "backup", link: "/developer/install/cli/backups-backup"},
                  {text: "restore", link: "/developer/install/cli/backups-restore"},
                  {text: "snapshots", link: "/developer/install/cli/backups-snapshots"},
                  ],
            },
            {
              text: "change-ip",
              link: "/developer/install/cli/change-ip",
            },
            {
              text: "download",
              link: "/developer/install/cli/download",
            },
            { text: "info", link: "/developer/install/cli/info" },
            {
              text: "install",
              link: "/developer/install/cli/install",
            },
            {
              text: "user activate",
              link: "/developer/install/cli/user-activate",
            },
            {
              text: "logs",
              link: "/developer/install/cli/logs",
            },
            {
              text: "precheck",
              link: "/developer/install/cli/precheck",
            },
            {
              text: "prepare",
              link: "/developer/install/cli/prepare",
            },
            {
              text: "release",
              link: "/developer/install/cli/release",
            },
            {
              text: "start",
              link: "/developer/install/cli/start",
            },
            {
              text: "stop",
              link: "/developer/install/cli/stop",
            },
            {
              text: "uninstall",
              link: "/developer/install/cli/uninstall",
            },
          ],
        },
            {
            text: "Olares versioning",
            link: "/developer/install/versioning",
          },
    ],
  },
    {
      text: "Develop Olares apps",
      link: "/developer/develop/",
      items: [
        {
          text: "Develop with Studio",
          collapsed: true,
          link: "/developer/develop/tutorial/",
          items: [
            {
              text: "Deploy an app",
              link: "/developer/develop/tutorial/deploy",
            },
            {
              text: "Develop in a dev container",
              link: "/developer/develop/tutorial/develop",
            },
            {
              text: "Package and upload",
              link: "/developer/develop/tutorial/package-upload",
            },
            {
              text: "Add app assets",
              link: "/developer/develop/tutorial/assets",
            },
          ],
        },
        {
          text: "Application package",
          collapsed: true,
          items: [
            {
              text: "Application chart",
              link: "/developer/develop/package/chart",
            },
            {
              text: "OlaresManifest",
              link: "/developer/develop/package/manifest",
            },
            /*{
              text: "Recommendation",
              link: "/developer/develop/package/recommend",
            },*/
            {
              text: "Helm extension",
              link: "/developer/develop/package/extension",
            },
          ],
        },
        //  {
        //   text: "Advanced",
        //   collapsed: true,
        //  items: [
        //    {
        //       text: "terminus-info",
      //      link: "/developer/develop/advanced/terminus-info",
        //     },
        //     {
        //       text: "Service provider",
      //      link: "/developer/develop/advanced/provider",
        //      },
        //     {
        //        text: "AI",
      //        link: "/developer/develop/advanced/ai",
        //       },
        //      { text: "Cookie", link: "/developer/develop/advanced/cookie" },
        //     { text: "Database", link: "/developer/develop/advanced/database" },
        //     {
        //       text: "Account",
      //       link: "/developer/develop/advanced/account",
        //      },
        //     {
        //       text: "Market",
      //       link: "/developer/develop/advanced/market",
        //      },
        //       {
        //       text: "Websocket",
      //       link: "/developer/develop/advanced/websocket",
        //     },
        //      {
        //       text: "File upload",
      //       link: "/developer/develop/advanced/file-upload",
        //      },
        //     {
        //       text: "Secret",
      //       link: "/developer/develop/advanced/secret",
        //      },
        //      {
        //        text: "Kubesphere",
      //      link: "/developer/develop/advanced/kubesphere",
        //      },
      //   ],
        //   },
        {
          text: "Submit application",
          collapsed: true,
          link: "/developer/develop/submit/",
        },
      ],
    },
    {
      text: "Contribute to Olares",
      items: [
        {
          text: "Develop system app",
          collapsed: true,
          items: [
            {
              text: "Overview",
              link: "/developer/contribute/system-app/overview",
            },
            {
              text: "Configure deployment",
              link: "/developer/contribute/system-app/deployment",
            },
            {
              text: "Configure permissions",
              link: "/developer/contribute/system-app/olares-manifest",
            },
            {
              text: "Install",
              link: "/developer/contribute/system-app/install",
            },
            {
              text: "Other",
              link: "/developer/contribute/system-app/other",
            },
          ],
        },
      {
        text: "Develop protocols",
        collapsed: true,
        items: [
          {
            text: "Contract",
            link: "/developer/contribute/olares-id/contract/contract",
            collapsed: true,
            items: [
              {
                text: "Architecture",
                link: "/developer/contribute/olares-id/contract/architecture",
              },
              {
                text: "DID",
                collapsed: true,
                items: [
                  {
                    text: "Design",
                    link: "/developer/contribute/olares-id/contract/did/design",
                  },
                  {
                    text: "Official Taggers",
                    link: "/developer/contribute/olares-id/contract/did/official-taggers",
                  },
                  {
                    text: "Release History",
                    link: "/developer/contribute/olares-id/contract/did/release-history",
                  },
                  {
                    text: "FAQ",
                    link: "/developer/contribute/olares-id/contract/did/faq",
                  },
                ],
              },
              {
                text: "Reputation",
                link: "/developer/contribute/olares-id/contract/contract-reputation",
              },
              {
                text: "Manage",
                collapsed: true,
                items: [
                  {
                    text: "Contract",
                    link: "/developer/contribute/olares-id/contract/manage/contract",
                  },
                  {
                    text: "SDK",
                    link: "/developer/contribute/olares-id/contract/manage/sdk",
                  },
                  {
                    text: "Environment",
                    link: "/developer/contribute/olares-id/contract/manage/environment",
                  },
                ],
              },
            ],
          },
          {
            text: "Verifiable Credential",
            link: "/developer/contribute/olares-id/verifiable-credential/overview",
            collapsed: true,
            items: [
              {
                text: "Issuer",
                link: "/developer/contribute/olares-id/verifiable-credential/issuer",
              },
              {
                text: "Verifer",
                link: "/developer/contribute/olares-id/verifiable-credential/verifer",
              },
              {
                text: "Olares",
                link: "/developer/contribute/olares-id/verifiable-credential/olares",
              },
            ],
          },
        ],
       },
      ],
    },
  ],
};

export const en = defineConfig({
  lang: "en",
  themeConfig: {
    //logo: "/icon.png",
    socialLinks: [{ icon: "github", link: "https://github.com/beclab/olares" }],

    nav: [
      { text: "Olares", link: "/manual/overview" },
      { text: "Olares Space", link: "/space/" },
      { text: "Use Cases", link: "/use-cases/" },
      { text: "Developer Guide", link: "/developer/concepts/" },
    ],

    sidebar: side,
  },
});
