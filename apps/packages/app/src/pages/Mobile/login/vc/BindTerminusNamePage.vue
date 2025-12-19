<template>
	<div class="bind-terminus-name-root">
		<BindTerminusVCContent
			class="terminus-name"
			:has-btn="true"
			:btn-title="t('continue')"
			:btn-status="btnStatus"
			@onConfirm="precheckTerminusName"
		>
			<template v-slot:content>
				<terminus-edit
					v-model="terminusNameRef"
					:label="t('Create Olares ID')"
					class="q-mt-lg"
					@update:model-value="setButtonStatus"
					input-type="email"
				/>
				<div
					class="text-body3 q-mt-sm"
					:class="
						btnStatus == ConfirmButtonStatus.error
							? 'text-negative'
							: 'text-ink-3'
					"
				>
					{{
						t(
							'Must be 8 to 63 characters long, containing only numbers [0-9] and lowercase letters [a-z].'
						)
					}}
				</div>
				<div
					class="reminder bg-background-3 q-mt-xl q-pa-md text-body2 text-ink-1 row"
				>
					<q-icon
						name="sym_r_error"
						size="16px"
						color="light-blue-default"
						class="q-mr-sm"
					/>
					<div
						class="q-mb-sm text-ink-2"
						style="width: calc(100% - 25px)"
						v-html="
							t(
								'Olares ID is your unique identifier within the Olares ecosystem. lt is a decentralizedID, meaning you have full control over it-noother individual or organization including Olares can access or control it.'
							)
						"
					/>
				</div>
			</template>
		</BindTerminusVCContent>
		<div class="bind-terminus-name-root__img row items-center justify-end">
			<TerminusChangeUserHeader
				:scan="true"
				:redefinedAvatar="true"
				@redefinedAvatarAction="bindVC"
			>
				<template v-slot:avatar>
					<q-icon name="sym_r_display_settings" size="24px" color="grey-8" />
				</template>
			</TerminusChangeUserHeader>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import BindTerminusVCContent from './BindTerminusVCContent.vue';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import TerminusEdit from '../../../../components/common/TerminusEdit.vue';
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '../../../../stores/user';
// import { getDomainNameByType } from '../../../../utils/contact';
import {
	basicTerminusNameRule,
	basicTerminusNameMinLength,
	basicTerminusNameMaxLength,
	getBasicTerminusName,
	precheckDidHadBindTerminusName
} from './BindVCBusiness';
import { useQuasar } from 'quasar';
import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';
import TerminusChangeUserHeader from '../../../../components/common/TerminusChangeUserHeader.vue';
import { getNativeAppPlatform } from 'src/application/platform';
// import UserStatusCommonDialog from '../../../../components/userStatusDialog/UserStatusCommonDialog.vue';

const { t } = useI18n();

const terminusNameRef = ref('');

const btnStatus = ref(ConfirmButtonStatus.disable);

const normalRule = new RegExp('^[a-z0-9]+$');
const terminusRule = new RegExp(basicTerminusNameRule);
const userStore = useUserStore();

const $q = useQuasar();
const router = useRouter();

function setButtonStatus() {
	if (!terminusNameRef.value) {
		btnStatus.value = ConfirmButtonStatus.disable;
		return;
	}

	if (!normalRule.test(terminusNameRef.value)) {
		btnStatus.value = ConfirmButtonStatus.error;
		return;
	}

	if (terminusNameRef.value.length < basicTerminusNameMinLength) {
		btnStatus.value = ConfirmButtonStatus.disable;
		return;
	}
	if (terminusNameRef.value.length > basicTerminusNameMaxLength) {
		btnStatus.value = ConfirmButtonStatus.error;
		return;
	}
	btnStatus.value = ConfirmButtonStatus.normal;
}

async function precheckTerminusName() {
	if (!terminusRule.test(terminusNameRef.value)) {
		return;
	}
	if (!(await userStore.unlockFirst())) {
		return;
	}

	if (!$q.platform.is.nativeMobile) {
		notifyFailed('not support');
		return;
	}

	// const reminderMessage =
	// 	t('Your Olares ID will be {olaresID}', {
	// 		olaresID:
	// 			terminusNameRef.value +
	// 			'@' +
	// 			getDomainNameByType(userStore.defaultDomain)
	// 	}) +
	// 	(userStore.defaultDomain != 'cn'
	// 		? '<br>' +
	// 		  '<br>' +
	// 		  t(
	// 				'If you are in mainland China, make sure you use olares.cn as your default domain for best activation & access experience.'
	// 		  )
	// 		: '');

	// $q.dialog({
	// 	component: UserStatusCommonDialog,
	// 	componentProps: {
	// 		title: t('Confirm your Olares ID'),
	// 		message: reminderMessage,
	// 		messageClasses: 'text-ink-2',
	// 		messageCenter: true,
	// 		// addSkip: true,
	// 		btnTitle: t('confirm'),
	// 		resetDomainTitle: t('Change default domain'),
	// 		resetDomainClasses: 'item-common-border text-subtitle1 text-ink-1',
	// 		setDomain: true
	// 	}
	// }).onOk((value: string) => {
	// 	if (value == 'reset') {
	// 		bindVC();
	// 	} else if (value == 'confirm') {
	// 		createTerminusName();
	// 	}
	// });
	createTerminusName();
}

const bindVC = () => {
	router.push({
		path: '/bind_vc'
	});
};

const createTerminusName = async () => {
	const deviceId = await getNativeAppPlatform().getDeviceId();
	const appVersion = (await getNativeAppPlatform().getDeviceInfo()).appVersion;
	if (deviceId.length == 0) {
		return;
	}
	$q.loading.show();
	await getBasicTerminusName(
		terminusNameRef.value,
		deviceId,
		appVersion,
		window.navigator.userAgent,
		{
			onSuccess() {
				$q.loading.hide();
				userStore.isNewCreateUser = true;
				router.replace({
					path: '/Activate/1/1'
				});
			},
			onFailure(message: string) {
				$q.loading.hide();
				notifyFailed(message);
			}
		}
	);
};

onMounted(async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	const hadBind = await precheckDidHadBindTerminusName();
	if (hadBind) {
		// router.replace({ path: '/connectLoading' });
		userStore.isNewCreateUser = true;
		router.replace({
			path: '/Activate/1/1'
		});
	}
});
</script>

<style lang="scss" scoped>
.bind-terminus-name-root {
	width: 100%;
	height: 100%;
	position: relative;

	.terminus-name {
		width: 100%;
		height: 100%;
	}

	.reminder {
		border: 1px solid $separator;
		border-radius: 12px;
	}

	&__img {
		width: 100%;
		height: 40px;
		position: absolute;
		top: 20px;
	}
}
</style>
