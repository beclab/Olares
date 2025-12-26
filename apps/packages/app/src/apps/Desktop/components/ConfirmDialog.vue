<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef" @hide="onCancel">
		<q-card class="card-container" v-if="tokenStore.deviceInfo.isMobile" flat>
			<div class="text-ink-1 text-subtitle2 row items-center justify-center">
				{{ title }}
			</div>

			<div class="dialog-desc left row item-center justify-center q-my-lg">
				{{ message }}
			</div>

			<div class="row item-center justify-between">
				<span
					class="cancel-mobile text-ink-2 q-mr-md"
					v-if="showCancel"
					@click="onCancel"
					>{{ t('buttons.cancel') }}</span
				>
				<span
					class="confirm-mobile bg-yellow-default text-ink-1"
					@click="submit"
					>{{ t('buttons.confirm') }}</span
				>
			</div>
		</q-card>

		<q-card class="card-container" v-else flat>
			<div class="dialog-header row items-center justify-between">
				<div class="text-ink-1">
					{{ title }}
				</div>
				<q-space />
				<q-btn
					dense
					flat
					icon="close"
					size="sm"
					color="ink-3"
					@click="onCancel"
				>
					<q-tooltip>{{ t('buttons.close') }}</q-tooltip>
				</q-btn>
			</div>

			<div class="dialog-desc left row item-center justify-start">
				<img class="img q-mr-md" :src="icon" v-if="icon" />
				<div class="message text-grey-8">
					{{ message }}
				</div>
			</div>

			<span class="confirm bg-blue-default text-white" @click="submit">{{
				t('buttons.confirm')
			}}</span>

			<span
				class="cancel text-ink-2 q-mr-md"
				v-if="showCancel"
				@click="onCancel"
				>{{ t('buttons.cancel') }}</span
			>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useDialogPluginComponent } from 'quasar';
import { ref } from 'vue';
import { useTokenStore } from '../../../stores/desktop/token';

defineProps({
	title: {
		type: String,
		default: 'Confirmation',
		required: false
	},
	message: {
		type: String,
		default: 'This is a message',
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
		width: 400px;
		border-radius: 12px;
		padding: 20px;

		.dialog-header {
			width: 100%;
			height: 36px;
			line-height: 36px;
			font-size: 16px;
			font-weight: 500;
			line-height: 24px;
			text-align: left;
			margin-bottom: 12px;
		}

		.dialog-desc {
			.img {
				width: 32px;
				height: 32px;
				border-radius: 8px;
			}
			.message {
				max-width: calc(100% - 45px);
				font-size: 12px;
				line-height: 16px;
			}
		}

		.confirm {
			width: 68px;
			text-align: center;
			margin-top: 40px;
			padding: 8px 12px;
			border-radius: 8px;
			font-size: 12px;
			line-height: 16px;
			float: right;
			cursor: pointer;
			&:hover {
				opacity: 0.8;
			}
		}

		.cancel {
			width: 68px;
			text-align: center;
			margin-top: 40px;
			padding: 8px 0;
			border-radius: 8px;
			font-size: 12px;
			line-height: 16px;
			float: right;
			cursor: pointer;
			border: 1px solid $btn-stroke;
			&:hover {
				opacity: 0.8;
			}
		}

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
