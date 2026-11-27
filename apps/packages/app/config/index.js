const PreloadWebpackPlugin = require('preload-webpack-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
const TerserPlugin = require('terser-webpack-plugin');
const path = require('path');
const proxyDefault = require('./proxyDefault');
const { splitStyleChunks } = require('../build/webpack/styleChunks');

let config = undefined;

const httpsApplications = [
	'WISE',
	'FILES',
	'WIZARD',
	'EDITOR',
	'MARKET',
	'SETTINGS',
	'DESKTOP',
	'DASHBOARD',
	'CONTROL_HUB'
];

const aliasKeys = [
	'src',
	'components',
	'layouts',
	'pages',
	'assets',
	'boot',
	'stores'
];
const appsAlias = (appName, fileName) => {
	const obj = {};
	aliasKeys.forEach((key) => {
		if (key === 'src') {
			obj[`@apps/${appName}${key ? '/' + key : ''}`] = path.resolve(
				__dirname,
				`../src/apps/${fileName}`
			);
		} else {
			obj[`@apps/${appName}${key ? '/' + key : ''}`] = path.resolve(
				__dirname,
				`../src/apps/${fileName}${key ? '/' + key : ''}`
			);
		}
	});
	return obj;
};

// Webpack entry chunks that must not be injected into the main BEX index.html.
// Overlay pages (button/list/bar) ship their own HTML; content scripts are
// loaded via manifest.json. Including their CSS (e.g. list.css sets html
// font-size: 10px) would break Side Panel typography.
const BEX_INDEX_EXCLUDE_CHUNKS = [
	'translate-content',
	'TermiPassProvider',
	'rssContentScript',
	'sidePanelContentScript',
	'content/trigger-autofill-script-injection',
	'content/bootstrap-autofill',
	'content/bootstrap-autofill-overlay',
	'content/autofiller',
	'content/notificationBar',
	'button',
	'list',
	'bar'
];

const initConfig = (ctx) => {
	if (process.env.APPLICATION === 'WISE') {
		config = require('./wise').getConfig(ctx);
	} else if (process.env.APPLICATION === 'SETTINGS') {
		config = require('./settings').getConfig(ctx);

	} else if (process.env.APPLICATION === 'FILES') {
		config = require('./files').getConfig(ctx);
	} else if (process.env.APPLICATION === 'WIZARD') {
		config = require('./wizard').getConfig(ctx);
	} else if (process.env.APPLICATION === 'EDITOR') {
		config = require('./editor').getConfig(ctx);
	} else if (process.env.APPLICATION === 'MARKET') {
		config = require('./market').getConfig(ctx);
	} else if (process.env.APPLICATION === 'DESKTOP') {
		config = require('./desktop').getConfig(ctx);
	} else if (process.env.APPLICATION === 'DASHBOARD') {
		config = require('./dashboard').getConfig(ctx);
	} else if (process.env.APPLICATION === 'CONTROL_HUB') {
		config = require('./control-hub').getConfig(ctx);
	} else {
		throw new Error('APPLICATION environment variable is not set or invalid');
	}
};

const extendWebpack = (ctx, cfg) => {
	const isBex = ctx.modeName === 'bex';
	if (isBex) {
		const htmlPlugin = cfg.plugins.find(
			(p) => p.constructor.name === 'HtmlWebpackPlugin'
		);
		if (htmlPlugin) {
			htmlPlugin.options.excludeChunks = [
				...new Set([
					...(htmlPlugin.options.excludeChunks || []),
					...BEX_INDEX_EXCLUDE_CHUNKS
				])
			];
		}
	}
	!ctx.dev &&
		cfg.plugins.push(
			new PreloadWebpackPlugin({
				rel: 'preload',
				include: 'allAssets',
				fileWhitelist: [/.+MaterialSymbolsRounded.+/],
				as: 'font'
			})
		);

	cfg.resolve.fallback = {
		fs: false,
		tls: false,
		net: false
	};

	cfg.resolve.alias = {
		...cfg.resolve.alias,
		'@apps/profile/src': path.resolve(__dirname, '../src'),
		'@apps/market/src': path.resolve(__dirname, '../src')
	};

	if (config && config.build.extendWebpack) {
		config.build.extendWebpack(ctx, cfg);
	}

	cfg.resolve.alias = {
		...cfg.resolve.alias,
		...appsAlias('dashboard', 'dashboard'),
		...appsAlias('control-hub', 'controlHub'),
		...appsAlias('control-panel-common', 'controlPanelCommon'),
		'@apps/control-panel-common/src': path.resolve(
			__dirname,
			'../src/apps/controlPanelCommon'
		),
		'@apps/desktop': path.resolve(__dirname, '../src/apps/Desktop'),
		'@bytetrade/ui': path.resolve(__dirname, '../src/components/ui/index.ts')
	};
};

const chainWebpack = (ctx, chain, { isClient }) => {
	const nodePolyfillWebpackPlugin = require('node-polyfill-webpack-plugin');
	chain.plugin('node-polyfill').use(nodePolyfillWebpackPlugin);

	// webpack 5 treats packages with "type": "module" as ESM with fullySpecified=true,
	// which prevents resolving bare directory imports (e.g. "./avatar" -> "./avatar/index.js").
	chain.module
		.rule('esm-fullyspecified')
		.test(/\.m?js$/)
		.resolve.set('fullySpecified', false);

	const isBex = ctx.modeName === 'bex';

	if (isClient && !isBex) {
		chain.plugin('css-minimizer-webpack-plugin').use(CssMinimizerPlugin, [
			{
				parallel: true,
				minimizerOptions: {
					preset: [
						'default',
						{
							mergeLonghand: true,
							cssDeclarationSorter: 'concentric',
							discardComments: { removeAll: true }
						}
					]
				}
			}
		]);

		if (!ctx.dev) {
			chain.optimization.minimizer('terser').use(TerserPlugin, [
				{
					terserOptions: {
						// parallel: true,
						sourceMap: true,
						// extractComments: false,
						compress: {
							drop_console: true,
							drop_debugger: true,
							pure_funcs: ['console.log']
						},
						output: {
							comments: false,
							ascii_only: true
						}
					}
				}
			]);
		}

		chain.optimization.splitChunks({
			// Use a function instead of 'all': non-worker chunks behave exactly like
			// 'all', but the shared websocket worker chunk is excluded from splitting so
			// it stays self-contained. Otherwise its vendor deps (e.g. util) get hoisted
			// into separate chunks that the worker must importScripts() at startup, which
			// throws "Failed to execute 'importScripts'" when that request fails.
			// The worker chunk is given the stable name "sharedWebsocketWorker" via a
			// webpackEntryOptions magic comment at its `new SharedWorker(new URL(...))`
			// call site; without that name webpack exposes it here with name===undefined
			// and it cannot be matched.
			chunks(chunk) {
				return !/sharedWebsocketWorker/i.test(chunk.name || '');
			},
			minSize: 20000, // Minimum split file size
			minRemainingSize: 0, // Minimum remaining file size after segmentation
			minChunks: 1, // The number of times it has been referenced before it is split
			maxAsyncRequests: 30, // Maximum number of asynchronous requests
			maxInitialRequests: 30, // Maximum number of initialization requests
			enforceSizeThreshold: 50000,
			cacheGroups: {
				// Cache Group configuration
				defaultVendors: {
					test: /[\\/]node_modules[\\/]/,
					priority: -10,
					reuseExistingChunk: true
				},
				default: {
					minChunks: 2,
					priority: -20,
					reuseExistingChunk: true //	Reuse the chunk that has been split
				}
			}
		});
		if (ctx.prod) {
			if (process.env.PLATFORM === 'MOBILE') {
				// Mobile routes can contain the same CSS modules in different orders.
				// Keep those route chunks distinct; merging them reintroduces CSS order
				// conflicts before splitChunks shares their JavaScript dependencies.
				// This duplicates some route CSS; vendor styles remain shared.
				chain.optimization.mergeDuplicateChunks(false);
			}
			chain.optimization.splitChunks(
				splitStyleChunks(chain.optimization.get('splitChunks'))
			);
		}
	}
	if (config && config.build.chainWebpack) {
		config.build.chainWebpack(ctx, chain, { isClient });
	}
};

const afterBuild = (ctx, params) => {
	if (config && config.build.afterBuild) {
		config.build.afterBuild(ctx, params);
	}
};

const electron = () => {
	if (config && config.electron) {
		return config.electron;
	}
};

const pwa = () => {
	if (config && config.pwa) {
		return config.pwa;
	}
};

module.exports = (ctx) => {
	initConfig(ctx);

	return {
		boot: config?.boot || [],
		css: ['base-font.css', ...(config?.css || [])],
		extras: config?.extras || [
			'material-icons',
			'bootstrap-icons',
			'roboto-font',
			ctx.dev ? 'material-symbols-rounded' : ''
		],
		build: {
			vueRouterMode: 'history',
			env: config?.build.env || {},
			extendWebpack: (cfg) => extendWebpack(ctx, cfg),
			chainWebpack: (chain, { isClient }) =>
				chainWebpack(ctx, chain, { isClient }),
			afterBuild: (params) => afterBuild(ctx, params),
			distDir: config.build.distDir || undefined
		},
		animations: config?.animations || [],
		pwa: pwa(),
		electron: electron(),
		sourceFiles: config?.sourceFiles || {},
		htmlVariables: config?.htmlVariables || {},
		devServer: {
			proxy: config?.devServer?.proxy || proxyDefault,
			https:
				config?.devServer?.https != undefined
					? config.devServer.https
					: process.env.PROTOCOL === 'https://',
			host:
				config?.devServer?.host != undefined
					? config.devServer.host
					: httpsApplications.includes(process.env.APPLICATION)
					? process.env.DEV_DOMAIN
					: 'localhost'
		}
	};
};
