<template>
	<div v-if="value" class="support-client row justify-between items-center">
		<div class="row justify-start items-center">
			<q-img class="client-icon" :src="imageRef" />
			<div class="text-body2 q-ml-sm text-ink-2">{{ label }}</div>
		</div>
		<div
			class="text-body2 q-ml-sm text-info info_link cursor-pointer"
			@click="onClickIcon"
		>
			{{ store }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { onMounted, PropType, ref } from 'vue';
import { getRequireImage } from '@apps/market/src/utils/imageUtils';
import { CLIENT_TYPE } from '@apps/market/src/constant/constants';
import { i18n } from '@apps/market/src/boot/i18n';

const props = defineProps({
	type: {
		type: Object as PropType<CLIENT_TYPE>,
		require: true
	},
	value: {
		type: String,
		default: '',
		require: true
	}
});

const imageRef = ref();
const label = ref();
const store = ref();

onMounted(() => {
	switch (props.type) {
		case CLIENT_TYPE.android:
			label.value = 'Android';
			store.value = 'Google Play';
			imageRef.value = getRequireImage('client/android.svg');
			break;
		case CLIENT_TYPE.ios:
			label.value = 'iOS';
			store.value = 'App Store';
			imageRef.value = getRequireImage('client/ios.svg');
			break;
		case CLIENT_TYPE.mac:
			label.value = 'Mac';
			store.value = 'Mac App Store';
			imageRef.value = getRequireImage('client/macos.svg');
			break;
		case CLIENT_TYPE.windows:
			label.value = 'Windows';
			store.value = i18n.global.t('detail.download');
			imageRef.value = getRequireImage('client/windows.svg');
			break;
		case CLIENT_TYPE.linux:
			label.value = 'Linux';
			store.value = i18n.global.t('detail.download');
			imageRef.value = getRequireImage('client/linux.svg');
			break;
		case CLIENT_TYPE.edge:
			label.value = 'Edge';
			store.value = 'Edge Addons';
			imageRef.value = getRequireImage('client/edge.svg');
			break;
		case CLIENT_TYPE.chrome:
			label.value = 'Chrome';
			store.value = 'Chrome Web Store';
			imageRef.value = getRequireImage('client/chrome.svg');
			break;
	}
});

const onClickIcon = () => {
	if (props.value) {
		window.open(props.value);
	}
};
</script>

<style scoped lang="scss">
.support-client {
	height: 32px;
	width: 100%;

	.client-icon {
		width: 32px;
		height: 32px;
	}

	.info_link {
		text-decoration: none;
	}
}
</style>
