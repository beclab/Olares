<template>
	<q-layout view="lhr lpr lfr" :container="platform == 'WEB' ? false : true">
		<VaultsDrawer />

		<q-page-container>
			<q-page
				class="container"
				:class="platform == 'DESKTOP' ? 'desktop-vault' : ''"
			>
				<router-view />
			</q-page>
		</q-page-container>
	</q-layout>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, onUnmounted } from 'vue';
import { date } from 'quasar';
import { app } from '../globals';
import { useMenuStore } from '../stores/menu';
import { getAppPlatform } from '../application/platform';

import VaultsDrawer from './TermipassLayout/VaultsDrawer.vue';

export default defineComponent({
	name: 'MainLayout',

	components: {
		VaultsDrawer
	},

	setup() {
		const platform = ref(process.env.PLATFORM);
		const menuStore = useMenuStore();

		onMounted(async () => {
			getAppPlatform().homeMounted();

			menuStore.syncInfo = {
				syncing: app.state.syncing || false,
				lastSyncTime: date.formatDate(app.state.lastSync, 'HH:mm:ss')
			};
			menuStore.handleSync();
		});

		onUnmounted(() => {
			getAppPlatform().homeUnMounted();
		});

		return {
			app,
			platform
		};
	}
});
</script>

<style lang="scss" scoped>
.container {
	height: 100%;
	&.desktop-vault {
		width: calc(100% - 4px);
		margin-left: 4px;
	}
}
</style>
