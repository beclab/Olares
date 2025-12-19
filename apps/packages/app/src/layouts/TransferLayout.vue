<template>
	<q-layout view="lHh LpR lFr" :container="true">
		<TransferDrawer />
		<q-page-container>
			<q-page
				class="transfer-table"
				:class="
					$q.platform.is.win && $q.platform.is.electron
						? 'transfer-table-win'
						: isPad
						? 'transfer-table-pad'
						: 'transfer-table-common'
				"
			>
				<router-view />
			</q-page>
		</q-page-container>
	</q-layout>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import TransferDrawer from './TermipassLayout/TransferDrawer.vue';
import { getAppPlatform } from '../application/platform';

const isPad = ref(getAppPlatform() && getAppPlatform().isPad);
</script>

<style lang="scss" scoped>
.transfer-table-common {
	height: calc(100vh - 143px) !important;
}

.transfer-table-win {
	height: calc(100vh - 165px) !important;
}

.transfer-table-pad {
	height: calc(100vh - 165px) !important;
	background: $white;
}
</style>
