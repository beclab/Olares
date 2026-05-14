const boot = ['i18n', 'baseAxios', 'profileUI', 'application/profile'];
const css = ['profile/app.profile.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}

	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpackMobile = require('./Profile/config');
		chainWebpackMobile.build.chainWebpack(ctx, chain, { isClient });
	};

	return {
		boot,
		css,
		build: {
			env: {
				URL: process.env.URL,
				WS_URL: process.env.WS_URL
			},
			chainWebpack,
			distDir: 'dist/apps/profile-editor'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.profile.html',
			variables: 'profile/variables-editor.scss'
		},
		htmlVariables: {
			productName: 'Profile | Olares HomePage'
		}
	};
};

module.exports = {
	getConfig
};
