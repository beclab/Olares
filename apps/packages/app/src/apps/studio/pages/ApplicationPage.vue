<template>
	<div
		class="container"
		v-if="
			dockerStore.appStatus &&
			[
				APP_STATUS.DEPLOYED,
				APP_STATUS.DEPLOYING,
				APP_STATUS.UNDEPLOY,
				APP_STATUS.ABNORMAL
			].includes(dockerStore.appStatus)
		"
	>
		<EditComponent :app="store.current_app" />
	</div>

	<default-app-page v-else />
</template>

<script lang="ts" setup>
import { useDevelopingApps } from '../stores/app';
import { useDockerStore } from '../stores/docker';
import { APP_STATUS } from '@apps/studio/src/types/core';

import EditComponent from '../components/EditComponent.vue';
import DefaultAppPage from './DefaultAppPage.vue';

const store = useDevelopingApps();
const dockerStore = useDockerStore();
</script>
<style scoped lang="scss">
.container {
	height: 100%;
}
</style>
