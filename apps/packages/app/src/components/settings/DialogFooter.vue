<template>
	<AdaptiveLayout>
		<template v-slot:pc>
			<div class="row justify-end q-mt-md confirm-bg">
				<q-btn
					v-if="hasCancel"
					dense
					flat
					class="cancel-btn q-px-md"
					:label="cancelText.length > 0 ? cancelText : t('cancel')"
					@click="cancelAction"
				/>
				<q-btn
					v-if="hasConfirm"
					dense
					flat
					:class="isReminder ? 'reminder-btn q-px-md' : 'confirm-btn q-px-md'"
					:label="confirmText"
					@click="confirmAction"
					:disable="confirmDisable"
				/>
			</div>
		</template>
		<template v-slot:mobile>
			<div class="row items-center justify-between confirm-bg-mobile">
				<q-btn
					v-if="hasCancel"
					dense
					flat
					class="cancel-btn-mobile"
					:class="{
						'button-full-width': !hasConfirm,
						'button-half-width': hasConfirm
					}"
					:label="cancelText.length > 0 ? cancelText : t('cancel')"
					@click="cancelAction"
				/>

				<q-btn
					v-if="hasConfirm"
					dense
					flat
					:class="{
						'reminder-btn-mobile': isReminder,
						'confirm-btn-mobile': !isReminder,
						'button-full-width': !hasCancel,
						'button-half-width': hasCancel
					}"
					:label="confirmText"
					@click="confirmAction"
					:disable="confirmDisable"
				/>
			</div>
		</template>
	</AdaptiveLayout>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { i18n } from 'src/boot/i18n';
import AdaptiveLayout from './AdaptiveLayout.vue';
defineProps({
	cancelText: {
		type: String,
		required: false,
		default: ''
	},
	confirmText: {
		type: String,
		required: false,
		default: i18n.global.t('ok')
	},
	confirmDisable: {
		type: Boolean,
		required: false,
		default: false
	},
	hasCancel: {
		type: Boolean,
		required: false,
		default: true
	},
	isReminder: {
		type: Boolean,
		required: false,
		default: false
	},
	hasConfirm: {
		type: Boolean,
		required: false,
		default: true
	}
});

const cancelAction = () => {
	emit('cancelAction');
};
const confirmAction = () => {
	emit('confirmAction');
};

const emit = defineEmits(['cancelAction', 'confirmAction']);

const { t } = useI18n();
</script>

<style scoped lang="scss">
.confirm-bg {
	margin-top: 16px;

	.confirm-btn {
		width: auto;
		margin-top: 20px;
		min-width: 50px;
	}

	.cancel-btn {
		margin-top: 20px;
		margin-right: 8px;
	}

	.confirm-btn:hover {
		background-color: $blue-8;
	}

	.reminder-btn:hover {
		background-color: $red-8;
	}
}

.confirm-bg-mobile {
	margin-top: 36px;
	width: 100%;
	height: 48px;
	.cancel-btn-mobile {
		height: 100%;
	}
	.confirm-btn-mobile {
		height: 100%;
	}
	.reminder-btn-mobile {
		height: 100%;
		min-width: 50px;
	}

	.button-full-width {
		width: 100%;
	}

	.button-half-width {
		width: 47%;
	}
}
</style>
