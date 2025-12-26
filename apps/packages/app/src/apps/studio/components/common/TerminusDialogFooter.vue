<template>
	<q-card-actions class="row justify-end items-center q-mt-md q-mb-sm">
		<q-item
			v-if="showCancel"
			clickable
			dense
			class="but-cancel row justify-center items-center q-px-md q-mr-md"
			@click="onCancel"
		>
			{{ cancelText }}
		</q-item>
		<q-item
			clickable
			dense
			class="but-creat row justify-center items-center q-px-md q-mr-sm"
			@click="onOK"
			v-if="!loading"
		>
			{{ okText }}
		</q-item>
		<q-item
			v-else
			dense
			class="but-creat row justify-center items-center q-px-md q-mr-sm"
		>
			Loading
		</q-item>
	</q-card-actions>
</template>

<script lang="ts" setup>
defineProps({
	okText: {
		type: String,
		default: 'Sumbit',
		required: false
	},
	cancelText: {
		type: String,
		default: 'Cancel',
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
	background: $teal-default;
	color: #ffffff;
}
.but-cancel {
	border-radius: 8px;
	font-weight: 500;
	font-size: 12px;
	border: 1px solid $separator;
	color: $ink-1;
}

.card-action {
	margin: 20px;
	.cancel-button {
		width: 48%;
	}
}
</style>
