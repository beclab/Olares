import { Router } from 'vue-router';
import { useBexStore } from '../../../stores/bex';

export const useApproval = (router: Router) => {
	const resolveApproval = async (data?: any, forceReject = false) => {
		const store = useBexStore();
		const approval = await store.controller.getApproval();

		if (approval) {
			await store.controller.resolveApproval(data, forceReject);
		}
		setTimeout(() => {
			router.replace('/home');
		});
	};

	const rejectApproval = async (err?: any, isInternal = false) => {
		const store = useBexStore();
		const approval = await store.controller.getApproval();
		if (approval) {
			await store.controller.rejectApproval(err, isInternal);
		}
		setTimeout(() => {
			router.replace('/home');
		});
	};

	return { resolveApproval, rejectApproval };
};
