const boot = ['i18n', 'smartEnginEntrance', 'baseAxios', 'application/files'];
const css = ['app.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	const proxy = require('./Files/proxy');
	return {
		boot,
		css,
		build: {
			env: {},
			distDir: 'dist/apps/files'
		},
		devServer: {
			proxy: proxy,
			host: process.env.DEV_DOMAIN,
			https: true
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.files.html'
		},
		htmlVariables: {
			productName: 'Files'
		}
	};
};

module.exports = {
	getConfig
};
