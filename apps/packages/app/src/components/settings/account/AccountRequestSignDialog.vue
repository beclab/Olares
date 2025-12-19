<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:skip="false"
		:ok="t('confirm')"
		:cancel="false"
		size="medium"
		:platform="deviceStore.platform"
		@onSubmit="onOKClick"
	>
		<div class="policy-content">
			<div class="text-body3 policy-content__info">
				{{ info }}
			</div>
			<q-img
				class="policy-content__image"
				:src="imagePath"
				v-if="imagePath && imagePath.length"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from 'src/stores/settings/device';

defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	info: {
		type: String,
		required: false,
		default: ''
	},
	imagePath: {
		type: String,
		required: false,
		default: ''
	}
});

const { t } = useI18n();
const CustomRef = ref();
const deviceStore = useDeviceStore();
async function onOKClick() {
	CustomRef.value.onDialogOK();
}
</script>
<style lang="scss" scoped>
.policy-content {
	padding: 0px;

	&__info {
		color: $ink-2;
	}

	&__image {
		margin-top: 20px;
	}

	.confirm-r-btn {
		height: 32px;
		width: 83px;
	}
}
</style>
