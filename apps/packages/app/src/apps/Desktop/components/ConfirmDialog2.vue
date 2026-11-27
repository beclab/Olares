<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef" @hide="onCancel">
		<q-card class="card-container mobile" v-if="isMobile" flat>
			<div class="text-ink-1 text-subtitle2 row items-center justify-center">
				{{ title }}
			</div>

			<div class="dialog-desc left row item-center justify-center q-my-lg">
				{{ warningMessage || message }}
			</div>

			<div class="row item-center justify-between">
				<span
					class="cancel-mobile text-ink-2 q-mr-md"
					v-if="showCancel"
					@click="onCancel"
					>{{ cancelText || t('buttons.cancel') }}</span
				>
				<span
					class="confirm-mobile bg-yellow-default text-ink-1"
					@click="submit"
					>{{ confirmText || t('buttons.confirm') }}</span
				>
			</div>
		</q-card>

		<q-card class="card-container desktop" v-else flat>
			<div class="dialog-header row items-center justify-between">
				<div class="text-h6 text-ink-1">
					{{ title }}
				</div>
				<q-space />
				<q-btn
					dense
					flat
					icon="close"
					size="sm"
					color="ink-3"
					class="close-btn"
					@click="onCancel"
				/>
			</div>

			<div class="dialog-alert">
				<div class="text-orange-default text-body2">
					{{ warningMessage || message }}
				</div>
			</div>

			<div v-if="noteMessage" class="text-body2 text-ink-2 q-mt-lg">
				{{ noteMessage }}
			</div>

			<div class="dialog-actions">
				<q-btn
					v-if="showCancel"
					flat
					no-caps
					class="action-btn cancel-btn"
					@click="onCancel"
				>
					<span class="text-subtitle1 text-ink-2">{{
						cancelText || t('buttons.cancel')
					}}</span>
				</q-btn>
				<q-btn no-caps class="action-btn" color="blue-default" @click="submit">
					<span class="text-subtitle1 text-ink-on-brand">{{
						confirmText || t('buttons.confirm')
					}}</span>
				</q-btn>
			</div>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useDialogPluginComponent } from 'quasar';
import { computed, ref } from 'vue';
import { DeviceType } from '@bytetrade/core';
import { useTokenStore } from '../../../stores/desktop/token';

defineProps({
	title: {
		type: String,
		default: '',
		required: false
	},
	message: {
		type: String,
		default: '',
		required: false
	},
	warningMessage: {
		type: String,
		default: '',
		required: false
	},
	noteMessage: {
		type: String,
		default: '',
		required: false
	},
	confirmText: {
		type: String,
		default: '',
		required: false
	},
	cancelText: {
		type: String,
		default: '',
		required: false
	},
	icon: {
		type: String,
		default: '',
		required: false
	},
	showCancel: {
		type: Boolean,
		default: false,
		required: false
	}
});

const { t } = useI18n();
const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();
const tokenStore = useTokenStore();

const show = ref(true);
const isMobile = computed(
	() => tokenStore.deviceInfo.device !== DeviceType.DESKTOP
);

const submit = async () => {
	onDialogOK();
};

const onCancel = async () => {
	onDialogCancel();
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-container {
		width: 560px;
		max-width: min(90vw, 640px);
		border-radius: 12px;
		padding: 20px;
		background-color: $background-2;

		.dialog-header {
			width: 100%;
			min-height: 24px;
			margin-bottom: 20px;
		}

		.close-btn {
			margin-right: -4px;
		}

		.dialog-alert {
			border: 1px solid $orange-default;
			background: $orange-alpha;
			border-radius: 12px;
			padding: 20px;
		}

		.dialog-actions {
			margin-top: 20px;
			display: flex;
			justify-content: flex-end;
			gap: 12px;
		}
		.action-btn {
			height: 40px;
			min-width: 100px;
			border-radius: 8px;
			font-size: 16px;
			font-weight: 500;
			line-height: 24px;
			padding: 8px 20px;
		}
		.cancel-btn {
			border: 1px solid $btn-stroke;
		}
	}

	.mobile {
		width: 400px;
		box-shadow: none;
		.confirm-mobile {
			width: 46%;
			height: 48px;
			line-height: 48px;
			margin-top: 12px;
			text-align: center;
			border-radius: 8px;
			font-size: 16px;
			float: right;
			cursor: pointer;
			&:hover {
				opacity: 0.8;
			}
		}

		.cancel-mobile {
			width: 46%;
			height: 48px;
			line-height: 48px;
			margin-top: 12px;
			text-align: center;
			border-radius: 8px;
			font-size: 16px;
			float: right;
			cursor: pointer;
			border: 1px solid $btn-stroke;
			&:hover {
				opacity: 0.8;
			}
		}
	}
}
</style>
