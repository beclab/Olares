<template>
	<q-dialog ref="dialogRef">
		<q-card class="common-dialog">
			<div class="content-root">
				<DialogHeader
					:title="title"
					@close-action="dialogRef?.hide()"
				></DialogHeader>
				<div class="row items-center justify-between" v-if="message.length > 0">
					<div class="userName setting-medium-text">
						{{ message }}
					</div>
				</div>

				<div class="confirm-bg row items-center justify-end">
					<div class="cancel-btn" @click="cancelAction" v-if="useCancel">
						<div class="title setting-major-text">Cancel</div>
					</div>
					<div class="confirm-btn" @click="sureAction">
						<div class="title setting-major-text">OK</div>
					</div>
				</div>
			</div>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import { useDialogPluginComponent } from 'quasar';
import DialogHeader from './DialogHeader.vue';

withDefaults(
	defineProps<{
		title: string;
		message: string;
		useCancel: boolean;
	}>(),
	{
		title: '',
		message: '',
		useCancel: true
	}
);

const cancelAction = () => {
	onDialogCancel();
	dialogRef.value?.hide();
};

const sureAction = () => {
	onDialogOK();
	dialogRef.value?.hide();
};

const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();
</script>

<style scoped lang="scss">
.content-root {
	padding: 0 16px 24px;
}

.userName {
	font-size: 12px;
	font-weight: 400;
	line-height: 32px;
	height: 32px;
}

.userNameInput {
	width: 260px;
	// height: 32px;
	border-radius: 4px;
	border: 1px solid #e0e2e9;
	padding-left: 12px;
}

.confirm-bg {
	margin-top: 16px;

	.confirm-btn {
		height: 24px;
		background: #285afd;
		border-radius: 4px;
		width: auto;
	}

	.cancel-btn {
		height: 24px;
		background: rgba(0, 0, 0, 0.4);
		border-radius: 4px;
		width: auto;
		margin-right: 12px;
	}

	.title {
		font-size: 12px;
		font-weight: 400;
		line-height: 24px;
		margin-left: 16px;
		margin-right: 16px;
		color: #fff;
	}

	.confirm-btn:hover {
		background-color: #285acc;
	}
}
</style>
