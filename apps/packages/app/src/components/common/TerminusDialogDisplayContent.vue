<template>
	<q-card class="common-card">
		<div class="column content-bg">
			<div class="text-h5 row justify-between items-center title-part q-px-md">
				<div v-if="title">
					{{ title }}
				</div>
				<div v-else style="width: calc(100% - 42px); height: 100%">
					<slot name="title" />
				</div>
				<div class="close-part" @click="onDialogCancel">
					<q-icon name="sym_r_close" size="20px" class="close-icon" />
				</div>
			</div>
			<div class="content-part">
				<slot name="content" />
			</div>
		</div>
	</q-card>
</template>

<script lang="ts" setup>
import { QDialog } from 'quasar';
import { PropType } from 'vue';

const props = defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	dialogRef: {
		type: Object as PropType<QDialog | undefined>,
		required: true
	}
});

const onDialogCancel = () => {
	props.dialogRef?.hide();
};
</script>

<style lang="scss" scoped>
.common-card {
	width: 100%;
	max-height: calc(100% - 100px);
	position: fixed;
	bottom: 0;
	// background-color: red;
	border-top-left-radius: 12px;
	border-top-right-radius: 12px;
	background-color: $background-1;

	.content-bg {
		height: 100%;
		padding-bottom: env(safe-area-inset-bottom);

		.title-part {
			width: 100%;
			height: 64px;
			color: $ink-1;
		}

		.content-part {
			width: 100%;
			height: calc(100% - 64px);
		}
	}
}
</style>
