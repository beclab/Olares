const boot = ['i18n', 'settingsUI', 'baseAxios', 'application/settings'];
const css = ['app.scss', 'adaptive.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}

	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpackMobile = require('./Settings/config');
		chainWebpackMobile.build.chainWebpack(ctx, chain, { isClient });
	};
	return {
		boot,
		css,
		build: {
			env: {
				URL: process.env.URL,
				OLARES_SPACE_URL: process.env.OLARES_SPACE_URL,
				ACTION: process.env.ACTION,
				NODE_RPC: 'https://mainnet.optimism.io',
				CONTRACT_DID: '0x5da4fa8e567d86e52ef8da860de1be8f54cae97d',
				CONTRACT_ROOT_RESOLVER: '0xe2eaba0979277a90511f8873ae1e8ca26b54e740',
				CONTRACT_REGISTRY: '0x5da4fa8e567d86e52ef8da860de1be8f54cae97d',
				DEMO: process.env.DEMO,
				SETTINGS_URL: process.env.SETTINGS_URL
			},
			chainWebpack,
			distDir: 'dist/apps/settings'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.settings.html'
		},
		htmlVariables: {
			productName: 'Settings'
		}
	};
};

module.exports = {
	getConfig
};
