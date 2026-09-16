'use strict';

// Tests for the postinstall's integrity check. Run with:
//   node --test cli/npm/scripts/
//
// The download path is the one part of @olares/cli that decides which bytes
// become the binary on a user's PATH, and it is reached only by `npm install`
// on a real machine. These cover the decisions it makes before extracting.

const test = require('node:test');
const assert = require('node:assert');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

const { expectedDigest, archiveName, urls, PLATFORM_MAP } = require('./install.js');

const DIGEST_A = 'a'.repeat(64);
const DIGEST_B = 'b'.repeat(64);

function withChecksums(content) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'olares-checksums-'));
  const file = path.join(dir, 'checksums.txt');
  fs.writeFileSync(file, content);
  return file;
}

test('reads the digest for the named archive', () => {
  const file = withChecksums(
    `${DIGEST_A}  olares-cli-v1.2.3_linux_amd64.tar.gz\n` +
    `${DIGEST_B}  olares-cli-v1.2.3_darwin_arm64.tar.gz\n`
  );
  assert.strictEqual(expectedDigest('olares-cli-v1.2.3_darwin_arm64.tar.gz', file), DIGEST_B);
});

test('uppercase digests are normalised', () => {
  const file = withChecksums(`${'A'.repeat(64)}  olares-cli-v1.2.3_linux_amd64.tar.gz\n`);
  assert.strictEqual(expectedDigest('olares-cli-v1.2.3_linux_amd64.tar.gz', file), DIGEST_A);
});

test('a missing checksums.txt is fatal, not a fallback', () => {
  assert.throws(
    () => expectedDigest('olares-cli-v1.2.3_linux_amd64.tar.gz', '/nonexistent/checksums.txt'),
    /\[SECURITY\] checksums.txt not found/
  );
});

test('an archive with no entry is fatal', () => {
  const file = withChecksums(`${DIGEST_A}  olares-cli-v1.2.3_linux_amd64.tar.gz\n`);
  assert.throws(
    () => expectedDigest('olares-cli-v1.2.3_windows_amd64.tar.gz', file),
    /\[SECURITY\] checksums.txt has no entry/
  );
});

// A truncated or otherwise malformed digest must not be accepted as "close
// enough" -- the line is either a 64-hex-character sha256 or it is not a
// checksum, and treating a short one as a match would verify nothing.
test('a malformed digest does not count as an entry', () => {
  const file = withChecksums(`deadbeef  olares-cli-v1.2.3_linux_amd64.tar.gz\n`);
  assert.throws(
    () => expectedDigest('olares-cli-v1.2.3_linux_amd64.tar.gz', file),
    /has no entry/
  );
});

// GoReleaser writes plain names, but the "*" binary marker is common in
// sha256sum output and a checksums.txt produced by hand may carry it.
test('a binary-mode marker is tolerated', () => {
  const file = withChecksums(`${DIGEST_A} *olares-cli-v1.2.3_linux_amd64.tar.gz\n`);
  assert.strictEqual(expectedDigest('olares-cli-v1.2.3_linux_amd64.tar.gz', file), DIGEST_A);
});

test('every supported platform maps to a GoReleaser target', () => {
  for (const [key, target] of Object.entries(PLATFORM_MAP)) {
    assert.match(target, /^(linux|darwin|windows)_(amd64|arm64|arm)$/, `${key} → ${target}`);
  }
});

test('both sources are https and name the same archive', () => {
  const name = archiveName();
  for (const url of urls()) {
    assert.match(url, /^https:\/\//, url);
    assert.ok(url.endsWith(name), `${url} should end with ${name}`);
  }
});
