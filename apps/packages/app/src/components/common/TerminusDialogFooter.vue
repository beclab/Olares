<template>
	<div v-if="isMobile" class="card-action row justify-between items-center">
		<confirm-button
			v-if="showCancel"
			class="cancel-button"
			:btn-title="t('cancel')"
			bg-classes="bg-background-1"
			textClasses="text-ink-2"
			@onConfirm="mobileCancel()"
		/>
		<confirm-button
			class="confirm-button"
			:btn-title="okText ? okText : t('confirm')"
			@onConfirm="onOK"
			:style="{ width: showCancel ? '48%' : '100%' }"
			v-if="!loading"
		/>
		<confirm-button
			class="confirm-button"
			:btn-title="t('loading')"
			:style="{ width: showCancel ? '48%' : '100%' }"
			v-else
		/>
	</div>

	<q-card-actions v-else class="row justify-end items-center q-mt-md q-mb-sm">
		<q-item
			v-if="showCancel"
			clickable
			dense
			class="but-cancel text-body3 ink-2 row justify-center items-center q-px-md q-mr-md"
			@click="onCancel"
		>
			{{ cancelText }}
		</q-item>
		<q-item
			clickable
			dense
			:class="wiseChannel ? 'but-creat-wise' : 'but-creat'"
			class="text-body3 yellow-default ink-on-brand-black row justify-center items-center q-px-md q-mr-sm"
			@click="onOK"
			v-if="!loading"
		>
			{{ okText }}
		</q-item>
		<q-item
			v-else
			dense
			:class="wiseChannel ? 'but-creat-wise' : 'but-creat'"
			class="text-body3 yellow-default ink-on-brand-black row justify-center items-center q-px-md q-mr-sm"
		>
			{{ t('files.loading') }}
		</q-item>
	</q-card-actions>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useQuasar } from 'quasar';

import ConfirmButton from './ConfirmButton.vue';
import { useDataStore } from '../../stores/data';
import { i18n } from '../../boot/i18n';
import { useI18n } from 'vue-i18n';

defineProps({
	okText: {
		type: String,
		default: i18n.global.t('submit'),
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
	}
});

const $q = useQuasar();
const emit = defineEmits(['close', 'submit']);
const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);
const store = useDataStore();
const { t } = useI18n();
const wiseChannel = ref(process.env.APPLICATION === 'WISE');

const onCancel = () => {
	emit('close');
};

const onOK = (e: any) => {
	emit('submit', e);
};

const mobileCancel = () => {
	store.closeHovers();
	emit('close');
};
</script>

<style scoped lang="scss">
.but-creat {
	border-radius: 8px;
	background: $yellow;
	color: $grey-10;
}

.but-creat-wise {
	border-radius: 8px;
	background: $orange-default;
	color: $ink-on-brand;
}
.but-cancel {
	border-radius: 8px;
	border: 1px solid $btn-stroke;
}

.card-action {
	margin: 20px;
	.cancel-button {
		width: 48%;
		border: 1px solid $separator;
	}
}
</style>
