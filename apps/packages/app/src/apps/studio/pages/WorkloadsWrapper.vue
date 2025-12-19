<template>
	<Workloads class="studio-workload-wrapper" :refreshList="status"></Workloads>
</template>

<script setup lang="ts">
import { useDockerStore } from '@apps/studio/src/stores/docker';
import { APP_INSTALL_STATE } from '@apps/studio/src/types/core';
import Workloads from '@apps/control-hub/src/pages/ApplicationSpaces/Workloads/Workloads.vue';
import { computed, ref, watch, onUnmounted } from 'vue';
import debounce from 'lodash.debounce';

const dockerStore = useDockerStore();

const refreshTrigger = ref(0);
const status = computed(
	() =>
		dockerStore.appInstallState === APP_INSTALL_STATE.COMPLETED ||
		refreshTrigger.value > 0
);

let lastRefreshTransition: string | null = null;

const FINAL_STATES = [
	APP_INSTALL_STATE.COMPLETED,
	APP_INSTALL_STATE.RUNNING,
	APP_INSTALL_STATE.RESUMED,
	APP_INSTALL_STATE.FAILED,
	APP_INSTALL_STATE.CANCELED
];

const isFinalState = (state: APP_INSTALL_STATE) => {
	return FINAL_STATES.includes(state);
};

const triggerRefresh = (transition: string) => {
	if (lastRefreshTransition === transition) {
		return;
	}
	refreshTrigger.value++;
	lastRefreshTransition = transition;
};

const debouncedRefreshFast = debounce(triggerRefresh, 300);
const debouncedRefreshSlow = debounce(triggerRefresh, 500);

watch(
	() => dockerStore.appStatusInfo?.state,
	(newState, oldState) => {
		if (!newState || newState === oldState) {
			return;
		}

		const transition = `${oldState}_to_${newState}`;

		if (isFinalState(newState)) {
			const isFast =
				newState === APP_INSTALL_STATE.COMPLETED ||
				newState === APP_INSTALL_STATE.RUNNING ||
				newState === APP_INSTALL_STATE.RESUMED;
			isFast
				? debouncedRefreshFast(transition)
				: debouncedRefreshSlow(transition);
			return;
		}

		if (
			oldState === APP_INSTALL_STATE.DOWNLOADING &&
			(newState === APP_INSTALL_STATE.INSTALLING ||
				newState === APP_INSTALL_STATE.PROCESSING)
		) {
			debouncedRefreshSlow(transition);
			return;
		}

		if (
			oldState === APP_INSTALL_STATE.PROCESSING &&
			newState === APP_INSTALL_STATE.INSTALLING
		) {
			debouncedRefreshSlow(transition);
			return;
		}

		if (
			oldState &&
			oldState !== APP_INSTALL_STATE.UNINSTALLING &&
			newState === APP_INSTALL_STATE.UNINSTALLING
		) {
			debouncedRefreshFast(transition);
			return;
		}

		if (
			(oldState === APP_INSTALL_STATE.RUNNING ||
				oldState === APP_INSTALL_STATE.COMPLETED) &&
			newState === APP_INSTALL_STATE.UPGRADING
		) {
			debouncedRefreshFast(transition);
			return;
		}
	}
);

onUnmounted(() => {
	debouncedRefreshFast.cancel();
	debouncedRefreshSlow.cancel();
});
</script>

<style></style>
