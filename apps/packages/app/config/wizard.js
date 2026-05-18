const boot = ['i18n', 'wizardUI', 'baseAxios', 'application/wizard'];
const css = ['wizard/app.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpackWizard = require('./Wizard/config');
		chainWebpackWizard.build.chainWebpack(ctx, chain, { isClient });
	};

	const proxy = require('./Wizard/proxy');
	return {
		boot,
		css,
		build: {
			env: {},
			chainWebpack
		},
		devServer: {
			proxy: proxy,
			host: process.env.DEV_DOMAIN,
			https: true
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.wizard.html',
			variables: 'wizard/variables.scss'
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
