<template>
	<app-store-body
		v-if="installApp.length > 0"
		:show-body="true"
		:no-label-padding-bottom="true"
	>
		<template v-slot:body>
			<div
				:class="
					deviceStore.isMobile
						? 'app-store-workflow-mobile'
						: 'app-store-workflow'
				"
			>
				<function-app-card
					v-for="app in installApp"
					:key="app"
					:app-name="app"
					:source-id="sourceId"
					:manager="true"
					:version-display-mode="VERSION_DISPLAY_MODE.ONLY_MY"
				/>
			</div>
		</template>
	</app-store-body>
	<empty-view v-else :label="t('my.no_installed_app_tips')" class="full-view" />
</template>

<script setup lang="ts">
import FunctionAppCard from '../../../components/appcard/FunctionAppCard.vue';
import AppStoreBody from '../../../components/base/AppStoreBody.vue';
import EmptyView from '../../../components/base/EmptyView.vue';
import { useDeviceStore } from '../../../stores/settings/device';
import { VERSION_DISPLAY_MODE } from '../../../constant/constants';
import { useCenterStore } from '../../../stores/market/center';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';

const props = defineProps({
	sourceId: {
		type: String,
		require: true
	}
});

const { t } = useI18n();
const centerStore = useCenterStore();
const deviceStore = useDeviceStore();
const installApp = computed(() => {
	return centerStore.getSourceInstalledApp(props.sourceId);
});
</script>

<style lang="scss">
::-webkit-scrollbar {
	display: none;
}

.full-view {
	width: 100%;
	height: 100%;
}
</style>
