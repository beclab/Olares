<template>
	<div
		v-if="platform === 'mobile'"
		class="card-action row justify-between items-center q-ma-lg"
	>
		<q-btn
			class="but-mobile-cancel"
			v-if="cancel"
			no-caps
			flat
			:style="{ width: cancel ? '48%' : '100%' }"
			@click="onCancel"
		>
			{{ cancel === true ? 'Cancel' : cancel }}
		</q-btn>

		<q-btn
			v-if="okLoading"
			class="but-mobile-confirm"
			no-caps
			flat
			:style="{ width: cancel ? '48%' : '100%' }"
		>
			{{ okLoading === true ? 'Loading' : okLoading }}
		</q-btn>
		<q-btn
			v-else
			class="but-mobile-confirm"
			no-caps
			flat
			:style="{ width: cancel ? '48%' : '100%' }"
			@click="onOK"
		>
			{{ ok === true ? 'Confirm' : ok }}
		</q-btn>
	</div>

	<q-card-actions v-else class="row justify-end items-center q-mt-lg q-pa-none">
		<q-btn
			v-if="cancel"
			clickable
			flat
			type="reset"
			class="but-cancel row justify-center items-center q-px-md q-mr-md"
			@click="onCancel"
			no-caps
		>
			{{ cancel === true ? $t('CANCEL') : cancel }}
		</q-btn>
		<q-btn
			clickable
			flat
			class="but-creat row justify-center items-center q-px-md"
			@click="onOK"
			:style="okStyle"
			type="submit"
			v-if="!okLoading"
			no-caps
		>
			{{ ok === true ? $t('CONFIRM') : ok }}
		</q-btn>
		<q-item
			v-else
			dense
			class="but-creat row justify-center items-center q-px-md"
			:style="okStyle"
		>
			{{ okLoading === true ? 'Loading' : okLoading }}
		</q-item>
	</q-card-actions>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useQuasar } from 'quasar';

import ConfirmButton from './ConfirmButton.vue';

defineProps({
	platform: {
		type: String,
		default: 'web',
		required: false
	},
	ok: {
		type: [String, Boolean],
		default: true,
		required: false
	},
	okStyle: {
		type: Object as () => any,
		required: false
	},
	cancel: {
		type: [String, Boolean],
		default: false,
		required: false
	},
	okLoading: {
		type: [String, Boolean],
		default: false,
		required: false
	}
});

const $q = useQuasar();
const emit = defineEmits(['close', 'submit']);

const onCancel = () => {
	emit('close');
};

const onOK = (e: any) => {
	// emit('submit', e);
};
</script>

<style scoped lang="scss">
.but-creat {
	border-radius: 8px;
	font-weight: 500;
	font-size: 12px;
	background: $yellow-6;
	color: $yellow-default;
}
.but-cancel {
	border-radius: 8px;
	font-weight: 500;
	font-size: 12px;
	border: 1px solid $btn-stroke;
	color: $ink-2;
}

.card-action {
	.cancel-button {
		width: 48%;
	}
}

.but-mobile-cancel {
	height: 48px;
	font-size: 16px;
	font-weight: 500;
	border-radius: 8px;
	border: 1px solid $btn-stroke;
	color: $ink-2;
}

.but-mobile-confirm {
	height: 48px;
	font-size: 16px;
	font-weight: 500;
	border-radius: 8px;
	background: $yellow-6;
}
</style>
