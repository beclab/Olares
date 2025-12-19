// @ts-ignore
const globalConfig = {
	url:
		process.env.NODE_ENV === 'development' ? '' : process.env.PUBLIC_URL || '',
	isOfficial: !!(process.env.PUBLIC_URL && process.env.PUBLIC_URL.length > 0),
	install_en_docs: 'https://docs.olares.com/manual/get-started/?redirect=false',
	install_zh_docs:
		'https://docs.olares.com/zh/manual/get-started/?redirect=false'
};

export default globalConfig;
