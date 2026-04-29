#!/usr/bin/env node
// golden-gen.js — drive @bytetrade/core/src/monitoring.ts directly to
// produce a JSON oracle the Go format package compares against.
//
// Run from this directory:
//
//   cd cli/cmd/ctl/dashboard/format/testdata
//   node golden-gen.js > golden.json
//
// Updating the JS source means re-running this script and committing the
// resulting golden.json. The Go test loads the file from this directory
// and asserts byte-equality.
//
// Why a Node.js oracle (vs. hand-curated cases)? Because the upstream's
// edge-case handling (BigNumber.toPrecision quirks, parseFloat coercion,
// toFixed half-to-even pitfalls) is too easy to drift from accidentally.
// The fixture pins the *exact* JS output so any future change in the
// upstream surfaces as a Go test failure with a diff.
//
// IMPORTANT: This file relies on an installed `@bytetrade/core`. If you
// don't have one, `npm install @bytetrade/core` from the repo root before
// running.

const fs = require('fs');
const path = require('path');

let core;
try {
	// Prefer the live workspace dependency.
	core = require('@bytetrade/core');
} catch (e1) {
	// Fall back to a relative path used by the dashboard SPA — useful when
	// running this script from a TermiPass clone that vendors the source.
	const guesses = [
		path.resolve(__dirname, '../../../../../../../../Desktop/larepass/TermiPass/packages/app/dist/electron/UnPackaged/node_modules/@bytetrade/core/src/monitoring.ts'),
	];
	let lastErr = e1;
	for (const g of guesses) {
		try {
			core = require(g);
			break;
		} catch (e) {
			lastErr = e;
		}
	}
	if (!core) {
		console.error('cannot load @bytetrade/core; install it via `npm i @bytetrade/core` or run from a TermiPass workspace.');
		console.error('underlying:', lastErr && lastErr.message);
		process.exit(2);
	}
}

const { getValueByUnit, getSuitableValue, getSuitableUnit } = core;

// Mirror dashboard utils for worthValue / formatFrequency. We can't
// require them directly (they live in src/apps/dashboard/utils/, not in
// @bytetrade/core), so we re-implement here against the same primitives
// the upstream uses. Any drift surfaces as a golden mismatch — the
// resulting test failure tells the reviewer to bring this script in
// line with the upstream they just modified.
let BigNumber;
try {
	BigNumber = require('bignumber.js');
} catch (e) {
	console.error('install bignumber.js (`npm i bignumber.js`) before running');
	process.exit(2);
}

function worthValue(v) {
	const t = new BigNumber(v);
	let k = 4;
	if (t.isGreaterThanOrEqualTo(new BigNumber('1000'))) k = 0;
	if (t.isGreaterThanOrEqualTo(new BigNumber('100'))) k = 1;
	if (t.isGreaterThanOrEqualTo(new BigNumber('10'))) k = 2;
	if (t.isGreaterThanOrEqualTo(new BigNumber('1'))) k = 3;
	return new BigNumber(t.toPrecision(k)).toFormat();
}

function formatFrequency(value, fromUnit = 'Hz') {
	if (value === 0) return '0 Hz';
	const units = ['Hz', 'kHz', 'MHz', 'GHz'];
	const unitIndex = units.indexOf(fromUnit);
	let frequency = value;
	let currentUnitIndex = unitIndex;
	while (frequency >= 1000 && currentUnitIndex < units.length - 1) {
		frequency /= 1000;
		currentUnitIndex++;
	}
	while (frequency < 1 && currentUnitIndex > 0) {
		frequency *= 1000;
		currentUnitIndex--;
	}
	return `${Math.round(frequency * 100) / 100}${units[currentUnitIndex]}`;
}

// Cases mirror the manual table in format_test.go but emit the *actual*
// JS output, so tweaks here propagate through regeneration.
const cases = {
	getValueByUnit: [
		{ num: '12.345', unit: '', precision: 2 },
		{ num: '12.345', unit: 'default', precision: 2 },
		{ num: 'NAN', unit: 'Bytes', precision: 2 },
		{ num: 'abc', unit: 'Bytes', precision: 2 },
		{ num: '0.4567', unit: '%', precision: 2 },
		{ num: '0.123', unit: 'm', precision: 2 },
		{ num: '0.0001', unit: 'm', precision: 2 },
		{ num: '2048', unit: 'Ki', precision: 2 },
		{ num: '1048576', unit: 'Mi', precision: 2 },
		{ num: '1073741824', unit: 'Gi', precision: 2 },
		{ num: '1099511627776', unit: 'Ti', precision: 2 },
		{ num: '1024', unit: 'Bytes', precision: 2 },
		{ num: '1500', unit: 'K', precision: 2 },
		{ num: '1500000', unit: 'M', precision: 2 },
		{ num: '1500000000', unit: 'G', precision: 2 },
		{ num: '1500000000000', unit: 'T', precision: 2 },
		{ num: '100', unit: 'bps', precision: 2 },
		{ num: '1024', unit: 'Kbps', precision: 2 },
		{ num: '131072', unit: 'Mbps', precision: 2 },
		{ num: '0.1234', unit: 'ms', precision: 2 },
		{ num: '12.7', unit: 'iops', precision: 2 },
		{ num: '0.45', unit: 'fishtacos', precision: 2 },
		{ num: '0', unit: 'Mi', precision: 2 },
		{ num: '1.234567', unit: 'Bytes', precision: 2 },
	].map((c) => ({ ...c, want: getValueByUnit(c.num, c.unit, c.precision) })),

	getSuitableValue: [
		{ value: String(1.5 * 1024 * 1024 * 1024), unitType: 'memory' },
		{ value: '1024', unitType: 'memory' },
		{ value: '100', unitType: 'memory' },
		{ value: '0.05', unitType: 'cpu' },
		{ value: '1500', unitType: 'throughput' },
		{ value: 'not a number', unitType: 'memory' },
		{ value: '5', unitType: 'number' },
	].map((c) => ({ ...c, want: getSuitableValue(c.value, c.unitType, '0') })),

	worthValue: [
		'1234.5678',
		'1.23456',
		'0.001234',
		'0.5',
		'0',
		'-1234',
		'abc',
	].map((input) => ({ input, want: worthValue(input) })),

	formatFrequency: [
		{ value: 0, from: 'Hz' },
		{ value: 1500, from: 'Hz' },
		{ value: 3.5e9, from: 'Hz' },
		{ value: 0.5, from: 'kHz' },
		{ value: 1500, from: '' },
		{ value: 100, from: 'blub' },
		{ value: 2500, from: 'Hz' },
	].map((c) => ({ ...c, want: formatFrequency(c.value, c.from || 'Hz') })),
};

const out = path.join(__dirname, 'golden.json');
fs.writeFileSync(out, JSON.stringify(cases, null, 2) + '\n');
console.error('wrote', out);
