import { defineStore } from 'pinia';
import { OlaresInfo } from '@bytetrade/core';
import { TERMINUS_ID } from 'src/utils/localStorageConstant';
import axios from 'axios';

export type TerminusState = {
	terminusInfo: OlaresInfo | null;
};

export const useTerminusStore = defineStore('terminus', {
	state: () => {
		return {
			terminusInfo: null
		} as TerminusState;
	},
	getters: {
		olaresId(): string {
			return (
				this.terminusInfo?.olaresId || this.terminusInfo?.terminusName || ''
			);
		},
		olares_device_id(): string {
			return this.terminusInfo?.id || this.terminusInfo?.terminusId || '';
		}
	},

	actions: {
		async getTerminusInfo() {
			const parts = window.location.hostname.split('.');
			let url = '';
			if (parts.length > 1) {
				const processedParts = parts.slice(1);
				const processedHostname = processedParts.join('.');
				url = window.location.protocol + '//' + processedHostname;
			} else {
				url = window.location.protocol + '//' + window.location.hostname;
			}
			this.terminusInfo = await axios.get(url + '/api/olares-info', {});
		},
		async validateTerminusInfo(
			customValidator: (currentId: string, lastId: string | null) => boolean = (
				currentId,
				lastId
			) => currentId === (lastId ?? ''),
			onSuccess = () => {},
			onFailure = () => {},
			onFinally = () => {}
		) {
			if (!this.terminusInfo || !this.terminusInfo.id) {
				console.log('===> Validation check new machine');
				await onFailure();
				await onFinally();
				return;
			}

			const currentId = this.terminusInfo.id;
			const lastId = localStorage.getItem(TERMINUS_ID) ?? '';

			const isValid = customValidator(currentId, lastId);

			if (isValid) {
				console.log('===> Validation succeeded');
				await onSuccess();
			} else {
				localStorage.setItem(TERMINUS_ID, currentId);
				console.log('===> Validation failed');
				await onFailure();
			}
			await onFinally();
		}
	}
});
