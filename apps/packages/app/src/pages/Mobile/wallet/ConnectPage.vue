<template>
	<div class="connect-root">
		<connect-header
			:title="t('notification')"
			:dappLogo="dappLogo"
			:dappUrl="dappUrl"
			:action="t('requesting_connection')"
			:showDappInfo="true"
		/>
		<q-scroll-area class="scroll-area"></q-scroll-area>
		<div
			class="row justify-between items-center"
			style="margin-top: 20px; width: 100%"
		>
			<confirm-button
				style="width: 45%"
				text-classes="text-color-title"
				bg-classes="bg-color-white"
				:btn-title="t('cancel')"
				@onConfirm="rejectAction"
			/>
			<confirm-button
				style="width: 45%"
				:btn-title="t('ok')"
				@onConfirm="approveAction"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useApproval } from './approval';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import ConnectHeader from './ConnectHeader.vue';
import { useBexStore } from '../../../stores/bex';
import { useUserStore } from '../../../stores/user';
import { getDID } from '../../../did/did-key';
import { useI18n } from 'vue-i18n';

const $router = useRouter();
const store = useBexStore();
const { t } = useI18n();

const dappLogo = ref('');
const dappUrl = ref('');

store.controller.getApproval().then((approval) => {
	if (approval) {
		const origin = approval.data.session.origin;
		const icon = approval.data.session.icon;
		if (origin && typeof origin === 'string') {
			dappUrl.value = origin;
		}
		if (icon && typeof icon === 'string') {
			dappLogo.value = icon;
		}
	}
});

const rejectAction = async () => {
	const { rejectApproval } = useApproval($router);
	await rejectApproval();
};

const approveAction = async () => {
	const userStore = useUserStore();
	let mnemonicItem = userStore.current_mnemonic;
	const didKey: string = await getDID(mnemonicItem?.mnemonic);
	const { resolveApproval } = useApproval($router);
	await resolveApproval({ didKey: didKey });
};
</script>

<style lang="scss" scoped>
.connect-root {
	width: 100%;
	height: 100%;

	.scroll-area {
		width: calc(100% - 40px);
		height: 300px;
		margin-left: 20px;

		.params {
			margin-top: 30px;
			margin-bottom: 10px;
			// margin-left: 20px;
			text-overflow: ellipsis;
			word-break: break-word;
		}
	}
}
</style>
