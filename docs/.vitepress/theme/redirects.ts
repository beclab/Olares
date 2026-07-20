export const redirects = {
    // Root → default docs landing
    '/': '/manual/overview',

    // Studio docs are hidden; redirect legacy Studio URLs to the docs overview.
    '/manual/olares/studio/': '/manual/overview',
    '/manual/olares/studio/deploy': '/manual/overview',
    '/manual/olares/studio/develop': '/manual/overview',
    '/manual/olares/studio/package-upload': '/manual/overview',
    '/manual/olares/studio/assets': '/manual/overview',
    '/developer/develop/tutorial/studio': '/manual/overview',
    '/zh/developer/develop/tutorial/studio': '/zh/manual/overview',

    // Hidden Studio docs → docs overview
    '/developer/develop/tutorial/': '/manual/overview',
    '/developer/develop/tutorial/deploy': '/manual/overview',
    '/developer/develop/tutorial/develop': '/manual/overview',
    '/developer/develop/tutorial/package-upload': '/manual/overview',
    '/developer/develop/tutorial/assets': '/manual/overview',
    '/developer/develop/tutorial/note/': '/manual/overview',
    '/developer/develop/tutorial/note/create': '/manual/overview',
    '/developer/develop/tutorial/note/backend': '/manual/overview',
    '/developer/develop/tutorial/note/frontend': '/manual/overview',
    '/developer/contribute/system-app/overview': '/manual/overview',
    '/developer/contribute/system-app/deployment': '/manual/overview',
    '/developer/contribute/system-app/olares-manifest': '/manual/overview',
    '/developer/contribute/system-app/install': '/manual/overview',
    '/developer/contribute/system-app/other': '/manual/overview',
    '/one/deploy': '/manual/overview',
    '/zh/developer/develop/tutorial/': '/zh/manual/overview',
    '/zh/developer/develop/tutorial/deploy': '/zh/manual/overview',
    '/zh/developer/develop/tutorial/develop': '/zh/manual/overview',
    '/zh/developer/develop/tutorial/package-upload': '/zh/manual/overview',
    '/zh/developer/develop/tutorial/assets': '/zh/manual/overview',
    '/zh/developer/develop/tutorial/note/': '/zh/manual/overview',
    '/zh/developer/develop/tutorial/note/create': '/zh/manual/overview',
    '/zh/developer/develop/tutorial/note/backend': '/zh/manual/overview',
    '/zh/developer/develop/tutorial/note/frontend': '/zh/manual/overview',
    '/zh/developer/contribute/system-app/overview': '/zh/manual/overview',
    '/zh/developer/contribute/system-app/deployment': '/zh/manual/overview',
    '/zh/developer/contribute/system-app/olares-manifest': '/zh/manual/overview',
    '/zh/developer/contribute/system-app/install': '/zh/manual/overview',
    '/zh/developer/contribute/system-app/other': '/zh/manual/overview',
    '/zh/one/deploy': '/zh/manual/overview',

    // Refactor: /space/** → /manual/space/**
    '/space/': '/manual/space/',
    '/space/billing': '/manual/space/billing',
    '/space/manage-domain': '/manual/space/manage-domain',
    '/space/manage-accounts': '/manual/space/manage-accounts',
    '/space/backup-restore': '/manual/space/backup-restore',
    '/space/create-olares': '/manual/space/create-olares',
    '/space/manage-olares': '/manual/space/manage-olares',
    '/space/host-domain': '/manual/space/host-domain',

    // Refactor: /zh/space/** → /zh/manual/space/**
    '/zh/space/': '/zh/manual/space/',
    '/zh/space/billing': '/zh/manual/space/billing',
    '/zh/space/manage-domain': '/zh/manual/space/manage-domain',
    '/zh/space/manage-accounts': '/zh/manual/space/manage-accounts',
    '/zh/space/backup-restore': '/zh/manual/space/backup-restore',
    '/zh/space/create-olares': '/zh/manual/space/create-olares',
    '/zh/space/manage-olares': '/zh/manual/space/manage-olares',
    '/zh/space/host-domain': '/zh/manual/space/host-domain',

    // Rename: deerflow → deerflow2 (permanent: old name is retired)
    '/use-cases/deerflow': '/use-cases/deerflow2',
    '/zh/use-cases/deerflow': '/zh/use-cases/deerflow2',

    // Rename: ace-step → ace-step-1.5 (permanent: old name is retired)
    '/use-cases/ace-step': '/use-cases/ace-step-1.5',
    '/zh/use-cases/ace-step': '/zh/use-cases/ace-step-1.5',

    // Refactor: openwebui-ollama → openwebui (merged: Ollama is now Option A in the unified quick start)
    '/use-cases/openwebui-ollama': '/use-cases/openwebui',
    '/zh/use-cases/openwebui-ollama': '/zh/use-cases/openwebui',

    // Rename: descriptive slugs → brand slugs (permanent: old names retired)
    '/use-cases/stream-media': '/use-cases/jellyfin',
    '/zh/use-cases/stream-media': '/zh/use-cases/jellyfin',
    '/use-cases/stream-game': '/use-cases/steam-stream',
    '/zh/use-cases/stream-game': '/zh/use-cases/steam-stream',
    '/use-cases/play-games-directly': '/use-cases/steam-direct-play',
    '/zh/use-cases/play-games-directly': '/zh/use-cases/steam-direct-play',

    // Refactor: /manual/concepts/** → /developer/concepts/**
    '/manual/system-architecture': '/developer/concepts/system-architecture',
    '/manual/concepts/': '/developer/concepts/',
    '/manual/concepts/account': '/developer/concepts/account',
    '/manual/concepts/application': '/developer/concepts/application',
    '/manual/concepts/architecture': '/developer/concepts/architecture',
    '/manual/concepts/data': '/developer/concepts/data',
    '/manual/concepts/did': '/developer/concepts/did',
    '/manual/concepts/faq': '/developer/concepts/faq',
    '/manual/concepts/network': '/developer/concepts/network',
    '/manual/concepts/olares-id': '/developer/concepts/olares-id',
    '/manual/concepts/registry': '/developer/concepts/registry',
    '/manual/concepts/reputation': '/developer/concepts/reputation',
    '/manual/concepts/secrets': '/developer/concepts/secrets',
    '/manual/concepts/self-sovereign-network': '/developer/concepts/self-sovereign-network',
    '/manual/concepts/system-architecture': '/developer/concepts/system-architecture',
    '/manual/concepts/vc': '/developer/concepts/vc',
    '/manual/concepts/wallet': '/developer/concepts/wallet',

    // Rename: /developer/install/cli/olares-* → /developer/install/cli/*
    '/developer/install/cli/olares-info': '/developer/install/cli/info',
    '/developer/install/cli/olares-start': '/developer/install/cli/start',
    '/developer/install/cli/olares-stop': '/developer/install/cli/stop',
    '/developer/install/cli/olares-uninstall': '/developer/install/cli/uninstall',
    '/developer/install/cli/olares-release': '/developer/install/cli/release',
    '/developer/install/cli/olares-change-ip': '/developer/install/cli/change-ip',
    '/developer/install/cli/olares-download': '/developer/install/cli/download',
    '/developer/install/cli/olares-logs': '/developer/install/cli/logs',
    '/developer/install/cli/olares-backups': '/developer/install/cli/backups',

    // Rename: /zh/developer/install/cli/olares-* → /zh/developer/install/cli/*
    '/zh/developer/install/cli/olares-info': '/zh/developer/install/cli/info',
    '/zh/developer/install/cli/olares-start': '/zh/developer/install/cli/start',
    '/zh/developer/install/cli/olares-stop': '/zh/developer/install/cli/stop',
    '/zh/developer/install/cli/olares-uninstall': '/zh/developer/install/cli/uninstall',
    '/zh/developer/install/cli/olares-release': '/zh/developer/install/cli/release',
    '/zh/developer/install/cli/olares-change-ip': '/zh/developer/install/cli/change-ip',
    '/zh/developer/install/cli/olares-download': '/zh/developer/install/cli/download',
    '/zh/developer/install/cli/olares-logs': '/zh/developer/install/cli/logs',
    '/zh/developer/install/cli/olares-backups': '/zh/developer/install/cli/backups',

    // Refactor: /zh/manual/concepts/** → /zh/developer/concepts/**
    '/zh/manual/system-architecture': '/zh/developer/concepts/system-architecture',
    '/zh/manual/concepts/': '/zh/developer/concepts/',
    '/zh/manual/concepts/account': '/zh/developer/concepts/account',
    '/zh/manual/concepts/application': '/zh/developer/concepts/application',
    '/zh/manual/concepts/architecture': '/zh/developer/concepts/architecture',
    '/zh/manual/concepts/data': '/zh/developer/concepts/data',
    '/zh/manual/concepts/did': '/zh/developer/concepts/did',
    '/zh/manual/concepts/network': '/zh/developer/concepts/network',
    '/zh/manual/concepts/olares-id': '/zh/developer/concepts/olares-id',
    '/zh/manual/concepts/registry': '/zh/developer/concepts/registry',
    '/zh/manual/concepts/reputation': '/zh/developer/concepts/reputation',
    '/zh/manual/concepts/secrets': '/zh/developer/concepts/secrets',
    '/zh/manual/concepts/self-sovereign-network': '/zh/developer/concepts/self-sovereign-network',
    '/zh/manual/concepts/system-architecture': '/zh/developer/concepts/system-architecture',
    '/zh/manual/concepts/vc': '/zh/developer/concepts/vc',
    '/zh/manual/concepts/wallet': '/zh/developer/concepts/wallet',

    // Refactor: /manual/docs-home → /manual/overview
    '/manual/docs-home': '/manual/overview',
    
    // Refactor: /zh/manual/docs-home → /zh/manual/overview
    '/zh/manual/docs-home': '/zh/manual/overview',

    // Removed: legacy developer install-step pages (onboarding docs that once lived
    // under /developer/install/) → the real Get started install overview. Note:
    // /developer/install/ is now the "Cluster management" (olares-cli) page, so these
    // must NOT redirect there. No zh installation-troubleshooting: it never existed in zh.
    '/developer/install/activate-olares': '/manual/get-started/install-olares',
    '/developer/install/install-and-activate-olares': '/manual/get-started/install-olares',
    '/developer/install/log-in-to-olares': '/manual/get-started/install-olares',
    '/developer/install/installation-troubleshooting': '/manual/get-started/install-olares',
    '/developer/install/reusables': '/manual/get-started/install-olares',
    '/zh/developer/install/activate-olares': '/zh/manual/get-started/install-olares',
    '/zh/developer/install/install-and-activate-olares': '/zh/manual/get-started/install-olares',
    '/zh/developer/install/log-in-to-olares': '/zh/manual/get-started/install-olares',
    '/zh/developer/install/reusables': '/zh/manual/get-started/install-olares',

    // Removed: empty advanced-dev stub pages → advanced overview
    '/developer/develop/advanced/rss': '/developer/develop/advanced/',
    '/developer/develop/advanced/frontend': '/developer/develop/advanced/',
    '/developer/develop/advanced/notification': '/developer/develop/advanced/',
    '/zh/developer/develop/advanced/rss': '/zh/developer/develop/advanced/',
    '/zh/developer/develop/advanced/frontend': '/zh/developer/develop/advanced/',
    '/zh/developer/develop/advanced/notification': '/zh/developer/develop/advanced/',

    // Removed: empty contribute overview stub → contribute landing
    '/developer/contribute/overview': '/developer/contribute/olares',
    '/zh/developer/contribute/overview': '/zh/developer/contribute/olares',

    // Removed: stale olares-id section hero page → first olares-id doc
    '/developer/contribute/olares-id/': '/developer/contribute/olares-id/contract/contract',

    // Removed: single-gpu/multi-gpu were noindex include-only fragments of gpu-resource → consolidated page
    '/manual/olares/settings/single-gpu': '/manual/olares/settings/gpu-resource',
    '/manual/olares/settings/multi-gpu': '/manual/olares/settings/gpu-resource',
    '/zh/manual/olares/settings/single-gpu': '/zh/manual/olares/settings/gpu-resource',
    '/zh/manual/olares/settings/multi-gpu': '/zh/manual/olares/settings/gpu-resource',

    // Removed: /one/ software-feature duplicates consolidated into /manual/ and /use-cases/
    // (hardware/device docs under /one/ are kept). See docs cleanup for SEO consolidation.
    '/one/files': '/manual/olares/files/',
    '/one/vault': '/manual/olares/vault/',
    '/one/market': '/manual/olares/market/market',
    '/one/dashboard': '/manual/olares/resources-usage',
    '/one/gpu': '/manual/olares/settings/gpu-resource',
    '/one/backup-restore': '/manual/olares/settings/backup',
    '/one/customize': '/manual/olares/settings/language-appearance',
    '/one/space': '/manual/space/manage-olares',
    '/one/windows': '/use-cases/windows',
    '/one/steam-stream': '/use-cases/steam-stream',
    '/one/steam-direct-play': '/use-cases/steam-direct-play',
    '/one/wise-download': '/manual/olares/wise/',
    '/one/open-webui': '/use-cases/openwebui',
    '/one/comfyui': '/use-cases/comfyui',
    '/one/deerflow': '/use-cases/deerflow2',
    '/one/ace-step': '/use-cases/ace-step-1.5',
    '/one/create-users': '/manual/olares/settings/manage-team',
    '/one/config-app-access': '/manual/olares/settings/manage-entrance',
    '/zh/one/files': '/zh/manual/olares/files/',
    '/zh/one/vault': '/zh/manual/olares/vault/',
    '/zh/one/market': '/zh/manual/olares/market/market',
    '/zh/one/dashboard': '/zh/manual/olares/resources-usage',
    '/zh/one/gpu': '/zh/manual/olares/settings/gpu-resource',
    '/zh/one/backup-restore': '/zh/manual/olares/settings/backup',
    '/zh/one/customize': '/zh/manual/olares/settings/language-appearance',
    '/zh/one/space': '/zh/manual/space/manage-olares',
    '/zh/one/windows': '/zh/use-cases/windows',
    '/zh/one/steam-stream': '/zh/use-cases/steam-stream',
    '/zh/one/steam-direct-play': '/zh/use-cases/steam-direct-play',
    '/zh/one/wise-download': '/zh/manual/olares/wise/',
    '/zh/one/open-webui': '/zh/use-cases/openwebui',
    '/zh/one/comfyui': '/zh/use-cases/comfyui',
    '/zh/one/deerflow': '/zh/use-cases/deerflow2',
    '/zh/one/ace-step': '/zh/use-cases/ace-step-1.5',
    '/zh/one/create-users': '/zh/manual/olares/settings/manage-team',
    '/zh/one/config-app-access': '/zh/manual/olares/settings/manage-entrance',
}

// Temporary redirects (302): content is offline but the URL may be reused later.
// Once the page is restored at its original URL, remove the entry here.
// Once the move is confirmed permanent, promote the entry to `redirects` above.
export const temporaryRedirects = {
}
