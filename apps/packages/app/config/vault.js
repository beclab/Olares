const proxyDefault = require('./proxyDefault');

const boot = ['i18n', 'baseAxios', 'smartEnginEntrance', 'application/vault'];
const css = ['app.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpack = require('./Vault/index');
		chainWebpack.build.chainWebpack(ctx, chain, { isClient });
	};
	return {
		boot,
		css,
		build: {
			env: {},
			chainWebpack,
			distDir: 'dist/apps/vault'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.vault.html'
		},
		htmlVariables: {
			productName: 'Vault'
		},
		devServer: {
			proxy: {
				'/bfl/info/v1/olares-info': {
					target: `https://vault.${process.env.ACCOUNT_DOMAIN}`,
					changeOrigin: true
				},
				...proxyDefault,
				'/server': {
					target: `https://vault.${process.env.ACCOUNT_DOMAIN}`,
					changeOrigin: true
				}
			}
		}
	};
};

module.exports = {
	getConfig
};
