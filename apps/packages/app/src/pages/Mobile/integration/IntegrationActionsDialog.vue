<template>
	<q-dialog
		ref="dialogRef"
		position="bottom"
		transition-show="jump-up"
		transition-hide="jump-down"
		transition-duration="500"
	>
		<terminus-dialog-display-content
			:title="t('files.operate')"
			:dialog-ref="dialogRef"
		>
			<template v-slot:content>
				<q-item
					clickable
					v-close-popup
					class="account-operation-item q-pb-lg"
					@click="deleteAction"
				>
					<div class="account-operation-div row justify-start items-center">
						<q-icon name="sym_r_delete" size="20px" />
						<div class="account-operation-text text-body3">
							{{ t('delete') }}
						</div>
					</div>
				</q-item>
				<div style="width: 100%; height: 20px"></div>
			</template>
		</terminus-dialog-display-content>
	</q-dialog>
</template>

<script setup lang="ts">
import { useDialogPluginComponent } from 'quasar';

import TerminusDialogDisplayContent from '../../../components/common/TerminusDialogDisplayContent.vue';

import { useI18n } from 'vue-i18n';

const { dialogRef, onDialogOK } = useDialogPluginComponent();

const { t } = useI18n();

const deleteAction = () => {
	showDeleteAccountDialog();
};

const showDeleteAccountDialog = async () => {
	onDialogOK('delete');
};
</script>

<style lang="scss" scoped>
.account-operation-item {
	height: 48px;
	min-height: 48px;
	margin: 0;
	padding-left: 0;
	padding-right: 0;

	.account-operation-div {
		width: 100%;
		height: 100%;
		padding: 8px;

		.account-operation-text {
			color: $ink-2;
			margin-left: 8px;
		}
	}
}
</style>
