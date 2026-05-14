import MenuIcon from './menu.svg';
import MenuActiveIcon from './menu_active.svg';
export const data = [
	{
		key: 'sub1',
		label: 'Navigation One',
		children: [
			{
				key: '5',
				label: 'Shared Libraries',
				icon: 'home',
				disable: true,
				children: [
					{
						key: '5-1',
						label: 'Option 5-1',
						img: MenuIcon,
						activeImg: MenuActiveIcon,
						count: 10,
						defaultHide: true
					},
					{
						key: '5-2',
						label: 'Option 5-2',
						icon: 'home'
						// muted: true,
						// defaultHide: true
					}
				]
			},
			{
				key: '6',
				label: 'Option 6',
				icon: 'home'
			},
			{
				key: '7',
				label: 'Option 7',
				icon: 'home',
				disable: true
			}
		]
	},
	{
		key: 'sub2',
		label: 'Navigation Two',
		children: [
			{
				key: '10',
				label: 'Option 10',
				icon: 'home'
			},
			{
				key: 'sub3',
				label: 'Submenu',
				icon: 'home',
				children: [
					{
						key: '11',
						label: 'Option 11',
						icon: 'home'
					},
					{
						key: '12',
						label: 'Option 12',
						icon: 'home'
					}
				]
			}
		]
	}
];

export const rssData = [
	{
		label: 'Favorites',
		key: 'Favorites',
		children: [
			{
				label: 'Trend',
				key: 1,
				icon: 'sort_by_alpha'
			},
			{
				label: 'Feeds',
				key: 2,
				icon: 'rss_feed'
			},
			{
				label: 'Library',
				key: 3,
				icon: 'library_books'
			}
		]
	},
	{
		label: 'Utility',
		key: 'Utility',
		children: [
			{
				label: 'Add',
				key: 5,
				icon: 'add_box',
				muted: true
			},
			{
				label: 'Settings',
				key: 4,
				icon: 'settings_applications'
			}
		]
	}
];
