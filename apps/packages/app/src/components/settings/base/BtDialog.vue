<template>
	<q-card class="bt-dialog">
		<DialogHeader :title="title" @close-action="emit('onCancel')" />
		<div class="common-dialog bt-dialog-content" style="border-radius: 16px">
			<div class="column justify-start">
				<slot />
				<dialog-footer
					:confirm-text="t('confirm')"
					@cancel-action="emit('onCancel')"
					@confirm-action="emit('onConfirm')"
				/>
			</div>
		</div>
	</q-card>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import DialogHeader from '../DialogHeader.vue';
import DialogFooter from '../DialogFooter.vue';

defineProps({
	title: {
		type: String,
		require: true
	},
	confirmClickAble: {
		type: Boolean,
		default: true
	}
});

const { t } = useI18n();

const emit = defineEmits(['onCancel', 'onConfirm']);
</script>
<style lang="scss">
.bt-dialog {
	max-width: 400px;
	width: 400px;
	height: auto;
	padding: 0;
	position: relative;

	.bt-dialog-header {
		width: 100%;
		height: 32px;
		background: $background-3;
		padding-left: 12px;
		padding-right: 12px;
	}

	.bt-dialog-content {
		width: 100%;
		height: calc(100% - 32px);
		padding: 20px;
	}
}

.q-dialog__inner--minimized > div {
	max-width: 400px;
}

.q-dialog__inner > div {
	border-radius: 12px;
}
</style>
