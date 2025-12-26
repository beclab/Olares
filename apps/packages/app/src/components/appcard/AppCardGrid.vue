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
import { onBeforeUnmount, onMounted, PropType, ref } from 'vue';
import { getSliceArray } from '../../utils/utils';
import { useQuasar } from 'quasar';

const props = defineProps({
	rule: {
		type: String,
		require: true
	},
	appList: {
		type: Object as PropType<string[]>,
		default: [] as string[]
	},
	showSize: {
		type: String,
		default: '15,9,6'
	}
});

const $q = useQuasar();
const allAppList = ref();
const showAppList = ref();
const showAppSize = ref();
const deviceStore = useDeviceStore();
const sizeArray = deviceStore.isMobile ? ['3'] : props.showSize.split(',');

onMounted(async () => {
	if (props.appList) {
		allAppList.value = props.appList;
		updateAppList();
	}
	window.addEventListener('resize', () => {
		updateAppList();
	});
});

onBeforeUnmount(() => {
	window.removeEventListener('resize', () => {
		updateAppList();
	});
});

const updateAppList = () => {
	showAppSize.value = deviceStore.isMobile
		? Number(sizeArray[0])
		: $q.screen.lg || $q.screen.xl
		? Number(sizeArray[0])
		: $q.screen.md
		? Number(sizeArray[1])
		: Number(sizeArray[2]);
	if (props.appList) {
		showAppList.value = getSliceArray(allAppList.value, showAppSize.value);
	}
};
</script>

<style scoped lang="scss"></style>
