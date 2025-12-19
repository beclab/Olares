<template>
	<div
		class="terminus-qr-code-root"
		:style="{ '--size': `${size}px`, '--iconSize': `${size * 0.267}px` }"
	>
		<qrcode-vue :value="url" :size="size" />
		<div
			v-if="status === QR_STATUS.EXPIRED"
			class="refresh-background column justify-center items-center"
			@click="onClickRefresh"
		>
			<q-img
				:class="{ rotating: isLoading }"
				class="refresh-icon cursor-pointer"
				:src="getRequireImage('refresh_qr_code.svg')"
			/>
		</div>
		<div
			v-if="status === QR_STATUS.SUCCESSFUL"
			class="refresh-background column justify-center items-center"
		>
			<q-img
				:class="{ rotating: isLoading }"
				class="refresh-icon cursor-pointer"
				:src="getRequireImage('qr_refresh_success.svg')"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import QrcodeVue from 'qrcode.vue';
import { useI18n } from 'vue-i18n';
import { PropType, ref, watch } from 'vue';
import { QR_STATUS } from 'src/constant/index';
import { getRequireImage } from '../../utils/settings/helper';

const { t } = useI18n();

const props = defineProps({
	url: {
		type: String,
		required: true
	},
	size: {
		type: Number,
		required: true
	},
	status: {
		type: Object as PropType<QR_STATUS>,
		required: true
	},
	textStyle: {
		type: String,
		default: 'text-subtitle1'
	}
});

const isLoading = ref(false);

watch(
	() => props.status,
	(newValue) => {
		if (newValue == QR_STATUS.NORMAL) {
			isLoading.value = false;
		}
	}
);

const emit = defineEmits(['onRefresh']);

const onClickRefresh = () => {
	isLoading.value = true;
	emit('onRefresh');
};
</script>

<style lang="scss" scoped>
.terminus-qr-code-root {
	position: relative;
	width: var(--size);
	height: var(--size);

	.refresh-background {
		width: var(--size);
		height: var(--size);
		top: 0;
		left: 0;
		position: absolute;
		background: #ffffffe5;

		.refresh-icon {
			width: var(--iconSize);
			height: var(--iconSize);
		}

		.refresh-text {
			white-space: pre-wrap;
			text-align: center;
		}

		.rotating {
			animation: spin 1s linear infinite;
		}

		@keyframes spin {
			from {
				transform: rotate(0deg);
			}
			to {
				transform: rotate(360deg);
			}
		}
	}
}
</style>
