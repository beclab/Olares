const boot = ['i18n', 'loginUI', 'baseAxios', 'application/login'];
const css = ['app.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}

	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpackMobile = require('./Login/config');
		chainWebpackMobile.build.chainWebpack(ctx, chain, { isClient });
	};
	const proxy = require('./Login/proxy');
	return {
		boot,
		css,
		build: {
			env: {
				URL: process.env.URL
			},
			chainWebpack
		},
		devServer: {
			proxy: proxy,
			host: process.env.DEV_DOMAIN,
			https: true
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.login.html'
		},
		htmlVariables: {
			productName: 'Olares'
		},
		animations: ['fadeIn', 'fadeOut']
	};
};

module.exports = {
	getConfig
};
