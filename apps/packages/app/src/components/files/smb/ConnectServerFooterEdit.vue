<template>
	<div class="row items-center q-mt-md q-mb-sm q-px-none">
		<div class="operate row item-center justify-center">
			<div class="operate-item q-pr-xs">
				<q-icon
					name="sym_r_add"
					size="24px"
					@click="handleAdd"
					:style="{ color: showAdd ? '#5c5c5c' : '#adadad' }"
				/>
			</div>
			<div class="operate-item q-pl-xs">
				<q-icon
					name="sym_r_remove"
					size="24px"
					@click="handleRemove"
					:style="{ color: showRemove ? '#5c5c5c' : '#adadad' }"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { i18n } from '../../../boot/i18n';

defineProps({
	okText: {
		type: String,
		default: i18n.global.t('confirm'),
		required: false
	},
	cancelText: {
		type: String,
		default: i18n.global.t('cancel'),
		required: false
	},
	showCancel: {
		type: Boolean,
		default: true,
		required: false
	},
	loading: {
		type: Boolean,
		default: false,
		required: false
	},

	showAdd: {
		type: Boolean,
		default: false,
		required: false
	},
	showRemove: {
		type: Boolean,
		default: false,
		required: false
	}
});

const emits = defineEmits(['handleAdd', 'handleRemove']);

const handleAdd = () => {
	emits('handleAdd');
};

const handleRemove = () => {
	emits('handleRemove');
};
</script>

<style scoped lang="scss">
.operate {
	border: 1px solid $separator;
	border-radius: 4px;
	padding: 8px;
	box-sizing: border-box;
	.operate-item {
		width: 40px;
		height: 24px;
		line-height: 24px;
		text-align: center;
		cursor: pointer;
	}
	.operate-item:first-child {
		border-right: 1px solid $separator;
		box-sizing: border-box;
	}
}
.but-creat {
	width: 100px;
	height: 40px;
	line-height: 40px;
	border-radius: 8px;
	background: $yellow;
}
.but-cancel {
	width: 100px;
	height: 40px;
	line-height: 40px;
	border-radius: 8px;
	border: 1px solid $separator;
}
</style>
