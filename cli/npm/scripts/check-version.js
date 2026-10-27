#!/usr/bin/env node
'use strict';

// prepublishOnly gate. Both checks guard against publishing a package that
// cannot install: a placeholder version resolves to an archive nobody built,
// and a missing checksums.txt makes every postinstall fail closed (see
// scripts/install.js, which treats an unverifiable archive as fatal).

const fs = require('node:fs');
const path = require('node:path');

const v = require('../package.json').version;
let failed = false;

if (v === '0.0.0-placeholder') {
  console.error(`[@olares/cli] refusing to publish placeholder version "${v}".`);
  console.error('[@olares/cli] CI must run "npm version <semver>" before "npm publish".');
  console.error('[@olares/cli] Locally, do NOT publish from this clone.');
  failed = true;
}

const checksums = path.join(__dirname, '..', 'checksums.txt');
if (!fs.existsSync(checksums)) {
  console.error('[@olares/cli] refusing to publish without checksums.txt.');
  console.error('[@olares/cli] postinstall verifies the downloaded archive against it and');
  console.error('[@olares/cli] fails closed when it is absent, so a package published without');
  console.error('[@olares/cli] it can never install. CI must copy cli/output/checksums.txt from');
  console.error('[@olares/cli] the GoReleaser job into cli/npm/ before "npm publish".');
  failed = true;
} else {
  const entries = fs.readFileSync(checksums, 'utf8')
    .split('\n')
    .filter((line) => /^[0-9a-fA-F]{64}\s+\S+\.tar\.gz$/.test(line.trim()));
  if (entries.length === 0) {
    console.error(`[@olares/cli] checksums.txt at ${checksums} names no .tar.gz archive.`);
    failed = true;
  } else if (!entries.some((line) => line.includes(`-v${v}_`))) {
    // The archive names carry the version; a checksums.txt from another build
    // would pass the existence check and then fail every user's install.
    console.error(`[@olares/cli] checksums.txt names no archive for version ${v}.`);
    console.error('[@olares/cli] It is from a different build than the one being published.');
    failed = true;
  }
}

if (failed) {
  process.exit(1);
}
