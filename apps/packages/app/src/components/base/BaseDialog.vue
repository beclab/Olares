<template>
	<q-card class="card-container-pc column no-shadow">
		<terminus-dialog-bar
			:label="label"
			@close="emit('close')"
			tit-align="text-subtitle3"
		/>

		<div :class="padding ? padding : 'q-pa-lg'">
			<slot />
		</div>

		<terminus-dialog-footer
			@close="emit('close')"
			@submit="emit('submit')"
			@more="emit('more')"
			:ok-text="okText"
			:ok-disable="okDisable"
			:cancel-text="cancelText"
			:more-text="moreText"
			:show-more="showMore"
			:show-cancel="showCancel"
			:loading="loading"
		/>
	</q-card>
</template>

<script lang="ts" setup>
import TerminusDialogBar from './TerminusDialogBar';
import TerminusDialogFooter from './TerminusDialogFooter.vue';
import { i18n } from '../../boot/i18n';

defineProps({
	label: {
		type: String,
		default: '',
		required: false
	},
	padding: {
		type: String,
		default: ''
	},
	okText: {
		type: String,
		default: i18n.global.t('base.confirm'),
		required: false
	},
	okDisable: {
		type: Boolean,
		default: false
	},
	cancelText: {
		type: String,
		default: i18n.global.t('base.cancel'),
		required: false
	},
	moreText: {
		type: String,
		default: i18n.global.t('base.more'),
		required: false
	},
	showCancel: {
		type: Boolean,
		default: true,
		required: false
	},
	showMore: {
		type: Boolean,
		default: false,
		required: false
	},
	loading: {
		type: Boolean,
		default: false,
		required: false
	}
});
const emit = defineEmits(['submit', 'close', 'more']);
</script>

<style scoped lang="scss">
.card-dialog-pc {
	.card-container-pc {
		border-radius: 12px;
		width: 400px;
	}
}
</style>
