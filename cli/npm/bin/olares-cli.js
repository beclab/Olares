#!/usr/bin/env node
'use strict';

const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');

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

const res = spawnSync(bin, process.argv.slice(2), {
  stdio: 'inherit',
  windowsHide: true,
});

if (res.error) {
  console.error('[@olares/cli] failed to spawn vendor binary:', res.error.message);
  process.exit(1);
}
process.exit(typeof res.status === 'number' ? res.status : 1);
