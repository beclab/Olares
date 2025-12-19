<template>
	<bt-custom-dialog
		ref="customRef"
		:size="showCheckbox ? 'medium' : 'small'"
		:title="label"
		@onSubmit="onOK"
		:okLoading="loading"
		:cancel="cancelText"
		:ok="okText"
		:okDisabled="okDisable"
	>
		<div class="column">
			<div class="text-ink-2 text-body3">
				{{ content }}
			</div>

			<bt-check-box
				v-if="showCheckbox"
				:label="boxLabel"
				:model-value="selected"
				@update:model-value="onUpdate"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { i18n } from '../../boot/i18n';
import BtCheckBox from '../rss/BtCheckBox.vue';
import { ref } from 'vue';

const props = defineProps({
	label: {
		type: String,
		default: '',
		required: false
	},
	content: {
		type: String,
		default: '',
		required: false
	},
	modelValue: {
		type: Boolean,
		default: true,
		required: true
	},
	boxLabel: {
		type: String,
		default: '',
		required: false
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
	},
	showCheckbox: {
		type: Boolean,
		default: true,
		required: false
	}
});

const selected = ref(props.modelValue);
const customRef = ref();

const onUpdate = (status) => {
	selected.value = status;
};

const onOK = () => {
	customRef.value.onDialogOK(selected.value);
};
</script>

<style scoped lang="scss"></style>
