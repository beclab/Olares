const PreloadWebpackPlugin = require('preload-webpack-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
const TerserPlugin = require('terser-webpack-plugin');
const path = require('path');
const proxyDefault = require('./proxyDefault');

let config = undefined;

const httpsApplications = [
	'WISE',
	'FILES',
	'LAREPASS',
	'LOGIN',
	'WIZARD',
	'EDITOR',
	'PREVIEW',
	'MARKET',
	'SETTINGS',
	'DESKTOP',
	'DASHBOARD',
	'CONTROL_HUB',
	'STUDIO',
	'SHARE',
	'VAULT'
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

const initConfig = (ctx) => {
	if (process.env.APPLICATION === 'WISE') {
		config = require('./wise').getConfig(ctx);
	} else if (process.env.APPLICATION === 'SETTINGS') {
		config = require('./settings').getConfig(ctx);
	} else if (process.env.APPLICATION === 'LAREPASS') {
		config = require('./LarePass').getConfig(ctx);
	} else if (process.env.APPLICATION === 'FILES') {
		config = require('./files').getConfig(ctx);
	} else if (process.env.APPLICATION === 'VAULT') {
		config = require('./vault').getConfig(ctx);
	} else if (process.env.APPLICATION === 'LOGIN') {
		config = require('./login').getConfig(ctx);
	} else if (process.env.APPLICATION === 'WIZARD') {
		config = require('./wizard').getConfig(ctx);
	} else if (process.env.APPLICATION === 'EDITOR') {
		config = require('./editor').getConfig(ctx);
	} else if (process.env.APPLICATION === 'PREVIEW') {
		config = require('./preview').getConfig(ctx);
	} else if (process.env.APPLICATION === 'MARKET') {
		config = require('./market').getConfig(ctx);
	} else if (process.env.APPLICATION === 'DESKTOP') {
		config = require('./desktop').getConfig(ctx);
	} else if (process.env.APPLICATION === 'DASHBOARD') {
		config = require('./dashboard').getConfig(ctx);
	} else if (process.env.APPLICATION === 'CONTROL_HUB') {
		config = require('./control-hub').getConfig(ctx);
	} else if (process.env.APPLICATION === 'STUDIO') {
		config = require('./studio.js').getConfig(ctx);
	} else if (process.env.APPLICATION === 'SHARE') {
		config = require('./share').getConfig(ctx);
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
				...(htmlPlugin.options.excludeChunks || []),
				'translate-content'
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
		...appsAlias('studio', 'studio'),
		'@apps/control-panel-common/src': path.resolve(
			__dirname,
			'../src/apps/controlPanelCommon'
		),
		'@apps/desktop': path.resolve(__dirname, '../src/apps/Desktop')
	};
};

const chainWebpack = (ctx, chain, { isClient }) => {
	const nodePolyfillWebpackPlugin = require('node-polyfill-webpack-plugin');
	chain.plugin('node-polyfill').use(nodePolyfillWebpackPlugin);
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
			chain.plugin('terser').use(TerserPlugin, [
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

		// chain.optimization.minimizer('terser').use(TerserPlugin, [
		// 	{
		// 		terserOptions: {
		// 			parallel: true,
		// 			sourceMap: true,
		// 			extractComments: false,
		// 			compress: {
		// 				drop_console: true,
		// 				drop_debugger: true,
		// 				pure_funcs: ['console.log']
		// 			},
		// 			output: {
		// 				comments: false,
		// 				ascii_only: true
		// 			}
		// 		}
		// 	}
		// ]);

		chain.optimization.splitChunks({
			chunks: 'all', // The type of chunk that requires code segmentation
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
					: httpsApplications.includes(process.env.APPLICATION),
			host:
				config?.devServer?.host != undefined
					? config.devServer.host
					: httpsApplications.includes(process.env.APPLICATION)
					? process.env.DEV_DOMAIN
					: 'localhost'
		}
	};
};
