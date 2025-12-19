<template>
	<q-dialog ref="dialogRef" @hide="onDialogHide">
		<q-card class="q-dialog-plugin">
			<slot></slot>
			<q-card-actions align="right">
				<q-btn color="primary" label="OK" @click="onOKClick" />
				<q-btn color="primary" label="Cancel" @click="onDialogCancel" />
			</q-card-actions>
		</q-card>
	</q-dialog>
</template>

<script setup>
import { useDialogPluginComponent } from 'quasar';

const props = defineProps({});

defineEmits([...useDialogPluginComponent.emits]);

const { dialogRef, onDialogHide, onDialogOK, onDialogCancel } =
	useDialogPluginComponent();

function onOKClick() {
	onDialogOK();
}
</script>
