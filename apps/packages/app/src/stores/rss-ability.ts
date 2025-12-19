import { defineStore } from 'pinia';
import { getAbiAbility } from 'src/api/wise';

export const useAbilityStore = defineStore('ability', {
	state: () => ({
		rssubscribe: false,
		twitter: false,
		ytdlp: false
	}),

	actions: {
		async getAbiAbility() {
			const { rssubscribe, twitter, ytdlp }: any = await getAbiAbility();
			this.rssubscribe = rssubscribe;
			this.twitter = twitter;
			this.ytdlp = ytdlp;
		}
	}
});
