<template>
	<div class="biometric-unlock-dialog column justify-center items-center">
		<div class="text-h5 biometric-unlock-dialog__title" v-if="title.length > 0">
			{{ title }}
		</div>
		<div
			class="biometric-unlock-dialog__desc q-mt-md"
			:class="messageClasses"
			:style="messageCenter ? 'text-align: center' : 'text-align: left'"
			v-html="message"
		></div>
		<div class="biometric-unlock-dialog__more" v-if="$slots.more">
			<slot name="more" />
		</div>
		<div
			class="biometric-unlock-dialog__group row justify-between items-center"
			v-if="!btnRedefined"
		>
			<confirm-button
				class="biometric-unlock-dialog__group__btn"
				:btn-title="btnTitle ?? t('confirm')"
				@onConfirm="onDialogOK"
			/>
		</div>
		<div
			v-else
			class="biometric-unlock-dialog__group row justify-between items-center"
		>
			<slot name="buttons" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import ConfirmButton from '../common/ConfirmButton.vue';
const { t } = useI18n();
defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	message: {
		type: String,
		required: false,
		default: ''
	},
	btnTitle: {
		type: String,
		required: false,
		default: ''
	},
	btnRedefined: {
		type: Boolean,
		required: false,
		default: false
	},
	messageClasses: {
		type: String,
		default: 'text-ink-3',
		required: false
	},
	messageCenter: {
		type: Boolean,
		required: false,
		default: false
	}
});

const onDialogOK = () => {
	emits('onDialogOK');
};

const emits = defineEmits(['onDialogOK']);
</script>

<style lang="scss" scoped>
.biometric-unlock-dialog {
	padding: 0px;

	&__title {
		color: $ink-1;
	}

	&__desc {
		width: 100%;
	}

	&__more {
		width: 100%;
		//
	}

	&__img {
		width: 280px;
		background-color: $grey-1;
		border-radius: 8px;
	}

	&__group {
		width: 100%;
		&__btn {
			margin-top: 40px;
			width: 100%;
		}
	}
}
</style>
