<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files.connect_to_server')"
		:skip="false"
		:okLoading="loading ? t('loading') : false"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="medium"
		:persistent="true"
		@onCancel="onCancel"
	>
		<div class="dialog-desc row items-center justify-between">
			<div class="connect-icon row items-center justify-center q-mr-md">
				<q-icon name="sym_r_language" size="24px" color="white" />
			</div>
			<div class="connect-content">
				<div class="text-ink-1 text-body2">Connecting</div>
				<div class="text-ink-1 text-body2">{{ smb_url }}</div>
			</div>
		</div>

		<div class="connecting">
			<div class="move"></div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

import { useDataStore } from '../../../stores/data';

defineProps({
	origin_id: {
		type: Number,
		required: true
	},
	smb_url: {
		type: String,
		required: true
	}
});

const store = useDataStore();
const { t } = useI18n();

const loading = ref(false);

const CustomRef = ref();

const onCancel = () => {
	store.closeHovers();
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		width: 560px;
		border-radius: 12px;
		padding-bottom: 20px;

		.dialog-desc {
			width: 100%;
			padding: 0 20px;
			.connect-icon {
				width: 40px;
				height: 40px;
				border-radius: 10px;
				background-color: $light-blue-default;
			}
			.connect-content {
				flex: 1;
			}
		}

		.connecting {
			width: calc(100% - 40px);
			margin: 20px 20px 0 20px;
			height: 16px;
			border: 1px solid $input-stroke;
			border-radius: 9px;
			position: relative;
			overflow: hidden;
			.move {
				width: 48px;
				height: 8px;
				background-color: $light-blue-default;
				border-radius: 4px;
				position: absolute;
				top: 3px;
				left: 5px;
				animation: moveAni 2s infinite alternate ease-in-out;
			}
		}
	}
}

@keyframes moveAni {
	from {
		transform: translateX(-30px);
	}
	to {
		transform: translateX(500px);
	}
}
</style>
