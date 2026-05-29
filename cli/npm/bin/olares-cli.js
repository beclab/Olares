#!/usr/bin/env node
'use strict';

const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');

// `npx @olares/cli install` is a first-run wizard implemented purely in JS:
// it promotes the npx invocation to a global `npm install -g @olares/cli`
// and then installs the agent skills via `npx skills add beclab/Olares`. It
// does NOT install Olares OS (use `curl -fsSL https://olares.sh | bash` for
// that on a Linux host) and does NOT touch the Go binary. Intercept the verb
// here so it works even when the vendor binary failed to download.
const args = process.argv.slice(2);
if (args[0] === 'install') {
  require('../scripts/install-wizard.js');
  return;
}

const isWindows = process.platform === 'win32';
const binName = isWindows ? 'olares-cli.exe' : 'olares-cli';
const bin = path.join(__dirname, '..', 'vendor', binName);

if (!fs.existsSync(bin)) {
  console.error(`[@olares/cli] vendor binary not found at ${bin}`);
  console.error('[@olares/cli] Re-run `npm install -g @olares/cli` to repopulate it, or set');
  console.error('[@olares/cli] OLARES_CLI_DOWNLOAD_MIRROR / OLARES_CLI_SKIP_DOWNLOAD if the');
  console.error('[@olares/cli] postinstall step was skipped on purpose.');
  process.exit(1);
}

// OLARES_CLI_REMOTE_ONLY=1 tells the Go binary's root command tree to skip
// registering host-side verbs (uninstall, upgrade, node, os, gpu, disk,
// wizard, user, osinfo, amdgpu) that require an Olares host filesystem laid
// down by the install wizard. See cli/cmd/ctl/root.go. npx users never run
// that wizard, so exposing those verbs would just produce confusing
// manifest-not-found errors. The host-bundled binary at
// /usr/local/bin/olares-cli is invoked by install.sh without this env var
// and keeps the full verb set.
const res = spawnSync(bin, args, {
  stdio: 'inherit',
  windowsHide: true,
  env: { ...process.env, OLARES_CLI_REMOTE_ONLY: '1' },
});

if (res.error) {
  console.error('[@olares/cli] failed to spawn vendor binary:', res.error.message);
  process.exit(1);
}
process.exit(typeof res.status === 'number' ? res.status : 1);
