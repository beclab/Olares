<template>
	<div
		class="bottom-root"
		:style="{
			'--btn_height': isBex || $q.platform.is.capacitor ? '48px' : '48px'
		}"
	>
		<q-btn
			:class="getCurrentStatusClass()"
			flat
			padding="0px"
			@click="confirmAction()"
			:disable="btnStatus === ConfirmButtonStatus.disable"
			no-caps
			class="row items-center justify-center"
		>
			<q-icon
				v-if="btnIcon.length"
				:name="btnIcon"
				size="20px"
				class="q-mr-sm"
				color="grey-10"
			/>
			<div>
				{{ btnTitle.length > 0 ? btnTitle : t('confirm') }}
			</div>
		</q-btn>
	</div>
</template>

<script setup lang="ts">
import { ConfirmButtonStatus } from '../../utils/constants';
import { useQuasar } from 'quasar';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
const $q = useQuasar();
const { t } = useI18n();

const props = defineProps({
	btnTitle: {
		type: String,
		required: false,
		default: ''
	},
	btnStatus: {
		type: Number,
		required: false,
		default: ConfirmButtonStatus.normal
	},
	bgClasses: {
		type: String,
		required: false,
		default: 'bg-yellow-default'
	},
	bgDisabledClasses: {
		type: String,
		required: false,
		default: 'bg-yellow-disabled'
	},
	textClasses: {
		type: String,
		required: false,
		default: 'text-grey-10'
	},
	btnIcon: {
		type: String,
		required: false,
		default: ''
	}
});

const confirmAction = () => {
	if (props.btnStatus === ConfirmButtonStatus.error) {
		emit('onError');
	} else if (props.btnStatus === ConfirmButtonStatus.normal) {
		emit('onConfirm');
	}
};

const isBex = ref(process.env.IS_BEX);

const emit = defineEmits(['onConfirm', 'onError']);

const getCurrentStatusClass = () => {
	const classes = [props.textClasses];

	classes.push(
		isBex.value || $q.platform.is.capacitor
			? 'confirm-btn-mobile text-subtitle1'
			: 'confirm-btn-desktop text-subtitle2'
	);

	if (props.btnStatus === ConfirmButtonStatus.error) {
		classes.push('bg-color-error');
	}

	if (props.btnStatus === ConfirmButtonStatus.normal) {
		classes.push(props.bgClasses);
	}

	if (props.btnStatus === ConfirmButtonStatus.disable) {
		classes.push(props.bgDisabledClasses);
	}

	return classes.join(' ');
};
</script>

<style lang="scss" scoped>
.bottom-root {
	height: var(--btn_height);
	width: 100%;
	border-radius: 8px;

	.confirm-btn-mobile {
		width: 100%;
		height: 100%;
		text-align: center;
	}

	.confirm-btn-desktop {
		width: 100%;
		height: 100%;
		border-radius: 8px;
	}

	.bg-color-error {
		background-color: $red;
	}

	.confirm-btn[disabled] {
		.q-btn {
			padding: 0;
		}
	}
}
</style>
