/* eslint-env node */

/*
 * This file runs in a Node context (it's NOT transpiled by Babel), so use only
 * the ES6 features that are supported by your Node version. https://node.green/
 */

// Configuration for your app
// https://v2.quasar.dev/quasar-cli-webpack/quasar-config-js

/* eslint-disable @typescript-eslint/no-var-requires */

const { configure } = require('quasar/wrappers');

const dotenv = require('dotenv');
const packageJson = require('./package.json');
const changeVariablesCssSource = require('./build/quasarVariables');
dotenv.config();

module.exports = configure(function (ctx) {
	const config = require('./config')(ctx);

	return {
		// https://v2.quasar.dev/quasar-cli-webpack/supporting-ts
		supportTS: {
			tsCheckerConfig: {
				eslint: {
					enabled: true,
					files: './src/**/*.{ts,tsx,js,jsx,vue}'
				}
			}
		},

		// https://v2.quasar.dev/quasar-cli-webpack/prefetch-feature
		preFetch: true,

		// app boot file (/src/boot)
		// --> boot files are part of "main.js"
		// https://v2.quasar.dev/quasar-cli-webpack/boot-files
		boot: config.boot,

		// https://v2.quasar.dev/quasar-cli-webpack/quasar-config-js#Property%3A-css
		css: config.css,

		// https://github.com/quasarframework/quasar/tree/dev/extras
		extras: config.extras,

		vendor: {
			remove: ['moment', '@bytetrade/ui', 'video.js']
		},

		// Full list of options: https://v2.quasar.dev/quasar-cli-webpack/quasar-config-js#Property%3A-build
		build: {
			distDir: config.build.distDir,
			vueRouterMode: config.build.vueRouterMode, // available values: 'hash', 'history'
			uglifyOptions: {
				// drop_console: true
				compress: {
					drop_console: true
				}
			},

			env: {
				PL_SERVER_URL: process.env.PL_SERVER_URL,
				PLATFORM: process.env.PLATFORM,
				DEV_PLATFORM: process.env.DEV_PLATFORM,
				APPLICATION_SUB: process.env.APPLICATION_SUB,
				APPLICATION_SUB_IS_BEX: process.env.APPLICATION_SUB == 'BEX',
				DEV_PLATFORM_BEX: process.env.DEV_PLATFORM == 'BEX',
				PLATFORM_BEX_ALL:
					process.env.DEV_PLATFORM == 'BEX' || process.env.PLATFORM == 'BEX',
				BFL_URL: process.env.URL,
				IS_PC_TEST: process.env.IS_PC_TEST,
				WS_URL: process.env.NODE_ENV === 'production' ? '' : process.env.WS_URL,
				IS_BEX: process.env.PLATFORM == 'BEX',
				APP_SERVICES: process.env.APP_SERVICES,
				CACHE_CONTROL: 'max-age=31536000, public',
				EXPIRES: 'Wed, 01 Jan 2025 00:00:00 GMT',
				APP_VERSION: packageJson.version,
				VERSIONTAG: process.env.VERSIONTAG,
				APPLICATION: process.env.APPLICATION,
				IS_DEV: process.env.NODE_ENV == 'development',
				IS_PROD: process.env.NODE_ENV == 'production',
				RSS_DEBUG_URL: process.env.RSS_DEBUG_URL,
				WISE_SUB_DOMAIN: process.env.WISE_SUB_DOMAIN,
				...config.build.env
			},
			// transpile: false,

			// Add dependencies for transpiling with Babel (Array of string/regex)
			// (from node_modules, which are by default not transpiled).
			// Applies only if "transpile" is set to true.
			// transpileDependencies: [],

			// rtl: true, // https://quasar.dev/options/rtl-support
			// preloadChunks: true,
			// showProgress: false,
			gzip: true,
			// analyze: true,
			extractCSS: true,

			// Options below are automatically set depending on the env, set them if you want to override
			// extractCSS: false,
			extendWebpack(cfg) {
				config.build.extendWebpack(cfg);
			},

			// https://v2.quasar.dev/quasar-cli-webpack/handling-webpack
			// "chain" is a webpack-chain object https://github.com/neutrinojs/webpack-chain
			chainWebpack(chain, { isClient }) {
				config.build.chainWebpack(chain, { isClient });
			},
			afterBuild(params) {
				config.build.afterBuild(ctx, params);
			},
			beforeBuild({ quasarConf }) {
				changeVariablesCssSource(config);
			},
			beforeDev: async ({ quasarConf }) => {
				changeVariablesCssSource(config);
			}
		},
		// Full list of options: https://v2.quasar.dev/quasar-cli-webpack/quasar-config-js#Property%3A-devServer
		devServer: {
			server: {
				type: 'http'
			},
			// headers: {
			// 	'Cross-Origin-Embedder-Policy': 'require-corp',
			// 	'Cross-Origin-Opener-Policy': 'same-origin'
			// },
			https: config.devServer.https,
			host: config.devServer.host,
			port: process.env.DEV_PORT,
			open: true,
			proxy: config.devServer.proxy
		},

		// https://v2.quasar.dev/quasar-cli-webpack/quasar-config-js#Property%3A-framework
		framework: {
			config: {
				// dark: false, // Boolean true/false
				dark: false, //ctx.modeName === 'capacitor' ? true : 'auto',
				capacitor: {
					backButton: false
				}
			},

			// iconSet: 'material-icons', // Quasar icon set
			// lang: 'en-US', // Quasar language pack

			// For special cases outside of where the auto-import strategy can have an impact
			// (like functional components as one of the examples),
			// you can manually specify Quasar components/directives to be available everywhere:
			//
			// components: [],
			// directives: [],

			// Quasar plugins
			plugins: ['Dialog', 'Notify', 'Loading', 'Cookies', 'Meta']
		},

		// animations: 'all', // --- includes all animations
		// https://quasar.dev/options/animations
		animations: config.animations,

		// https://v2.quasar.dev/quasar-cli-webpack/developing-ssr/configuring-ssr
		ssr: {
			pwa: false,

			// manualStoreHydration: true,
			// manualPostHydrationTrigger: true,

			prodPort: 3000, // The default port that the production server should use
			// (gets superseded if process.env.PORT is specified at runtime)

			maxAge: 1000 * 60 * 60 * 24 * 30,
			// Tell browser when a file from the server should expire from cache (in ms)

			// chainWebpackWebserver (/* chain */) {},

			middlewares: [
				ctx.prod ? 'compression' : '',
				'render' // keep this as last one
			]
		},

		// https://v2.quasar.dev/quasar-cli-webpack/developing-pwa/configuring-pwa
		pwa: config.pwa,

		// Full list of options: https://v2.quasar.dev/quasar-cli-webpack/developing-cordova-apps/configuring-cordova
		cordova: {
			// noIosLegacyBuildFlag: true, // uncomment only if you know what you are doing
		},

		// Full list of options: https://v2.quasar.dev/quasar-cli-webpack/developing-capacitor-apps/configuring-capacitor
		capacitor: {
			hideSplashscreen: true
		},

		// Full list of options: https://v2.quasar.dev/quasar-cli-webpack/developing-electron-apps/configuring-electron
		electron: config.electron,
		sourceFiles: config.sourceFiles,
		htmlVariables: config.htmlVariables
	};
});
