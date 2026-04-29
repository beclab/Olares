export const redirects = {
    // Root → default docs landing
    '/': '/manual/overview',

    // Refactor: /manual/olares/studio/** → /developer/develop/tutorial/**
    '/manual/olares/studio/': '/developer/develop/tutorial/',
    '/manual/olares/studio/deploy': '/developer/develop/tutorial/deploy',
    '/manual/olares/studio/develop': '/developer/develop/tutorial/develop',
    '/manual/olares/studio/package-upload': '/developer/develop/tutorial/package-upload',
    '/manual/olares/studio/assets': '/developer/develop/tutorial/assets',
    '/developer/develop/tutorial/studio': '/developer/develop/tutorial',
    '/zh/developer/develop/tutorial/studio': '/zh/developer/develop/tutorial',

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
}

// Temporary redirects (302): content is offline but the URL may be reused later.
// Once the page is restored at its original URL, remove the entry here.
// Once the move is confirmed permanent, promote the entry to `redirects` above.
export const temporaryRedirects = {
    // /one/deerflow content temporarily offline; redirecting to deerflow2 in the meantime
    '/one/deerflow': '/use-cases/deerflow2',
    '/zh/one/deerflow': '/zh/use-cases/deerflow2',

    // /one/ace-step content temporarily offline; redirecting to ace-step-1.5 in the meantime
    '/one/ace-step': '/use-cases/ace-step-1.5',
    '/zh/one/ace-step': '/zh/use-cases/ace-step-1.5',
}