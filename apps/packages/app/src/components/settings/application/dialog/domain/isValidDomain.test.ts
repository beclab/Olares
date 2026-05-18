// isValidDomain.test.ts

import { isValidDomain } from 'src/components/settings/application/dialog/domain/isValidDomain';

describe('isValidDomain', () => {
	const valids = [
		// Simple domain names
		'a.com',
		'abc.com',
		'foo-bar.com',
		'foo123.com',
		'foo.bar',
		'foo.bar.baz',

		// Containing numbers / short labels
		'a.b',
		'a1.b2',
		'abc-123.xyz',
		'test.byte-trade.com',

		// Max label length (63): 1 letter + 61 middle chars + 1 ending char
		`a${'a'.repeat(61)}b.com`, // 63-character label + .com

		// Near maximum total length (253)
		'a.'.repeat(50).slice(0, -1) + 'com'
	];

	const invalids = [
		// Empty / Excessively long
		'',
		' '.repeat(10),
		// Total length > 253
		`${'a'.repeat(64)}.`.repeat(5) + 'com',

		// Label starts with number (violates "starts with letter" rule)
		'1abc.com',
		'foo.1bar.com',

		// Label starts/ends with hyphen
		'-abc.com',
		'abc-.com',
		'foo.-bar.com',
		'foo.bar-.com',

		// Contains invalid characters
		'abc_.com',
		'ab$c.com',
		'foo@bar.com',
		'foo,bar.com',
		'foo+bar.com',

		// Uppercase
		'ABC.com',
		'Foo-Bar.COM',

		// Consecutive dots / empty labels
		'a..b.com',
		'.abc.com',
		'abc.com.',
		'...',

		// Overlength label: 64 characters
		`${'a'.repeat(64)}.com`,
		`foo.${'b'.repeat(64)}.com`,

		// Single dot / invalid structure
		'.',
		'com.',
		'.com',
		'a..'
	];

	test.each(valids)('valid: %s', (domain) => {
		expect(isValidDomain(domain)).toBe(true);
	});

	test.each(invalids)('invalid: %s', (domain) => {
		expect(isValidDomain(domain)).toBe(false);
	});
});
