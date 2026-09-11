import { defineConfig, type DefaultTheme } from "vitepress";
import { oneSidebar } from './one.en.ts';
import { useCaseSidebar } from './usecase.en.ts';
import { developerSidebar } from './developer.en.ts';
const side = {
  "/manual/": [
    {
      text: "Overview",
      link: "/manual/overview",
      items: [
        {
          text: "Update notes",
          collapsed: true,
          items: [
            {
              text: "Olares 1.12.6",
              link: "/manual/update-guides/1.12.6",
            },
          ],
        },
        {
          text: "What's new in docs",
          link: "/manual/release-notes",
        },
        {
          text: "FAQs",
          link: "/manual/help/faqs",
          collapsed: true,
          items: [
            {
              text: "About Olares",
              link: "/manual/help/olares",
            },
            {
              text: "Installation and activation",
              link: "/manual/help/installation",
            },
            {
              text: "Using Olares",
              link: "/manual/help/usage",
            },
          ],
        },
        {
          text: "Get support",
          link: "/manual/help/request-technical-support",
        },
      ],
    },
    {
      text: "Get started",
      collapsed: false,
      link: "/manual/get-started/",
      items: [
        {
          text: "Download LarePass",
          link: "/manual/larepass/",
        },
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
              text: "DGX Spark",
              collapsed: true,
              items: [
                {
                  text: "Using script (Recommended)",
                  link: "/manual/get-started/install-spark-script",
                },
                {
                  text: "Using ISO image",
                  link: "/manual/get-started/install-spark-iso",
                },
              ],
            },
            /* {
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
            }, */
            /* {
              text: "Windows (WSL 2)",
              collapsed: true,
              items: [
                {
                  text: "Using script",
                  link: "/manual/get-started/install-windows-script",
                },
              ],
            }, */
            /* {
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
            }, */
            /* {
              text: "Raspberry Pi",
              link: "/manual/get-started/install-raspberry-pi",
            }, */
          ],
        },
        {
          text: "Join an existing Olares as a member",
          link: "/manual/get-started/join-olares",
        },
        {
          text: "Get familiar with Desktop",
          link: "/manual/olares/desktop",
        },
        {
          text: "What's next",
          link: "/manual/get-started/next-steps",
        },
      ],
    },
    {
      text: "Accounts and access",
      collapsed: false,
      items: [
        {
          text: "Olares ID",
          collapsed: true,
          items: [
            {
              text: "Manage Olares IDs in LarePass",
              link: "/manual/larepass/manage-accounts",
            },
            {
              text: "Change your password and manage signed-in devices",
              link: "/manual/password-and-devices",
            },
            {
              text: "Set up a custom-domain Olares ID",
              link: "/manual/best-practices/set-custom-domain",
            },
            {
              text: "Manage your Olares Space account and billing",
              link: "/manual/space/manage-accounts",
            },
            {
              text: "Back up your mnemonic phrase",
              link: "/manual/larepass/back-up-mnemonics",
            },
          ],
        },
        {
          text: "Team",
          collapsed: true,
          items: [
            {
              text: "Create and manage team members",
              link: "/manual/olares/settings/manage-team",
            },
            {
              text: "Roles and permissions",
              link: "/manual/olares/settings/roles-permissions",
            },
          ],
        },
        {
          text: "Access Olares",
          collapsed: true,
          items: [
            {
              text: "Access Olares on your local network",
              link: "/manual/best-practices/local-access",
            },
            {
              text: "Access Olares remotely with LarePass VPN",
              link: "/manual/larepass/private-network",
            },
            {
              text: "Connect to Olares via SSH",
              link: "/manual/access-olares-terminal",
            },
          ],
        },
        {
          text: "Passwords",
          collapsed: true,
          items: [
            {
              text: "Create and manage Vault items",
              link: "/manual/olares/vault/vault-items",
            },
            {
              text: "Share Vault items securely",
              link: "/manual/olares/vault/share-vault-items",
            },
            {
              text: "Autofill passwords with LarePass",
              link: "/manual/larepass/autofill",
            },
            {
              text: "Generate two-factor authentication codes",
              link: "/manual/larepass/two-factor-verification",
            },
          ],
        },
        {
          text: "Personal settings",
          collapsed: true,
          items: [
            {
              text: "Change language and appearance",
              link: "/manual/olares/settings/language-appearance",
            },
            {
              text: "Create an Olares profile",
              link: "/manual/olares/profile",
            },
          ],
        },
      ],
    },
    {
      text: "Apps",
      collapsed: false,
      items: [
        {
          text: "Install, update, and remove apps",
          link: "/manual/olares/market/market",
        },
        {
          text: "Manually update an app image",
          link: "/manual/update-app-image",
        },
        {
          text: "Clone an app",
          link: "/manual/olares/market/clone-apps",
        },
        {
          text: "Configure environment variables for an app",
          link: "/manual/olares/settings/manage-app-env",
        },
        {
          text: "Set system environment variables",
          link: "/manual/manage-system-env",
        },
        {
          text: "Allocate accelerator resources to an app",
          link: "/manual/olares/settings/gpu-resource",
        },
        {
          text: "Change an app's web address",
          link: "/manual/olares/settings/custom-app-domain",
        },
        {
          text: "Make an app available on the local network",
          link: "/manual/olares/settings/overlay-gateway",
        },
        {
          text: "Control who can access an app",
          link: "/manual/olares/settings/manage-entrance",
        },
        {
          text: "Connect an AI app to a model service",
          link: "/manual/best-practices/connect-ai-apps",
        },
        {
          text: "Migrate legacy shared applications",
          link: "/manual/migrate-shared-apps",
        },
        {
          text: "About shared applications",
          link: "/manual/olares/market/shared-apps",
        },
      ],
    },
    {
      text: "Files and storage",
      link: "/manual/olares/files/",
      collapsed: false,
      items: [
        {
          text: "Work with files",
          collapsed: true,
          items: [
            {
              text: "Upload, edit, and download files",
              link: "/manual/olares/files/add-edit-download",
            },
            {
              text: "Compress and extract files",
              link: "/manual/olares/files/compress-extract-files",
            },
            {
              text: "Share files",
              link: "/manual/olares/files/share-files",
            },
            {
              text: "Sync files with a computer",
              link: "/manual/olares/files/sync-files",
            },
            {
              text: "Configure file search",
              link: "/manual/olares/settings/search",
            },
            {
              text: "Manage shared AI model files",
              link: "/manual/olares/files/files-common",
            },
          ],
        },
        {
          text: "Connect external storage",
          collapsed: true,
          items: [
            {
              text: "Use a USB drive",
              link: "/manual/use-usb-drive",
            },
            {
              text: "Mount a local disk",
              link: "/manual/mount-local-disk",
            },
            {
              text: "Connect an SMB share",
              link: "/manual/olares/files/mount-SMB",
            },
            {
              text: "Connect an NFS share",
              link: "/manual/olares/files/mount-nfs",
            },
            {
              text: "Connect cloud storage",
              link: "/manual/olares/files/mount-cloud-storage",
            },
          ],
        },
      ],
    },
    {
      text: "System",
      collapsed: false,
      items: [
        {
          text: "Check system and app resource usage",
          link: "/manual/olares/resources-usage",
        },
        {
          text: "Check Olares status and Olares Space usage",
          link: "/manual/space/manage-olares",
        },
        {
          text: "Update Olares",
          link: "/manual/olares/settings/update",
        },
        {
          text: "Reactivate Olares",
          link: "/manual/larepass/activate-olares",
        },
        {
          text: "Manage your Olares device",
          link: "/manual/larepass/manage-olares",
        },
        {
          text: "Back up and restore Olares",
          link: "/manual/olares/settings/backup",
        },
        {
          text: "Cluster",
          collapsed: true,
          items: [
            {
              text: "Install a multi-node Olares cluster",
              link: "/manual/best-practices/install-olares-multi-node",
            },
          ],
        },
        {
          text: "Storage",
          collapsed: true,
          items: [
            {
              text: "Free up disk space",
              link: "/manual/free-up-disk-space",
            },
            {
              text: "Expand Olares system storage",
              link: "/manual/best-practices/expand-storage-in-olares",
            },
          ],
        },
        {
          text: "Advanced administration",
          link: "/manual/olares/controlhub/",
          collapsed: true,
          items: [
            {
              text: "Manage workloads",
              link: "/manual/olares/controlhub/manage-workload",
            },
            {
              text: "Inspect and manage containers",
              link: "/manual/olares/controlhub/manage-container",
            },
            {
              text: "Manage resource configurations",
              link: "/manual/olares/controlhub/manage-resource",
            },
            {
              text: "Manage middleware",
              link: "/manual/olares/controlhub/manage-middleware",
            },
            {
              text: "Run commands in the Control Hub terminal",
              link: "/manual/olares/controlhub/terminal",
            },
            {
              text: "Configure host name resolution in Olares",
              link: "/manual/olares/settings/set-up-hosts",
            },
            {
              text: "Change the reverse proxy",
              link: "/manual/olares/settings/change-frp",
            },
            {
              text: "Configure video playback",
              link: "/manual/olares/settings/video",
            },
            {
              text: "Manage repositories and images",
              link: "/manual/olares/settings/developer",
            },
          ],
        },
      ],
    },
    {
      text: "Advanced setup",
      link: "/manual/best-practices/",
      collapsed: true,
      items: [
        {
          text: "Install and activate Olares from the command line",
          link: "/manual/best-practices/activate-olares-using-cli",
        },
        {
          text: "Install a specific NVIDIA driver version",
          link: "/manual/best-practices/install-specific-cuda-version",
        },
      ],
    },
    {
      text: "Help and troubleshooting",
      link: "/manual/help/",
      collapsed: true,
      items: [
        {
          text: "Collect diagnostic information",
          link: "/manual/collect-diagnostic-information",
        },
        {
          text: "Troubleshooting",
          link: "/manual/help/troubleshooting-guide",
          collapsed: true,
          items: [
            {
              text: "Missing apps in Market",
              link: "/manual/help/ts-missing-apps",
            },
            {
              text: "Cannot remove a stopped app in App exclusive mode",
              link: "/manual/help/ts-cs-app-reappears",
            },
            {
              text: "Application status differs after Control Hub actions",
              link: "/manual/help/ts-inconsistent-app-status",
            },
            {
              text: "App fails during or after installation or update",
              link: "/manual/help/ts-app-fails-after-update",
            },
            {
              text: "Model or engine is not ready",
              link: "/manual/help/ts-model-engine-not-ready",
            },
            {
              text: "Memory is insufficient or not freed",
              link: "/manual/help/ts-free-memory",
            },
            {
              text: "GPU application remains stopped after installation or resume",
              link: "/manual/help/ts-vram-shortage",
            },
            {
              text: "LarePass VPN is not working",
              link: "/manual/help/ts-larepass-vpn-not-working",
            },
            {
              text: "Network not ready or Olares connection error",
              link: "/manual/help/ts-network-not-ready",
            },
            {
              text: "Custom route ID prevents app access",
              link: "/manual/help/ts-custom-route-domain",
            },
            {
              text: "Slow or delayed Steam streaming",
              link: "/manual/help/ts-steam-stream-lag",
            },
            {
              text: "Disk space is full",
              link: "/manual/free-up-disk-space",
            },
            {
              text: "Forgotten desktop login password",
              link: "/manual/help/ts-forget-login-password",
            },
            {
              text: "Login and activation error messages",
              link: "/manual/login-and-activation-errors",
            },
            {
              text: "System error in LarePass",
              link: "/manual/help/ts-system-error",
            },
            {
              text: "Olares One has no power or display",
              link: "/one/ts-no-power-or-display",
            },
          ],
        },
        {
          text: "Known issues",
          link: "/manual/help/known-issues",
        },
        { text: "Glossary", link: "/manual/glossary" },
      ],
    },
  ],
};

export const en = defineConfig({
  lang: "en",
  themeConfig: {
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
