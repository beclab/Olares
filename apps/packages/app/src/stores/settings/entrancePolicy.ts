import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import _ from 'lodash';
import { useApplicationStore } from 'src/stores/settings/application';
import {
	notifyFailed,
	notifySuccess,
	notifyWarning
} from 'src/utils/settings/btNotify';
import { AUTH_LEVEL, authLevelOptions, EntrancePolicy } from 'src/constant';

export const useEntrancePolicyStore = defineStore('entrancePolicy', () => {
	const applicationStore = useApplicationStore();

	const applicationName = ref<string>('');
	const entranceName = ref<string>('');

	const authorizationLevel = ref<string>();
	const factorMode = ref<string>();
	const oneTimeMode = ref(true);
	const validDuration = ref(0);
	const sub_policies = ref<EntrancePolicy[]>([]);

	const oldAuthorizationLevel = ref<string>();
	const oldFactorMode = ref<string>();
	const oldOneTimeMode = ref(false);
	const oldValidDuration = ref(0);
	const oldSubPolicies = ref<EntrancePolicy[]>([]);

	const isLoading = ref(true);

	const policiesCount = computed(() => sub_policies.value.length);

	const hasAuthLevelChanges = computed(() => {
		return oldAuthorizationLevel.value !== authorizationLevel.value;
	});

	const hasFactorModelChanges = computed(() => {
		return (
			oldOneTimeMode.value !== oneTimeMode.value ||
			oldFactorMode.value !== factorMode.value ||
			oldValidDuration.value !== validDuration.value
		);
	});

	const hasPoliciesChanges = computed(() => {
		return (
			JSON.stringify(oldSubPolicies.value) !==
			JSON.stringify(sub_policies.value)
		);
	});

	const hasAnyChanges = computed(() => {
		return (
			hasAuthLevelChanges.value ||
			hasFactorModelChanges.value ||
			hasPoliciesChanges.value
		);
	});

	const resultCode = computed(() => {
		const condition1 = !hasAuthLevelChanges.value;
		const condition2 =
			!hasFactorModelChanges.value && !hasPoliciesChanges.value;
		return (condition1 ? 2 : 0) | (condition2 ? 1 : 0);
	});

	async function init(appName: string, entName: string) {
		applicationName.value = appName;
		entranceName.value = entName;
		isLoading.value = true;

		if (!(appName in applicationStore.entrances)) {
			await applicationStore.getEntrances(appName);
		}

		await fetchPolicy();
		await fetchAuthLevel();
		isLoading.value = false;
	}

	async function fetchPolicy() {
		const application = applicationStore.getApplicationById(
			applicationName.value
		);
		const res = await applicationStore.getPolicy(
			application?.name,
			entranceName.value
		);

		factorMode.value = res.default_policy;
		oneTimeMode.value = res.one_time;
		validDuration.value = res.valid_duration;
		sub_policies.value = res.sub_policies || [];

		oldFactorMode.value = res.default_policy;
		oldOneTimeMode.value = res.one_time;
		oldValidDuration.value = res.valid_duration;
		oldSubPolicies.value = _.cloneDeep(res.sub_policies || []);
	}

	async function fetchAuthLevel() {
		const res =
			applicationStore.entrances[applicationName.value][entranceName.value];
		authorizationLevel.value = res.authLevel || AUTH_LEVEL.Public;
		oldAuthorizationLevel.value = res.authLevel || AUTH_LEVEL.Public;
	}

	async function submitFactorModel(
		t: (key: string, params?: any) => string,
		isSilent = false
	) {
		const params = {
			default_policy: factorMode.value,
			one_time: oneTimeMode.value,
			valid_duration: validDuration.value,
			sub_policies: sub_policies.value.length <= 0 ? null : sub_policies.value
		};

		const findEmptyIndex = sub_policies.value.findIndex((item) => !item.uri);
		if (findEmptyIndex > -1) {
			notifyWarning(
				t('the_item_index_is_empty', {
					index: findEmptyIndex + 1
				})
			);
			throw new Error('Empty URI found');
		}

		const application = applicationStore.getApplicationById(
			applicationName.value
		);
		await applicationStore.set_appFa2(
			params,
			application?.name,
			entranceName.value
		);

		if (!isSilent) notifySuccess(t('success'));
	}

	async function submitAuthLevel(
		t: (key: string, params?: any) => string,
		isSilent = false
	) {
		if (
			!authorizationLevel.value ||
			authLevelOptions().find((e) => e.value === authorizationLevel.value) ===
				undefined
		) {
			notifyWarning(
				t('auth_level_is_error_error', {
					error: authorizationLevel.value
				})
			);
			throw new Error('Invalid auth level');
		}

		const application = applicationStore.getApplicationById(
			applicationName.value
		);
		await applicationStore.setupAuthLevel(
			application?.name,
			entranceName.value,
			{
				authorization_level: authorizationLevel.value
			}
		);

		if (!isSilent) notifySuccess(t('success'));
	}

	async function submitAll(t: (key: string, params?: any) => string) {
		isLoading.value = true;
		const tasks: Promise<any>[] = [];

		try {
			switch (resultCode.value) {
				case 0:
					tasks.push(submitAuthLevel(t, true));
					tasks.push(submitFactorModel(t, true));
					break;
				case 1:
					tasks.push(submitAuthLevel(t, true));
					break;
				case 2:
					tasks.push(submitFactorModel(t, true));
					break;
			}

			if (tasks.length > 0) {
				await Promise.all(tasks);
				notifySuccess(t('success'));
			}
		} catch (e: any) {
			notifyFailed(e.message);
		} finally {
			await fetchAuthLevel();
			await fetchPolicy();
			isLoading.value = false;
		}
	}

	async function submitPolicies(t: (key: string, params?: any) => string) {
		isLoading.value = true;

		try {
			await submitFactorModel(t);
			await fetchPolicy();
		} catch (e: any) {
			if (e.message !== 'Empty URI found') {
				notifyFailed(e.message);
			}
		} finally {
			isLoading.value = false;
		}
	}

	function $reset() {
		applicationName.value = '';
		entranceName.value = '';
		authorizationLevel.value = undefined;
		factorMode.value = undefined;
		oneTimeMode.value = true;
		validDuration.value = 0;
		sub_policies.value = [];
		oldAuthorizationLevel.value = undefined;
		oldFactorMode.value = undefined;
		oldOneTimeMode.value = false;
		oldValidDuration.value = 0;
		oldSubPolicies.value = [];
		isLoading.value = true;
	}

	return {
		applicationName,
		entranceName,
		authorizationLevel,
		factorMode,
		oneTimeMode,
		validDuration,
		sub_policies,
		isLoading,

		policiesCount,
		hasAuthLevelChanges,
		hasFactorModelChanges,
		hasPoliciesChanges,
		hasAnyChanges,
		resultCode,

		init,
		fetchPolicy,
		fetchAuthLevel,
		submitAll,
		submitPolicies,
		$reset
	};
});
