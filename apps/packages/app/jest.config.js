/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
	rootDir: __dirname,
	preset: 'ts-jest',
	testEnvironment: 'node',
	moduleFileExtensions: ['ts', 'js', 'json', 'vue'],
	transform: {
		'^.+\\.vue$': '@vue/vue3-jest',
		// Translate suites import ESM .js next to .ts; allowJs lets ts-jest
		// compile both so those tests do not need a separate babel pipeline.
		'^.+\\.[tj]s$': ['ts-jest', { tsconfig: { allowJs: true } }]
	},
	// jsdom defaults to the `browser` export condition, which resolves ESM-only
	// builds (vue-i18n, @intlify/*) that jest cannot parse.
	testEnvironmentOptions: { customExportConditions: ['node', 'require'] },
	moduleNameMapper: {
		'\\.(svg|png|jpe?g|gif|webp|ico)$': '<rootDir>/jest/assetMock.ts',
		'\\.css$': '<rootDir>/jest/assetMock.ts',
		'^webextension-polyfill-ts$':
			'<rootDir>/src/platform/interface/bex/translate/__mocks__/webextensionPolyfill.ts',
		'^webextension-polyfill$':
			'<rootDir>/src/platform/interface/bex/translate/__mocks__/webextensionPolyfill.ts',
		'^src/boot/dashboard-i18n$': '<rootDir>/jest/i18nMock.ts',
		'^@apps/control-panel-common/src/(.*)$':
			'<rootDir>/src/apps/controlPanelCommon/$1',
		'^@apps/dashboard/src/(.*)$': '<rootDir>/src/apps/dashboard/$1',
		'^@apps/profile/src/(.*)$': '<rootDir>/src/$1',
		'^src/(.*)$': '<rootDir>/src/$1'
	},
	setupFiles: ['<rootDir>/src/platform/interface/bex/translate/jest.setup.js'],
	/**
	 * The crawler walks everything under `rootDir` and every worker carries its
	 * own copy of the result. `dist/` is 79k files of build output including a
	 * nested `node_modules` and a `package.json` duplicating this package's name,
	 * which is also where the "Haste module naming collision: @didvault/app"
	 * warning came from. `src-capacitor/` is a further 23k files of iOS/Android
	 * sources that no suite can run.
	 */
	modulePathIgnorePatterns: ['<rootDir>/src-capacitor/', '<rootDir>/dist/'],
	watchPathIgnorePatterns: ['<rootDir>/dist/'],
	/**
	 * ts-jest type-checks every suite, so a worker holds a TypeScript program on
	 * top of the module registry — measured at ~1.5GB on the store suites, and it
	 * keeps climbing from one file to the next rather than settling. Unbounded,
	 * jest runs `cpus - 1` of those at once and takes the machine's memory with
	 * it.
	 *
	 * The cap bounds how many can exist; the idle limit is what answers the
	 * growth, retiring a worker once it crosses the threshold. That check only
	 * happens between suites, so at worst this costs restart time — it cannot
	 * fail a suite that genuinely needs the headroom.
	 */
	maxWorkers: '50%',
	workerIdleMemoryLimit: '1GB',
	testMatch: [
		'<rootDir>/src/**/*.test.ts',
		'<rootDir>/tests/unit/**/*.spec.ts'
	]
};
