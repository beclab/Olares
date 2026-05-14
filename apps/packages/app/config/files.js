const boot = ['i18n', 'filesUI', 'baseAxios', 'application/files'];
const css = ['files/app.scss'];

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
			https: process.env.PROTOCOL === 'https://'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.files.html',
			variables: 'files/variables.scss'
		},
		htmlVariables: {
			productName: 'Files'
		}
	};
};

module.exports = {
	getConfig
};
