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
  step1Eexist:
    'Detected an existing olares-cli at /usr/local/bin/olares-cli (likely the OS bundle on a Linux Olares host).\n' +
    'npm refuses to overwrite it. Two safe workarounds:\n' +
    '  1) Side-by-side install:  npm install -g %s --prefix=$HOME/.olares-cli-npm\n' +
    '                            then add $HOME/.olares-cli-npm/bin to your PATH (before /usr/local/bin).\n' +
    '  2) One-off ops via npx:   npx %s@latest <verb>\n' +
    'See cli/README.md "On a Linux Olares host" for details.',
  preflightKeepRelease:
    'Detected release olares-cli at %s (%s); keeping it (npm copy will install side-by-side if paths differ).',
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

function semverLessThan(a, b) {
  const pa = a.replace(/-.*$/, '').split('.').map(Number);
  const pb = b.replace(/-.*$/, '').split('.').map(Number);
  for (let i = 0; i < 3; i++) {
    if ((pa[i] || 0) < (pb[i] || 0)) return true;
    if ((pa[i] || 0) > (pb[i] || 0)) return false;
  }
  return false;
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

// Classify the version string from `olares-cli --version` (Cobra default
// emits "olares-cli version X.Y.Z[-PRE]"). Returns true ONLY for what the
// release pipeline can produce:
//   - stable:     1.12.7
//   - prerelease: 1.12.8-rc1, 1.13.0-beta.1, 1.14.0-alpha2
// Returns false for everything else (treat as dev/test, safe to replace):
//   - 0.0.0-development (placeholder default in cli/version/version.go)
//   - 1.12.5-cli.2-3-gddae4ca9c, 1.12.7-rc1-3-gabc-dirty (git describe output)
//   - 1.12.7-12345678 (check.yaml PR-test build; numeric suffix isn't a tag)
//   - `dev` (Makefile no-git fallback) or any unparseable output
function isReleaseGradeVersion(verStr) {
  if (!verStr) return false;
  // Capture core MAJOR.MINOR.PATCH plus an optional pre-release suffix
  // (everything up to the next whitespace). Anchoring on whitespace/EOL
  // ensures we keep `-3-gabc-dirty` inside `pre` instead of silently dropping
  // it and mis-classifying the version as stable.
  const m = verStr.match(/version\s+v?(\d+\.\d+\.\d+)(?:-(\S+))?(?:\s|$)/);
  if (!m) return false;
  const [, mmp, pre] = m;
  if (mmp === '0.0.0') return false;
  if (!pre) return true;
  // Strictly `rc[N]`, `beta[.N]`, `alpha[.N]`, etc. Anything trailing -- a
  // git-describe `-N-gHASH`, a `-dirty` marker, or check.yaml's bare
  // `-12345678` -- makes this a dev/test build.
  return /^(rc|beta|alpha)(?:\.?\d+)?$/i.test(pre);
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
  const needsUpgrade = installedVer && latestVer && semverLessThan(installedVer, latestVer);

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
    await runSilentAsync('npm', ['install', '-g', PKG], { timeout: 120000 });
    if (s) s.stop(doneLine); else console.log(doneLine);
  } catch (err) {
    if (looksLikeEexistConflict(err)) {
      if (s) s.stop(fmt(msg.step1Fail, PKG)); else console.error(fmt(msg.step1Fail, PKG));
      const hint = fmt(msg.step1Eexist, PKG, PKG);
      if (interactive) p.log.warn(hint); else console.error(hint);
    } else {
      if (s) s.stop(fmt(msg.step1Fail, PKG)); else console.error(fmt(msg.step1Fail, PKG));
    }
    process.exit(1);
  }
}

async function skillsAlreadyInstalled() {
  try {
    const out = await runSilentAsync('npx', ['-y', 'skills', 'ls', '-g'], { timeout: 120000 });
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
    await runSilentAsync('npx', ['-y', 'skills', 'add', SKILLS_REPO, '-y', '-g'], { timeout: 120000 });
    if (s) s.stop(msg.step2Done); else console.log(msg.step2Done);
  } catch (_) {
    const line = fmt(msg.step2Fail, SKILLS_REPO);
    if (s) s.stop(line); else console.error(line);
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
