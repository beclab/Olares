const boot = ['i18n', 'wizardUI', 'baseAxios', 'application/wizard'];
const css = ['app.scss', 'animation.scss', 'wizard.index.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpackWizard = require('./Wizard/config');
		chainWebpackWizard.build.chainWebpack(ctx, chain, { isClient });
	};
	return {
		boot,
		css,
		build: {
			env: {},
			chainWebpack
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.wizard.html'
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
