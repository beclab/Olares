#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const { execFileSync, execFile } = require('node:child_process');
const p = require('@clack/prompts');

const PKG = '@olares/cli';
const SKILLS_REPO = 'beclab/Olares';
const isWindows = process.platform === 'win32';
const isLinux = process.platform === 'linux';

// Canonical system paths an Olares OS bundle or `make install` build ends up
// at -- outside any npm prefix, so npm can't see/manage them.
const SYSTEM_OLARES_CLI_PATHS = [
  '/usr/local/bin/olares-cli',
  '/usr/bin/olares-cli',
];

// ---------------------------------------------------------------------------
// Messages (English only for now; --lang/zh planned as a follow-up)
// ---------------------------------------------------------------------------

const msg = {
  setup:           'Setting up Olares CLI...',
  step1:           'Installing %s globally...',
  step1Upgrade:    'Upgrading %s (v%s -> v%s)...',
  step1Skip:       'Already installed (v%s). Skipped',
  step1Done:       'Installed globally',
  step1Upgraded:   'Upgraded to v%s',
  step1Fail:       'Failed to install globally. Run manually: npm install -g %s',
  step1EexistStop: 'Skipped global install: existing olares-cli left in place at /usr/local/bin/olares-cli',
  step1Eexist:
    'On a Linux Olares host the OS bundle owns /usr/local/bin/olares-cli; npm will not overwrite it.\n' +
    'The wizard exited before installing skills -- finish the side-by-side install yourself (keeps the OS bundle for system-layer verbs):\n' +
    '  npm install -g %s --prefix=$HOME/.olares-cli-npm\n' +
    '  export PATH="$HOME/.olares-cli-npm/bin:$PATH"   # before /usr/local/bin\n' +
    '  npx skills add %s -y -g\n' +
    'See cli/README.md "On a Linux Olares host" for details.',
  preflightKeepRelease:
    'Detected release olares-cli at %s (%s); keeping it.',
  preflightReplaceDev:
    'Detected non-release olares-cli at %s (%s); replacing.',
  preflightNoPermission:
    'Cannot remove %s (%s).\n' +
    'Re-run the wizard with sudo so it can replace the dev build:\n' +
    '  sudo $(command -v npx) -y %s@latest install\n' +
    'Or remove it yourself:\n' +
    '  sudo rm %s\n' +
    'then re-run `npx %s@latest install`.',
  step2Spinner:    'Installing AI skills...',
  step2Skip:       'Skills already installed. Skipped',
  step2Done:       'Skills installed',
  step2Fail:       'Failed to install skills. Run manually: npx skills add %s -y -g',
  done:
    'You are all set!\n\n' +
    'Next:\n' +
    '  olares-cli profile login --olares-id <your-olares-id>   # authenticate (browser/password + optional TOTP)\n' +
    '  olares-cli profile current                              # verify\n\n' +
    'Then tell your AI agent: "Load the olares-shared skill, then use olares-cli to ..."',
  nonTtyHint:
    'To finish setup, run:\n' +
    '  olares-cli profile login --olares-id <your-olares-id>\n' +
    '  olares-cli profile current',
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function fmt(template, ...values) {
  let i = 0;
  return template.replace(/%s/g, () => values[i++] ?? '');
}

function execCmd(cmd, args, opts) {
  if (isWindows) {
    return execFileSync('cmd.exe', ['/c', cmd, ...args], opts);
  }
  return execFileSync(cmd, args, opts);
}

function runSilent(cmd, args, opts = {}) {
  return execCmd(cmd, args, {
    stdio: ['ignore', 'pipe', 'pipe'],
    ...opts,
  });
}

function runSilentAsync(cmd, args, opts = {}) {
  const actualCmd = isWindows ? 'cmd.exe' : cmd;
  const actualArgs = isWindows ? ['/c', cmd, ...args] : args;
  return new Promise((resolve, reject) => {
    execFile(actualCmd, actualArgs, {
      stdio: ['ignore', 'pipe', 'pipe'],
      ...opts,
    }, (err, stdout, stderr) => {
      if (err) {
        err.stderr = stderr;
        err.stdout = stdout;
        reject(err);
      } else {
        resolve(stdout);
      }
    });
  });
}

function getLatestVersion() {
  try {
    const out = runSilent('npm', ['view', PKG, 'version'], { timeout: 15000 });
    const ver = out.toString().trim();
    return /^\d+\.\d+\.\d+/.test(ver) ? ver : null;
  } catch (_) {
    return null;
  }
}

function getGloballyInstalledVersion() {
  try {
    const out = runSilent('npm', ['list', '-g', PKG], { timeout: 15000 });
    const match = out.toString().match(/@olares\/cli@(\d+\.\d+\.\d+[^\s]*)/);
    return match ? match[1] : null;
  } catch (_) {
    return null;
  }
}

// Heuristic: did `npm install -g` fail because /usr/local/bin/olares-cli already exists?
// On Linux Olares hosts the OS bundle owns that path and npm refuses to clobber it.
function looksLikeEexistConflict(err) {
  if (!isLinux) return false;
  const blob = `${err.message || ''}\n${err.stderr || ''}`;
  if (/EEXIST/i.test(blob) && /olares-cli/.test(blob)) return true;
  try {
    return fs.existsSync('/usr/local/bin/olares-cli');
  } catch (_) {
    return false;
  }
}

// Locate an OS-bundle or make-install copy at the canonical system paths.
// These live outside any npm prefix; npm can neither see nor manage them.
function detectSystemOlaresCli() {
  for (const candidate of SYSTEM_OLARES_CLI_PATHS) {
    try {
      if (fs.existsSync(candidate)) return candidate;
    } catch (_) { /* fall through */ }
  }
  return null;
}

function readOlaresCliVersion(binPath) {
  try {
    const out = execFileSync(binPath, ['--version'], {
      encoding: 'utf8',
      timeout: 5000,
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    return out.trim();
  } catch (_) {
    return null;
  }
}

function isAllDigits(s) {
  if (!s) return false;
  for (let i = 0; i < s.length; i++) {
    const c = s.charCodeAt(i);
    if (c < 48 || c > 57) return false;
  }
  return true;
}

// Cobra emits "olares-cli version X.Y.Z[-PRE]". Pull the token after
// "version ", strip an optional leading "v", and stop at the first
// whitespace. Returns null if no version token can be found.
function extractVersionToken(verStr) {
  const marker = 'version ';
  const idx = verStr.indexOf(marker);
  if (idx < 0) return null;
  const tail = verStr.slice(idx + marker.length).trim();
  if (!tail) return null;
  let end = tail.length;
  for (let i = 0; i < tail.length; i++) {
    const c = tail.charCodeAt(i);
    if (c === 0x20 || c === 0x09 || c === 0x0a || c === 0x0d) { end = i; break; }
  }
  let token = tail.slice(0, end);
  if (token.startsWith('v')) token = token.slice(1);
  return token || null;
}

// Accept only the shapes the release pipeline can produce:
//   rc, rc1, rc.1, beta, beta2, beta.2, alpha, alpha3, alpha.3 (case-insensitive)
// Anything trailing -- a `-3-gHASH` from git describe, a `-dirty` marker, a
// numeric-only suffix, or anything else -- falls through to false.
function isReleasePre(pre) {
  const lower = pre.toLowerCase();
  for (const prefix of ['rc', 'beta', 'alpha']) {
    if (!lower.startsWith(prefix)) continue;
    const rest = pre.slice(prefix.length);
    if (rest === '') return true;                                     // rc, beta, alpha
    if (isAllDigits(rest)) return true;                               // rc1, beta2, alpha3
    if (rest[0] === '.' && isAllDigits(rest.slice(1))) return true;   // rc.1, beta.2, alpha.3
    return false;
  }
  return false;
}

// Classify the version string from `olares-cli --version`. Returns true ONLY
// for what the release pipeline can produce:
//   - stable:     1.12.7              <-- this is the "正式版本", never replaced
//   - prerelease: 1.12.8-rc1, 1.13.0-beta.1, 1.14.0-alpha2
// Returns false for everything else (treat as dev/test, safe to replace):
//   - 0.0.0-development (placeholder default in cli/version/version.go)
//   - 1.12.5-cli.2-3-gddae4ca9c, 1.12.7-rc1-3-gabc-dirty (git describe output)
//   - 1.12.7-12345678 (check.yaml PR-test build; numeric suffix isn't a tag)
//   - `dev` (Makefile no-git fallback) or any unparseable output
function isReleaseGradeVersion(verStr) {
  if (!verStr) return false;
  const token = extractVersionToken(verStr);
  if (!token) return false;

  const dash = token.indexOf('-');
  const core = dash === -1 ? token : token.slice(0, dash);
  const pre  = dash === -1 ? ''    : token.slice(dash + 1);

  const parts = core.split('.');
  if (parts.length !== 3) return false;
  for (const p of parts) {
    if (!isAllDigits(p)) return false;
  }
  if (core === '0.0.0') return false;

  if (pre === '') return true;            // 正式版本 (stable, no -PRE suffix)
  return isReleasePre(pre);
}

function tryUnlink(filePath) {
  try {
    fs.unlinkSync(filePath);
    return { ok: true };
  } catch (err) {
    return { ok: false, err };
  }
}

// ---------------------------------------------------------------------------
// Steps
// ---------------------------------------------------------------------------

async function stepInstallGlobally(interactive) {
  // Preflight: if a system-path olares-cli is present (OS bundle or
  // `make install` artifact), decide keep-vs-replace based on its version.
  // Release-grade versions (stable/rc/beta/alpha) are left alone -- the
  // npm copy will install side-by-side and `looksLikeEexistConflict` below
  // gives the user the workaround if npm would otherwise clobber it.
  // Dev / test / unparseable versions are removed so npm can install over
  // them. If we can't remove (typical: not running as root), bail with a
  // sudo hint so the user knows what to do next.
  const sysCli = detectSystemOlaresCli();
  if (sysCli) {
    const verStr = readOlaresCliVersion(sysCli);
    const verDisplay = verStr || 'unknown';
    if (isReleaseGradeVersion(verStr)) {
      const line = fmt(msg.preflightKeepRelease, sysCli, verDisplay);
      if (interactive) p.log.info(line); else console.log(line);
    } else {
      const line = fmt(msg.preflightReplaceDev, sysCli, verDisplay);
      if (interactive) p.log.warn(line); else console.warn(line);
      const rm = tryUnlink(sysCli);
      if (!rm.ok) {
        const reason = rm.err.code || rm.err.message || 'unknown error';
        const hint = fmt(msg.preflightNoPermission, sysCli, reason, PKG, sysCli, PKG);
        if (interactive) p.log.error(hint); else console.error(hint);
        process.exit(1);
      }
    }
  }

  const installedVer = getGloballyInstalledVersion();
  const latestVer = getLatestVersion();
  // Exact-inequality only -- we never had to *compare* versions, just decide
  // whether to call `npm install -g` again. This matters for npm's `latest`
  // dist-tag: `1.12.5-cli.2` and `1.12.5-cli.4` share the same MAJOR.MINOR.PATCH
  // so a semver core compare wrongly reported "already at latest".
  const needsUpgrade = installedVer && latestVer && installedVer !== latestVer;

  if (installedVer && !needsUpgrade) {
    const line = fmt(msg.step1Skip, installedVer);
    if (interactive) p.log.info(line); else console.log(line);
    return;
  }

  const startLine = needsUpgrade
    ? fmt(msg.step1Upgrade, PKG, installedVer, latestVer)
    : fmt(msg.step1, PKG);
  const doneLine = needsUpgrade
    ? fmt(msg.step1Upgraded, latestVer)
    : msg.step1Done;

  const s = interactive ? p.spinner() : null;
  if (s) s.start(startLine); else console.log(startLine);

  try {
    await runSilentAsync('npm', ['install', '-g', PKG], { timeout: 600000 });
    if (s) s.stop(doneLine); else console.log(doneLine);
  } catch (err) {
    if (looksLikeEexistConflict(err)) {
      if (s) s.stop(msg.step1EexistStop); else console.error(msg.step1EexistStop);
      const hint = fmt(msg.step1Eexist, PKG, SKILLS_REPO);
      if (interactive) p.log.warn(hint); else console.error(hint);
    } else {
      if (s) s.stop(fmt(msg.step1Fail, PKG)); else console.error(fmt(msg.step1Fail, PKG));
    }
    process.exit(1);
  }
}

async function skillsAlreadyInstalled() {
  try {
    const out = await runSilentAsync('npx', ['-y', 'skills', 'ls', '-g'], { timeout: 600000 });
    return /^olares-/m.test(out.toString());
  } catch (_) {
    return false;
  }
}

async function stepInstallSkills(interactive) {
  const s = interactive ? p.spinner() : null;
  if (s) s.start(msg.step2Spinner); else console.log(msg.step2Spinner);

  try {
    if (await skillsAlreadyInstalled()) {
      if (s) s.stop(msg.step2Skip); else console.log(msg.step2Skip);
      return;
    }
    await runSilentAsync('npx', ['-y', 'skills', 'add', SKILLS_REPO, '-y', '-g'], { timeout: 600000 });
    if (s) s.stop(msg.step2Done); else console.log(msg.step2Done);
  } catch (err) {
    const line = fmt(msg.step2Fail, SKILLS_REPO);
    if (s) s.stop(line); else console.error(line);

    const isTimeout = !!(err && (err.code === 'ETIMEDOUT' || err.killed));
    const details = [];
    if (err && err.code) {
      details.push(`  code: ${err.code}`);
    } else if (err && err.signal) {
      details.push(`  signal: ${err.signal}${err.killed ? ' (killed)' : ''}`);
    }
    const stderrTail = err && err.stderr ? err.stderr.toString().trim().slice(-2048) : '';
    const stdoutTail = err && err.stdout ? err.stdout.toString().trim().slice(-2048) : '';
    if (stderrTail) details.push(`  stderr (tail):\n${stderrTail}`);
    if (stdoutTail) details.push(`  stdout (tail):\n${stdoutTail}`);
    if (details.length) {
      const blob = details.join('\n');
      if (interactive) p.log.error(blob); else console.error(blob);
    }

    if (isTimeout) {
      const hint = `Timed out after 10 min. Likely cause: slow / proxied connection to github.com during git clone.\nRetry outside the wizard: npx skills add ${SKILLS_REPO} -y -g`;
      if (interactive) p.log.warn(hint); else console.error(hint);
    }

    process.exit(1);
  }
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

async function main() {
  const interactive = !!process.stdin.isTTY && !!process.stdout.isTTY;

  if (interactive) {
    p.intro(msg.setup);
    await stepInstallGlobally(true);
    await stepInstallSkills(true);
    p.outro(msg.done);
  } else {
    console.log(msg.setup);
    await stepInstallGlobally(false);
    await stepInstallSkills(false);
    console.log(msg.nonTtyHint);
  }
}

main().catch((err) => {
  const line = 'Unexpected error: ' + (err && err.message ? err.message : err);
  try { p.cancel(line); } catch (_) { console.error(line); }
  process.exit(1);
});
