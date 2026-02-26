import { defineConfig, type DefaultTheme } from "vitepress";
import { oneSidebar } from './one.en.ts';
import { useCaseSidebar } from './usecase.en.ts';
import { developerSidebar } from './developer.en.ts';
const side = {
"/manual/": [
    {
      text: "What is Olares",
      link: "/manual/overview",
      items: [
        // { text: "Compare Olares and NAS", link: "/manual/olares-vs-nas" },
        {
          text: "FAQs",
          //           link: "/manual/help/faqs",
          collapsed: true,
          items: [
            {
              text: "Olares FAQs",
              link: "/manual/help/olares",
            },
            {
              text: "Setup & access FAQs",
              link: "/manual/help/installation",
            },
            {
              text: "Usage FAQs",
              link: "/manual/help/usage",
            },
            //             {
            //               text: "Request support",
            //               link: "/manual/help/request-technical-support",
            //             },

          ],
        },
        {
          text: "Troubleshooting",
          collapsed: true,
          items: [
            {
              text: "Insufficient memory or memory not freed",
              link: "/manual/help/ts-free-memory",
            },
          ],
        },
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
          text: "Access Olares securely",
          link: "/manual/get-started/local-access",
        },
        {
          text: "What's next",
          link: "/manual/get-started/next-steps",
        },
      ],
    },
    {
      text: "Olares applications",
      link: "/manual/olares/",
      items: [
        { text: "Desktop", link: "/manual/olares/desktop", },
        {
          text: "Market",
          collapsed: true,
          items: [
            {
              text: "Basic operations",
              link: "/manual/olares/market/market",
            },
            {
              text: "Clone applications",
              link: "/manual/olares/market/clone-apps",
            },
            {
              text: "Manage paid applications",
              link: "/manual/olares/market/purchase-paid-apps",
            },
          ],
        },
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
              text: "Share files",
              link: "/manual/olares/files/share-files",
            },
            {
              text: "Sync files to local",
              link: "/manual/olares/files/sync-files",
            },
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
            /*{
              text: "Get recommendation engine",
              link: "/manual/olares/wise/recommend",
            },*/
            {
              text: "Manage your feeds",
              link: "/manual/olares/wise/subscribe",
            },
            {
              text: "Manage cookies",
              link: "/manual/olares/wise/manage-cookies",
            },
            {
              text: "Organize with filters",
              link: "/manual/olares/wise/filter",
              collapsed: true,
              items: [
                {
                  text: "Filter syntax",
                  link: "/manual/olares/wise/filter-syntax-guide",
                },
                {
                  text: "Filter example",
                  link: "/manual/olares/wise/filter-examples",
                },
              ],
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
              link: "/manual/olares/settings/integrations",
            },
            {
              text: "Customize appearance",
              link: "/manual/olares/settings/language-appearance",
            },
            { text: "Manage VPN", link: "/manual/olares/settings/remote-access", },
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
                  link: "/manual/olares/settings/set-up-hosts",
                },
              ],
            },
            { text: "Manage GPU", link: "/manual/olares/settings/gpu-resource" },
            { text: "Set video playback", link: "/manual/olares/settings/video" },
            { text: "Manage search rules", link: "/manual/olares/settings/search" },
            {
              text: "Backup and restore",
              collapsed: true,
              items: [
                { text: "Backup", link: "/manual/olares/settings/backup" },
                { text: "Restore", link: "/manual/olares/settings/restore" },
              ],
            },
            { text: "Developer resources", link: "/manual/olares/settings/developer" },
          ]
        },
        { text: "Dashboard", link: "/manual/olares/resources-usage" },
        { text: "Profile", link: "/manual/olares/profile" },
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
            { text: "Create accounts", link: "/manual/larepass/create-account" },
            { text: "Back up mnemonics", link: "/manual/larepass/back-up-mnemonics" },
            { text: "Manage integrations", link: "/manual/larepass/integrations" },
          ],
        },
        { text: "Use VPN", link: "/manual/larepass/private-network" },
        {
          text: "Manage device",
          collapsed: true,
          items: [
            { text: "Activate Olares", link: "/manual/larepass/activate-olares" },
            { text: "Manage Olares", link: "/manual/larepass/manage-olares" },
          ],
        },
        { text: "Manage files", link: "/manual/larepass/manage-files" },
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
            { text: "Autofill passwords", link: "/manual/larepass/autofill" },
            { text: "Generate 2FA codes", link: "/manual/larepass/two-factor-verification" },
          ],
        },
        /*{
          text: "Manage knowledge",
          link: "/manual/larepass/manage-knowledge",
        },*/
      ],
    },
    {
      text: "Olares Space",
      link: "/manual/space/",
      collapsed: true,
      items: [
        {
          text: "Manage accounts",
          link: "/manual/space/manage-accounts",
        },
        {
          text: "Host Olares",
          collapsed: true,
          items: [
            {
              text: "Create Olares",
              link: "/manual/space/create-olares",
            },
            {
              text: "Manage Olares",
              link: "/manual/space/manage-olares",
            },
          ],
        },
        {
          text: "Host domains",
          collapsed: true,
          items: [
            {
              text: "Set up a custom domain",
              link: "/manual/space/host-domain",
            },
            {
              text: "Manage a domain",
              link: "/manual/space/manage-domain",
            },
          ],
        },
        {
          text: "Back up and restore",
          link: "/manual/space/backup-restore",
        },
        { text: "Billing", link: "/manual/space/billing" },
      ],
    },
    {
      text: "Tutorials",
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
        {
          text: "Access Olares locally",
          link: "/manual/best-practices/local-access",
        },
      ],
    },
    { text: "Glossary", link: "/manual/glossary" },
  ],
};

export const en = defineConfig({
  lang: "en",
  themeConfig: {
    //logo: "/icon.png",
    socialLinks: [{ icon: "github", link: "https://github.com/beclab/olares" }],

    nav: [
      { text: "Olares OS", link: "/manual/overview" },
      { text: "Olares One", link: "/one/" },
      { text: "Use cases", link: "/use-cases/" },
      { text: "Developer guide", link: "/developer/concepts/" },
    ],

    sidebar: {
      ...side,
      ...oneSidebar,
      ...useCaseSidebar,
      ...developerSidebar,
    },
  },
});