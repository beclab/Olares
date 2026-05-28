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

// ---------------------------------------------------------------------------
// Steps
// ---------------------------------------------------------------------------

async function stepInstallGlobally(interactive) {
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
