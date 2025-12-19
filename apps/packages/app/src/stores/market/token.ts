import { defineStore } from 'pinia';
import { Token } from 'src/constant/constants';
//import axios from 'axios';

export type RootState = {
	token: Token;
	url: string | null;
};

export const useTokenStore = defineStore('token', {
	state: () => {
		return {
			// token: {},
			// url: ''
		} as RootState;
	},
	actions: {
		//
	}
});
