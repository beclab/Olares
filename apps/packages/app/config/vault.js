const boot = ['i18n', 'baseAxios', 'smartEnginEntrance', 'application/vault'];
const css = ['vault/app.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpack = require('./Vault/index');
		chainWebpack.build.chainWebpack(ctx, chain, { isClient });
	};

	const proxy = require('./Vault/proxy');

	return {
		boot,
		css,
		build: {
			env: {},
			chainWebpack,
			distDir: 'dist/apps/vault'
		},
		devServer: {
			proxy: proxy,
			host: process.env.DEV_DOMAIN,
			https: process.env.PROTOCOL === 'https://'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.vault.html',
			variables: 'vault/variables.scss'
		},
		htmlVariables: {
			productName: 'Vault'
		}
		// devServer: {
		// 	proxy: {
		// 		'/bfl/info/v1/olares-info': {
		// 			target: `https://vault.${process.env.ACCOUNT_DOMAIN}`,
		// 			changeOrigin: true
		// 		},
		// 		...proxyDefault,
		// 		'/server': {
		// 			target: `https://vault.${process.env.ACCOUNT_DOMAIN}`,
		// 			changeOrigin: true
		// 		}
		// 	}
		// }
	};
};

module.exports = {
	getConfig
};
