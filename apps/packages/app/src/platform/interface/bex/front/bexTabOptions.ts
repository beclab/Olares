export const collectIdentify = 'collect';
export const subscribeIdentify = 'subscribe';

export const tabsIgnore = [collectIdentify, subscribeIdentify];

export const tabbarItems = [
	{
		name: 'home',
		identify: 'home',
		normalImage: 'home',
		hoverImage: 'home-hover',
		activeImage: 'home-highlight',
		to: '/home'
	},
	{
		name: 'collect',
		identify: collectIdentify,
		normalImage: 'collect',
		hoverImage: 'collect-hover',
		activeImage: 'collect-highlight',
		to: '/collect'
	},
	{
		name: 'translate',
		identify: 'translate',
		normalImage: 'translate',
		hoverImage: 'translate-hover',
		activeImage: 'translate-highlight',
		to: '/translate'
	},
	{
		name: 'vault',
		identify: 'secret',
		normalImage: 'vault',
		hoverImage: 'vault-hover',
		activeImage: 'vault-highlight',
		to: '/items'
	},
	{
		name: 'application',
		identify: 'application',
		normalImage: 'application',
		hoverImage: 'application-hover',
		activeImage: 'application-highlight',
		to: '/application'
	}
	// {
	// 	name: 'setting',
	// 	identify: 'setting',
	// 	normalImage: 'tab_setting_normal',
	// 	activeImage: 'tab_setting_active',
	// 	to: '/setting'
	// }
	// {
	// 	name: 'sidepanel',
	// 	identify: 'sidepanel',
	// 	normalImage: 'vault',
	// 	activeImage: 'vault-highlight',
	// 	to: '/sidepanel'
	// }
];
