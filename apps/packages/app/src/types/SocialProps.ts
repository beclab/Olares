import { Social } from 'src/types/User';

export enum SOCIAL_TYPE {
	TWITTER = 'Twitter',
	FACEBOOK = 'Facebook',
	INSTAGRAM = 'Instagram',
	TIKTOK = 'Tiktok',
	GITHUB = 'Github',
	CLUBHOUSE = 'Clubhouse',
	PATREON = 'Patreon',
	VENMO = 'Venmo',
	CASH_APP = 'CashApp',
	DEEZER = 'Deezer',
	PANDORA = 'Pandora',
	TIDAL = 'Tidal',
	BANDCAMP = 'Bandcamp',
	MEDIUM = 'Medium',
	BEEHIIV = 'Beehiiv',
	FLIPBOARD = 'Flipboard',
	GOODREADS = 'Goodreads',
	MASTODON = 'Mastodon',
	STEAM = 'Steam',
	TUMBLR = 'Tumblr'
}

// YOUTUBE = 'Youtube',
// 	LINKEDIN = 'Linkedin',
// 	PINTEREST = 'Pinterest',
// 	SNAPCHAT = 'Snapchat',
// 	SPOTIFY = 'Spotify',
// 	THREADS = 'Threads',
// 	REDDIT = 'Reddit',
// 	TWITCH = 'Twitch',
// 	DRIBBBLE = 'Dribbble',
// 	DISCORD = 'Discord',
// 	SUBSTACK = 'Substack',
// 	TELEGRAM = 'Telegram',

export const SocialMap: Record<string, Social> = {
	Instagram: {
		platform: SOCIAL_TYPE.INSTAGRAM,
		username: '',
		url: 'https://instagram.com/${username}'
	},
	Facebook: {
		platform: SOCIAL_TYPE.FACEBOOK,
		username: '',
		url: 'https://facebook.com/${username}'
	},
	Twitter: {
		platform: SOCIAL_TYPE.TWITTER,
		username: '',
		url: 'https://twitter.com/${username}'
	},
	Tiktok: {
		platform: SOCIAL_TYPE.TIKTOK,
		username: '',
		url: 'https://tiktok.com/@${username}'
	},
	Github: {
		platform: SOCIAL_TYPE.GITHUB,
		username: '',
		url: 'https://github.com/${username}'
	},
	Clubhouse: {
		platform: SOCIAL_TYPE.CLUBHOUSE,
		username: '',
		url: ''
	},
	Patreon: {
		platform: SOCIAL_TYPE.PATREON,
		username: '',
		url: 'https://patreon.com/${username}'
	},
	Venmo: {
		platform: SOCIAL_TYPE.VENMO,
		username: '',
		url: 'https://venmo.com/${username}'
	},
	CashApp: {
		platform: SOCIAL_TYPE.CASH_APP,
		username: '',
		url: 'https://cash.app/$${username}'
	},
	Deezer: {
		platform: SOCIAL_TYPE.DEEZER,
		username: '',
		url: 'https://deezer.com/${username}'
	},
	Pandora: {
		platform: SOCIAL_TYPE.PANDORA,
		username: '',
		url: ''
	},
	Tidal: {
		platform: SOCIAL_TYPE.TIDAL,
		username: '',
		url: ''
	},
	Bandcamp: {
		platform: SOCIAL_TYPE.BANDCAMP,
		username: '',
		url: 'https://${username}.bandcamp.com'
	},
	Medium: {
		platform: SOCIAL_TYPE.MEDIUM,
		username: '',
		url: 'https://medium.com/@${username}'
	},
	Beehiiv: {
		platform: SOCIAL_TYPE.BEEHIIV,
		username: '',
		url: ''
	},
	Flipboard: {
		platform: SOCIAL_TYPE.FLIPBOARD,
		username: '',
		url: 'https://flipboard.com/@${username}'
	},
	Goodreads: {
		platform: SOCIAL_TYPE.GOODREADS,
		username: '',
		url: ''
	},
	Mastodon: {
		platform: SOCIAL_TYPE.MASTODON,
		username: '',
		url: 'https://mastodon.social/@${username}'
	},
	Steam: {
		platform: SOCIAL_TYPE.STEAM,
		username: '',
		url: 'https://steamcommunity.com/id/${username}'
	},
	Tumblr: {
		platform: SOCIAL_TYPE.TUMBLR,
		username: '',
		url: 'https://tumblr.com/${username}'
	}
	// youtube: {
	// 	platform: SOCIAL_TYPE.YOUTUBE,
	// 	username: '',
	// 	url: 'https://www.youtube.com/@${username}'
	// },
	// linkedin: {
	// 	platform: SOCIAL_TYPE.LINKEDIN,
	// 	username: '',
	// 	url: 'https://www.linkedin.com/in/${username}'
	// },
	// pinterest: {
	// 	platform: SOCIAL_TYPE.PINTEREST,
	// 	username: '',
	// 	url: 'https://www.pinterest.com/${username}'
	// },
	// snapchat: {
	// 	platform: SOCIAL_TYPE.SNAPCHAT,
	// 	username: '',
	// 	url: 'https://www.snapchat.com/add/${username}'
	// },
	// spotify: {
	// 	platform: SOCIAL_TYPE.SPOTIFY,
	// 	username: '',
	// 	url: 'https://open.spotify.com/user/${username}'
	// },
	// threads: {
	// 	platform: SOCIAL_TYPE.THREADS,
	// 	username: '',
	// 	url: 'https://threads.net/@${username}'
	// },
	// reddit: {
	// 	platform: SOCIAL_TYPE.REDDIT,
	// 	username: '',
	// 	url: 'https://www.reddit.com/user/${username}'
	// },
	// twitch: {
	// 	platform: SOCIAL_TYPE.TWITCH,
	// 	username: '',
	// 	url: 'https://twitch.tv/${username}'
	// },
	// dribbble: {
	// 	platform: SOCIAL_TYPE.DRIBBBLE,
	// 	username: '',
	// 	url: 'https://dribbble.com/${username}'
	// },
	// discord: {
	// 	platform: SOCIAL_TYPE.DISCORD,
	// 	username: '',
	// 	url: 'https://discord.gg/${username}'
	// },
	// substack: {
	// 	platform: SOCIAL_TYPE.SUBSTACK,
	// 	username: '',
	// 	url: 'https://substack.com/@${username}'
	// },
	// telegram: {
	// 	platform: SOCIAL_TYPE.TELEGRAM,
	// 	username: '',
	// 	url: 'https://t.me/${username}'
	// }
};
