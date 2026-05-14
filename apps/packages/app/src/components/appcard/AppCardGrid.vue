<template>
	<div v-if="appList" :class="deviceStore.isMobile ? rule + '-mobile' : rule">
		<template v-for="item in showAppList" :key="item">
			<slot name="card" :app="item" />
		</template>
		<app-card-hide-border />
	</div>
	<div v-else :class="deviceStore.isMobile ? rule + '-mobile' : rule">
		<template v-for="index in showAppSize" :key="index">
			<slot name="card" />
		</template>
		<app-card-hide-border />
	</div>
</template>

<script lang="ts" setup>
import AppCardHideBorder from '../../components/appcard/AppCardHideBorder.vue';
import { useDeviceStore } from '../../stores/settings/device';
import { computed, PropType } from 'vue';
import { getSliceArray } from '../../utils/utils';
import { useQuasar } from 'quasar';

const props = defineProps({
	rule: {
		type: String,
		require: true
	},
	appList: {
		type: Array as PropType<string[]>,
		default: undefined
	},
	showSize: {
		type: String,
		default: '15,9,6'
	}
});

const $q = useQuasar();
const deviceStore = useDeviceStore();
const sizeArray = computed(() => {
	if (deviceStore.isMobile) {
		return [3];
	}
	return props.showSize
		.split(',')
		.map((item) => Number(item.trim()))
		.filter((item) => !Number.isNaN(item));
});

const showAppSize = computed(() => {
	const [lg = 15, md = lg, sm = md] = sizeArray.value;
	if (deviceStore.isMobile) {
		return lg;
	}
	if ($q.screen.lg || $q.screen.xl) {
		return lg;
	}
	if ($q.screen.md) {
		return md;
	}
	return sm;
});

const showAppList = computed(() => {
	if (!props.appList) {
		return [];
	}
	return getSliceArray(props.appList, showAppSize.value);
});
</script>

<style scoped lang="scss"></style>
