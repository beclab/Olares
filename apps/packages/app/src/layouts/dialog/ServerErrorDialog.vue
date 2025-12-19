<template>
	<q-dialog ref="root" @hide="onDialogHide">
		<q-card class="q-dialog-plugin row root">
			<div class="text-color-title text-h5 title">{{ title }}</div>

			<div class="row iterm-center justify-center calcelButton">
				<q-btn
					class="bg-grep-11 text-ink-2 cancel"
					:label="confirmTxt"
					@click="onCancelClick"
				/>
			</div>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useMenuStore } from '../../stores/menu';
defineProps({
	confirmTxt: {
		type: String,
		default: 'OK',
		required: false
	},
	cancelTxt: {
		type: String,
		default: 'Cancel',
		required: false
	}
});
const emit = defineEmits(['hide']);
const root = ref<any>(null);
const store = useMenuStore();

const title = ref('Can not connect Terminus');

const hide = () => {
	store.dialogShow = false;
	root.value.hide();
};

const onDialogHide = () => {
	emit('hide');
};

const onCancelClick = () => {
	hide();
};
</script>

<style lang="scss" scoped>
.root {
	border-radius: 10px;
	padding: 20px;
	.title {
		text-align: center;
		margin: 10px auto 20px;
	}

	.calcelButton {
		width: 100%;
		.confirm {
			width: 150px;
			height: 48px;
			border-radius: 10px;
			border: 0;
			&:before {
				box-shadow: none;
			}
		}
		.cancel {
			width: 150px;
			height: 48px;
			border-radius: 10px;
			box-shadow: none;
			&:before {
				box-shadow: none;
			}
		}
	}
}
</style>
