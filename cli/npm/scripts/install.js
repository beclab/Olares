#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const https = require('node:https');
const { pipeline } = require('node:stream/promises');
const { Readable } = require('node:stream');
const tar = require('tar');

const PKG = require('../package.json');
const VERSION_RAW = String(PKG.version).replace(/^v/, '');

const PLATFORM_MAP = {
  'linux-x64': 'linux_amd64',
  'linux-arm64': 'linux_arm64',
  'linux-arm': 'linux_arm',
  'darwin-x64': 'darwin_amd64',
  'darwin-arm64': 'darwin_arm64',
  'win32-x64': 'windows_amd64',
  'win32-arm64': 'windows_arm64',
};

const GH_BASE = 'https://github.com/beclab/Olares/releases/download';
const CDN_BASE = process.env.OLARES_CLI_DOWNLOAD_MIRROR || 'https://cdn.olares.com';

const SKIP = process.env.OLARES_CLI_SKIP_DOWNLOAD === '1';
const PLACEHOLDER = PKG.version === '0.0.0-placeholder';

const platformKey = `${process.platform}-${process.arch}`;
const target = PLATFORM_MAP[platformKey];
const isWindows = process.platform === 'win32';
const binName = isWindows ? 'olares-cli.exe' : 'olares-cli';
const vendorDir = path.join(__dirname, '..', 'vendor');
const vendorBin = path.join(vendorDir, binName);

function archiveName() {
  return `olares-cli-v${VERSION_RAW}_${target}.tar.gz`;
}

function urls() {
  const name = archiveName();
  return [
    `${GH_BASE}/${VERSION_RAW}/${name}`,
    `${CDN_BASE}/${name}`,
  ];
}

function get(url, redirectsLeft = 8) {
  return new Promise((resolve, reject) => {
    const req = https.get(url, { headers: { 'User-Agent': '@olares/cli postinstall' } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        if (redirectsLeft <= 0) {
          reject(new Error(`too many redirects fetching ${url}`));
          return;
        }
        res.resume();
        const next = new URL(res.headers.location, url).toString();
        resolve(get(next, redirectsLeft - 1));
        return;
      }
      if (res.statusCode !== 200) {
        reject(new Error(`${url} → HTTP ${res.statusCode}`));
        res.resume();
        return;
      }
      resolve(res);
    });
    req.on('error', reject);
    req.setTimeout(60_000, () => req.destroy(new Error(`timeout fetching ${url}`)));
  });
}

async function downloadAndExtract(url) {
  const res = await get(url);
  fs.mkdirSync(vendorDir, { recursive: true });
  await pipeline(
    Readable.from(res),
    tar.x({ cwd: vendorDir, strip: 0 }),
  );
}

async function main() {
  if (SKIP) {
    console.log('[@olares/cli] OLARES_CLI_SKIP_DOWNLOAD=1 set, skipping vendor download.');
    return;
  }
  if (PLACEHOLDER) {
    console.log('[@olares/cli] package version is the placeholder; skipping vendor download.');
    console.log('[@olares/cli] (CI sets the real version before publish.)');
    return;
  }
  if (!target) {
    console.error(`[@olares/cli] unsupported platform: ${platformKey}`);
    console.error('[@olares/cli] Supported:', Object.keys(PLATFORM_MAP).join(', '));
    process.exit(1);
  }

  fs.mkdirSync(vendorDir, { recursive: true });

  const tried = [];
  for (const url of urls()) {
    try {
      console.log(`[@olares/cli] downloading ${url}`);
      await downloadAndExtract(url);
      if (fs.existsSync(vendorBin)) {
        if (!isWindows) {
          try { fs.chmodSync(vendorBin, 0o755); } catch { /* ignore */ }
        }
        console.log(`[@olares/cli] installed vendor binary at ${vendorBin}`);
        return;
      }
      tried.push(`${url} (extracted but ${binName} missing)`);
    } catch (err) {
      tried.push(`${url} → ${err.message}`);
    }
  }

  console.error('[@olares/cli] failed to download vendor binary. Tried:');
  for (const t of tried) console.error(`  - ${t}`);
  console.error('[@olares/cli] Set OLARES_CLI_DOWNLOAD_MIRROR to a reachable mirror and retry,');
  console.error('[@olares/cli] or set OLARES_CLI_SKIP_DOWNLOAD=1 to install the JS shim without a binary.');
  process.exit(1);
}

main().catch((err) => {
  console.error('[@olares/cli] postinstall failed:', err);
  process.exit(1);
});
