<template>
	<div class="terminus-unlock-page column justify-center items-center">
		<q-img :src="getRequireImage(logo)" :width="`${logoWidth}px`" />
		<div class="terminus-unlock-box column justify-start items-center">
			<span class="terminus-unlock-box__desc login-sub-title">{{
				detailText
			}}</span>
			<terminus-edit
				v-model="passwordRef"
				:label="t('password')"
				:show-password-img="true"
				class="terminus-unlock-box__edit"
				@update:model-value="onTextChange"
				@keyup.enter="loginByPassword(passwordRef)"
			/>
			<confirm-button
				class="terminus-unlock-box__button"
				:btn-title="t('unlock.title')"
				:btn-status="btnStatusRef"
				@onConfirm="loginByPassword(passwordRef)"
			/>
			<q-btn
				v-if="cancel"
				class="q-mt-md common-width cancel"
				flat
				no-caps
				@click="onCancelClick"
			>
				<div>{{ t('cancel') }}</div>
			</q-btn>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ConfirmButtonStatus } from '../../../utils/constants';
import TerminusEdit from '../../../components/common/TerminusEdit.vue';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import { getRequireImage } from '../../../utils/imageUtils';
import { useI18n } from 'vue-i18n';
import { unlockByPwd, unlockByDefaultPwd } from '../../../utils/UnlockBusiness';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { sendUnlock } from '../../../utils/bexFront';

defineProps({
	detailText: {
		type: String,
		required: false,
		default: ''
	},
	logo: {
		type: String,
		required: false,
		default: ''
	},
	cancel: {
		type: Boolean,
		default: true,
		required: false
	},
	logoWidth: {
		type: Number,
		default: 180,
		required: false
	}
});

const passwordRef = ref('');
const { t } = useI18n();
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);

function onTextChange() {
	btnStatusRef.value =
		passwordRef.value.length < 8 || passwordRef.value.length > 32
			? ConfirmButtonStatus.disable
			: ConfirmButtonStatus.normal;
}

const loginByPassword = async (password: string) => {
	await unlockByPwd(password, unlockResult);
};

const unlockResult = {
	onSuccess: async (data: any) => {
		sendUnlock();
		emit('unlockSuccess', data);
	},
	onFailure: (message: string) => {
		notifyFailed(message);
	}
};

const onCancelClick = () => {
	emit('cancel');
};
const emit = defineEmits(['unlockSuccess', 'cancel']);

onMounted(() => {
	unlockByDefaultPwd(unlockResult);
});
</script>

<style scoped lang="scss">
.terminus-unlock-page {
	width: 100%;
	height: 100%;
	background: $background-1;
	min-height: 350px;

	.terminus-unlock-box {
		width: 400px;
		margin-top: 32px;
		border-radius: 12px;
		padding: 20px;
		background: $background-2;
		border: 1px solid $separator;

		&__desc {
			margin-top: 12px;
			text-align: center;
		}

		&__edit {
			margin-top: 20px;
			width: 100%;
		}

		&__button {
			margin-top: 30px;
			width: calc(100%);
		}

		.cancel {
			height: 48px;
			width: calc(100%);
			color: $light-blue-default;
		}

		.item {
			border: 1px solid $separator;
			padding: 5rem;
			margin: 0.25rem;
			position: relative;
		}

		.header {
			position: absolute;
			top: 0;
			left: 0;
		}
	}
}
</style>
