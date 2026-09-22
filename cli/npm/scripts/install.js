#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const https = require('node:https');
const crypto = require('node:crypto');
const { pipeline } = require('node:stream/promises');

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

// The CDN is the only source, because it is the only place these archives
// are published. A GitHub Releases URL used to be tried first and could never
// have worked on this channel: the release workflow tags locally and never
// pushes (.github/workflows/release-cli.yaml), GoReleaser has `release:
// disable: true`, and nothing uploads a release asset — so every version npm
// can install 404s there. It cost a round trip to github.com on every install
// (~0.7s from a fast network, far worse from one that reaches GitHub slowly)
// and put a red herring at the top of every failure report. Do not add it
// back without first making the release publish those assets under exactly
// these names; `gh api repos/beclab/Olares/releases --jq '.[].tag_name'` shows
// whether an X.Y.Z-cli.N release exists at all.
const DEFAULT_CDN_BASE = 'https://cdn.olares.com';
const MIRROR = process.env.OLARES_CLI_DOWNLOAD_MIRROR;
const CDN_BASE = MIRROR || DEFAULT_CDN_BASE;

const SKIP = process.env.OLARES_CLI_SKIP_DOWNLOAD === '1';
const PLACEHOLDER = PKG.version === '0.0.0-placeholder';

const platformKey = `${process.platform}-${process.arch}`;
const target = PLATFORM_MAP[platformKey];
const isWindows = process.platform === 'win32';
const binName = isWindows ? 'olares-cli.exe' : 'olares-cli';
const vendorDir = path.join(__dirname, '..', 'vendor');
const vendorBin = path.join(vendorDir, binName);

// checksums.txt is produced by GoReleaser over the same archives this script
// fetches, and is packed into the npm tarball by the release workflow. It is
// the root of trust here: npm's own integrity check covers the package, the
// package covers the archive. Nothing about the transport is trusted — not
// the CDN, not a redirect target, and not OLARES_CLI_DOWNLOAD_MIRROR, which
// is a user-set host that can now be pointed anywhere without also being a
// way to substitute the binary.
const checksumsPath = path.join(__dirname, '..', 'checksums.txt');

function archiveName() {
  return `olares-cli-v${VERSION_RAW}_${target}.tar.gz`;
}

// A list rather than a single string: the loop that consumes it reports every
// source it tried, and a second one can be added here the day there is a
// second place to add.
function urls() {
  return [`${CDN_BASE}/${archiveName()}`];
}

// expectedDigest reads the one line of checksums.txt that names this archive.
// Absence is fatal at every step -- a missing file, a missing entry, a digest
// that is not 64 hex characters. A postinstall that falls back to "extract it
// anyway" would make the whole check advisory, and the case where it matters
// is exactly the case where something upstream is not what it should be.
function expectedDigest(name, file) {
  const source = file || checksumsPath;
  if (!fs.existsSync(source)) {
    throw new Error(
      `[SECURITY] checksums.txt not found at ${source}. This package was not ` +
      `assembled by the release workflow; refusing to install an unverifiable binary.`
    );
  }
  const content = fs.readFileSync(source, 'utf8');
  for (const line of content.split('\n')) {
    // GoReleaser writes "<sha256>  <filename>".
    const match = line.trim().match(/^([0-9a-fA-F]{64})\s+\*?(\S+)$/);
    if (match && path.basename(match[2]) === name) {
      return match[1].toLowerCase();
    }
  }
  throw new Error(`[SECURITY] checksums.txt has no entry for ${name}`);
}

function get(url, redirectsLeft = 8) {
  return new Promise((resolve, reject) => {
    const parsed = new URL(url);
    if (parsed.protocol !== 'https:') {
      reject(new Error(`refusing non-https URL ${url}`));
      return;
    }
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

// downloadVerifyExtract lands the archive on disk first and hashes it there,
// so the digest that is checked is the bytes that are extracted. Streaming
// straight into tar.x would decide the answer after the files are already
// unpacked, which is too late to be a check.
async function downloadVerifyExtract(url, expected) {
  // Required here rather than at module load: the skip paths below never
  // extract anything, and the tests exercise the verification helpers in a
  // checkout that has not run `npm install`.
  const tar = require('tar');
  const res = await get(url);
  fs.mkdirSync(vendorDir, { recursive: true });
  const tmp = path.join(
    fs.mkdtempSync(path.join(os.tmpdir(), 'olares-cli-download-')),
    archiveName()
  );
  try {
    const hash = crypto.createHash('sha256');
    res.on('data', (chunk) => hash.update(chunk));
    await pipeline(res, fs.createWriteStream(tmp));

    const actual = hash.digest('hex');
    if (actual !== expected) {
      throw new Error(
        `[SECURITY] checksum mismatch for ${archiveName()}\n` +
        `             expected ${expected}\n` +
        `             actual   ${actual}`
      );
    }
    await tar.x({ file: tmp, cwd: vendorDir, strip: 0 });
  } finally {
    fs.rmSync(path.dirname(tmp), { recursive: true, force: true });
  }
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

  // Resolved before the first request: an archive nobody can verify is not
  // worth downloading, and failing here names the real problem rather than
  // letting every mirror fail the same way further down.
  let expected;
  try {
    expected = expectedDigest(archiveName());
  } catch (err) {
    console.error(`[@olares/cli] ${err.message}`);
    process.exit(1);
  }

  fs.mkdirSync(vendorDir, { recursive: true });

  const tried = [];
  for (const url of urls()) {
    try {
      console.log(`[@olares/cli] downloading ${url}`);
      await downloadVerifyExtract(url, expected);
      if (fs.existsSync(vendorBin)) {
        if (!isWindows) {
          try { fs.chmodSync(vendorBin, 0o755); } catch { /* ignore */ }
        }
        console.log(`[@olares/cli] verified sha256 ${expected}`);
        console.log(`[@olares/cli] installed vendor binary at ${vendorBin}`);
        return;
      }
      tried.push(`${url} (extracted but ${binName} missing)`);
    } catch (err) {
      tried.push(`${url} → ${err.message}`);
    }
  }

  console.error('[@olares/cli] failed to install vendor binary. Tried:');
  for (const t of tried) console.error(`  - ${t}`);
  // A checksum mismatch is not a transport problem, and telling someone to
  // retry against another mirror is the wrong advice for it.
  if (tried.some((t) => t.includes('[SECURITY]'))) {
    console.error('[@olares/cli] At least one source served bytes that do not match the checksum');
    console.error('[@olares/cli] shipped in this package. Do NOT work around this; report it at');
    console.error('[@olares/cli] https://github.com/beclab/Olares/issues');
  } else {
    console.error('[@olares/cli] Set OLARES_CLI_DOWNLOAD_MIRROR to a reachable mirror and retry,');
    console.error('[@olares/cli] or set OLARES_CLI_SKIP_DOWNLOAD=1 to install the JS shim without a binary.');
  }
  process.exit(1);
}

// Exported for scripts/install.test.js. Only the direct invocation installs;
// requiring this file must not download anything.
module.exports = { expectedDigest, archiveName, urls, PLATFORM_MAP };

if (require.main === module) {
  main().catch((err) => {
    console.error('[@olares/cli] postinstall failed:', err);
    process.exit(1);
  });
}
