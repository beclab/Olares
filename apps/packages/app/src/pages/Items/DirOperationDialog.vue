<template>
	<q-dialog
		class="code-dialog"
		position="bottom"
		ref="dialogRef"
		transition-show="jump-up"
		transition-hide="jump-down"
		transition-duration="300"
	>
		<div class="code bg-background-2">
			<div class="header text-ink-1 text-subtitle2">
				{{ t('vault_t.authenticator') }}
			</div>
			<div class="code-body row items-center justify-between">
				<div
					class="body-item column items-center justify-center"
					@click="createCode"
				>
					<q-icon name="sym_r_keyboard_onscreen" size="24px" />
					<div class="q-mt-sm">{{ t('vault_t.enter_key') }}</div>
				</div>

				<div
					class="body-item column items-center justify-center"
					@click="scanCode"
				>
					<q-icon name="sym_r_fullscreen" size="24px" />
					<div class="q-mt-sm">{{ t('vault_t.scan_code') }}</div>
				</div>
			</div>
		</div>
	</q-dialog>
</template>

<script setup lang="ts">
import { useDialogPluginComponent } from 'quasar';
import { useI18n } from 'vue-i18n';

const { dialogRef, onDialogOK } = useDialogPluginComponent();

const { t } = useI18n();

const createCode = () => {
	onDialogOK('createCode');
};

const scanCode = () => {
	onDialogOK('scanCode');
};
</script>

<style lang="scss" scoped>
.code-dialog {
	.code {
		width: 84% !important;
		height: 152px;
		border-radius: 12px;
		margin-bottom: 24px;
		overflow: hidden;
		padding: 0 20px;

		.header {
			width: 100%;
			height: 52px;
			line-height: 52px;
		}

		.code-body {
			width: 100%;
			display: flex;
			flex-wrap: wrap;
			gap: 10px;
		}

		.body-item {
			flex: 1 1 80px;
			height: 76px;
			border-radius: 8px;
			border: 1px solid $separator;
		}
	}
}
</style>
