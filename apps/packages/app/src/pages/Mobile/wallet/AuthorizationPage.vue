<template>
	<div class="authorization-root">
		<connect-header
			:title="t('notification')"
			:dappLogo="dappLogo"
			:dappUrl="dappUrl"
			:action="t('requesting_call_larepass')"
			:showDappInfo="true"
		/>
		<!-- <div class="text-ink-1 text-h6 q-ml-md">
			{{ approvalTypeRef }}
		</div> -->
		<q-scroll-area class="scroll-area">
			<!-- <div class="text-ink-2 text-body2 params">
				<div>{{ paramsRef.url }}</div>
				<div>{{ paramsRef.username }}</div>
				<div>{{ paramsRef.password }}</div>
			</div> -->
		</q-scroll-area>
		<div
			class="row justify-between items-center"
			style="
				margin-top: 40px;
				width: 100%;
				padding-left: 20px;
				padding-right: 20px;
			"
		>
			<confirm-button
				style="width: 45%"
				bg-classes="bg-white"
				:btn-title="t('reject')"
				@onConfirm="rejectAction"
				class="button-cancel"
			/>
			<confirm-button
				style="width: 45%"
				:btn-title="t('Approve')"
				@onConfirm="approveAction"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import ConnectHeader from './ConnectHeader.vue';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import { useRouter, useRoute } from 'vue-router';
import { ref, watchEffect } from 'vue';
import { useApproval } from './approval';
import { useI18n } from 'vue-i18n';
import { updateUIToAddWeb } from '../../../platform/addItem';
import { useBexStore } from '../../../stores/bex';
import { browser } from 'src/platform/interface/bex/browser/target';
import { APPROVAL_TYPE } from 'src/platform/interface/bex/provider/utils';

const $router = useRouter();
const route = useRoute();
const { t } = useI18n();

const dappLogo = ref('');
const dappUrl = ref('');
const approvalTypeRef = ref();
const paramsRef = ref();
const sessionId = ref();

const store = useBexStore();
const displayRef = ref();

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
		approvalTypeRef.value = approval.data.approvalType;
		paramsRef.value = approval.data.params;
		sessionId.value = approval.data.sessionId;
		if (approvalTypeRef.value == APPROVAL_TYPE.ADD_VAULT_ITEM) {
			displayRef.value = `URL: ${approval.data.params.url}<br>Username:${approval.data.params.username}\
			<br>Password:${approval.data.params.password}`;
		} else {
			displayRef.value = paramsRef.value;
		}
	}
});

watchEffect(() => {
	const users = route.query.users
		? JSON.parse(route.query.users as string)
		: [];
	paramsRef.value = users[0];
});

const rejectAction = async () => {
	const { rejectApproval } = useApproval($router);
	await rejectApproval();
};

const approveAction = async () => {
	const { resolveApproval } = useApproval($router);

	// await updateUIToAddWeb(
	// 	paramsRef.value.url,
	// 	$router,
	// 	paramsRef.value.username,
	// 	paramsRef.value.password,
	// 	true
	// );
	// await resolveApproval();

	if (approvalTypeRef.value === APPROVAL_TYPE.SIGN_PRESENTATION) {
		await resolveApproval({ routerPath: '/VC_card_list' });
	} else if (approvalTypeRef.value == APPROVAL_TYPE.ADD_VAULT_ITEM) {
		await updateUIToAddWeb(
			paramsRef.value.url,
			$router,
			paramsRef.value.username,
			paramsRef.value.password,
			true
		);

		if (sessionId.value) {
			await browser.tabs.sendMessage(sessionId.value, {
				command: 'saveCipherAttemptCompleted'
			});
		}

		await resolveApproval();
	} else {
		await resolveApproval();
	}
};
</script>

<style scoped lang="scss">
.authorization-root {
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
	.button-cancel {
		border: 1px solid $grey-2;
	}
}
</style>
