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

	<q-card-actions v-else class="row justify-end items-center q-mb-md">
		<q-item
			v-if="cancel"
			clickable
			dense
			class="but-cancel row justify-center items-center q-px-md q-mr-md"
			@click="onCancel"
		>
			{{ cancel === true ? 'Cancel' : cancel }}
		</q-item>
		<q-item
			clickable
			dense
			class="but-creat row justify-center items-center q-px-md q-mr-sm"
			@click="onOK"
			:style="okStyle"
			v-if="!okLoading"
		>
			{{ ok === true ? 'Confirm' : ok }}
		</q-item>
		<q-item
			v-else
			dense
			class="but-creat row justify-center items-center q-px-md q-mr-sm"
			:style="okStyle"
		>
			{{ okLoading === true ? 'Loading' : okLoading }}
		</q-item>
	</q-card-actions>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';

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
	emit('submit', e);
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
