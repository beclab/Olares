<template>
	<terminus-title-bar :hookBackAction="true" @onReturnAction="onReturn" />
	<BindTerminusVCContent
		style="height: calc(100% - 56px)"
		:has-btn="true"
		@on-confirm="onConfirm"
		:btn-status="btnStatusRef"
	>
		<template v-slot:desc>
			{{
				step === 1
					? t('bind_a_olares_space_to_continue')
					: t('after_verification_code_bind_olares')
			}}
		</template>

		<template v-slot:content>
			<div v-if="step == 1">
				<terminus-edit
					v-model="email"
					:label="t('olares_space_account')"
					:show-password-img="false"
					class="q-mt-lg"
					@update:model-value="onEmailUpdate"
					input-type="email"
				/>
				<terminus-edit
					v-model="password"
					:label="t('password')"
					:show-password-img="true"
					class="q-mt-md"
					@update:model-value="onPasswordChange"
				/>
			</div>

			<div v-if="step == 2">
				<terminus-edit
					v-if="step === 2 && email2FA && !email2FAFinished"
					v-model="emailCode"
					:label="
						t('enter_the_6_digit_code_sent_to_code', {
							code: loginToken?.emailMask
						})
					"
					:show-password-img="false"
					class="q-mt-lg"
					:hint-text="t('email_verification_code')"
					@update:model-value="emailCodeChange"
				>
					<template v-slot:right>
						<div
							class="row justify-center items-center q-ml-md text-subtitle1 email-code"
							v-if="countEmailDown > 0"
						>
							<span class="countdown text-body2">{{ countEmailDown }}s</span>
						</div>
						<div class="resend text-body2" v-else @click="sendEmailCode">
							{{ emailFirstSendCode ? t('send_code') : t('resend') }}
						</div>
					</template>
				</terminus-edit>

				<terminus-edit
					v-if="step === 2 && phone2FA && !phone2FAFinished"
					v-model="phoneCode"
					:label="t('olares_space_account')"
					:show-password-img="false"
					class="q-mt-lg"
					:hint-text="t('phone_verification_code')"
					@update:model-value="phoneCodeChange"
				>
				</terminus-edit>

				<terminus-edit
					v-if="step === 2 && ga2FA && !ga2FAFinished"
					v-model="ga2Code"
					:label="t('olares_space_account')"
					:show-password-img="false"
					class="q-mt-lg"
					:hint-text="t('ga_verification_code')"
					@update:model-value="ga2CodeChange"
				>
				</terminus-edit>
			</div>
		</template>
	</BindTerminusVCContent>
</template>

<script lang="ts" setup>
import { ref, onBeforeUnmount } from 'vue';
import { Loading } from 'quasar';
import { useCloudStore, TokenData } from '../../../../stores/cloud';
import { useUserStore } from '../../../../stores/user';
import api from 'axios';
import md5 from 'js-md5';
import { UserItem } from '@didvault/sdk/src/core';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import BindTerminusVCContent from '../../login/vc/BindTerminusVCContent.vue';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import TerminusEdit from '../../../../components/common/TerminusEdit.vue';
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';

const EMAIL_CODE_INTERVAL = 60; // requesting email code has a 120second interval

const { t } = useI18n();
const router = useRouter();
const step = ref(1);

const password = ref('');
const email = ref('');

const btnStatusRef = ref(ConfirmButtonStatus.disable);

const onEmailUpdate = () => {
	calBtnStatus();
};

const onPasswordChange = () => {
	calBtnStatus();
};

const calBtnStatus = () => {
	if (step.value == 1) {
		btnStatusRef.value =
			email.value.length > 0 && password.value.length > 0
				? ConfirmButtonStatus.normal
				: ConfirmButtonStatus.disable;
		return;
	}

	const emailEnable = email2FA.value && !email2FAFinished.value;

	const phoneEnable = email2FA.value && !phone2FAFinished.value;

	const ga2Enable = email2FA.value && !ga2FAFinished.value;

	btnStatusRef.value =
		(!emailEnable || (emailEnable && emailCode.value.length > 0)) &&
		(!phoneEnable || (phoneEnable && phoneCode.value.length > 0)) &&
		(!ga2Enable || (ga2Enable && ga2Code.value.length > 0))
			? ConfirmButtonStatus.normal
			: ConfirmButtonStatus.disable;
};

let email2FA = ref(false);
let phone2FA = ref(false);
let ga2FA = ref(false);
let email2FAFinished = ref(false);
let phone2FAFinished = ref(false);
let ga2FAFinished = ref(false);
const cloudStore = useCloudStore();
const loginToken = ref<TokenData | null>(null);

const emailCode = ref('');
const phoneCode = ref('');
const ga2Code = ref('');

const emailFirstSendCode = ref(true);

const emailCodeChange = () => {
	calBtnStatus();
};

const phoneCodeChange = () => {
	calBtnStatus();
};

const ga2CodeChange = () => {
	calBtnStatus();
};

const onConfirm = () => {
	if (step.value == 1) {
		handleLogin();
		return;
	}
	handleSubmit();
};

const handleLogin = async () => {
	if (email.value.length === 0) {
		return;
	}
	if (password.value.length === 0) {
		return;
	}
	Loading.show();
	try {
		const ob1: any = {};
		ob1.item = email.value;
		ob1.password = md5(password.value);
		const res = await api.post(cloudStore.getUrl() + '/v1/user/login', ob1);

		loginToken.value = res.data;

		email2FA.value = loginToken.value!.emailEnable || false;
		phone2FA.value = loginToken.value!.phoneEnable || false;
		ga2FA.value = loginToken.value!.gaEnable || false;

		email2FAFinished.value = loginToken.value!.emailEnable ? false : true;
		phone2FAFinished.value = loginToken.value!.phoneEnable ? false : true;
		ga2FAFinished.value = loginToken.value!.gaEnable ? false : true;

		Loading.hide();
		step.value = 2;
	} catch (e: any) {
		console.error(e);
		Loading.hide();
		notifyFailed(e.message || t('prompts.error'));
	}
	calBtnStatus();
};

const emailSentLoading = ref(false);
const codeEmailSent = ref(false);
const countEmailDown = ref(0);

const sendEmailCode = async () => {
	emailSentLoading.value = true;
	const ob1: any = {};
	ob1.userid = loginToken.value!.userid;
	ob1.token = loginToken.value!.token;
	await api
		.post(cloudStore.getUrl() + '/v1/user/request_email_verify', ob1)
		.finally(() => {
			Loading.hide();
		});
	emailFirstSendCode.value = false;
	codeEmailSent.value = true;
	startEmailCountDown();
	emailSentLoading.value = false;
};

const countEmailDownInterval = ref();

function startEmailCountDown() {
	if (countEmailDownInterval.value !== '')
		clearInterval(countEmailDownInterval.value);

	countEmailDown.value = EMAIL_CODE_INTERVAL;

	countEmailDownInterval.value = setInterval(() => {
		countEmailDown.value -= 1;
		if (countEmailDown.value === 0) clearInterval(countEmailDownInterval.value);
	}, 1000);
}

onBeforeUnmount(() => {
	if (countEmailDownInterval.value !== '')
		clearInterval(countEmailDownInterval.value);
});

let secondVerifyID: any = null;
const userStore = useUserStore();

const handleSubmit = async () => {
	try {
		if (!loginToken.value) {
			return;
		}
		var ob: any = {};
		ob.userid = loginToken.value.userid;
		ob.token = loginToken.value.token;

		if (secondVerifyID) {
			ob.secondVerifyID = secondVerifyID;
		}

		if (email2FA.value) {
			if (email2FAFinished.value == false) {
				if (isSixDigit(emailCode.value)) {
					ob.emailCode = emailCode.value;
				} else {
					notifyFailed(
						t('errors.please_Input_6_digit_email_verification_code')
					);
					//this.$refs.emailCode.validate();
					return;
				}
			}
		}

		Loading.show();

		const response: any = await api.post(
			cloudStore.getUrl() + '/v1/user/verify',
			ob
		);

		Loading.hide();

		var res = response.data;

		if (response.code == 200 && loginToken.value) {
			const user: UserItem = userStore.users!.items.get(userStore.current_id!)!;
			user.cloud_id = loginToken.value.userid;
			// user.cloud_email = email.value;
			user.cloud_token = loginToken.value.token;
			// user.cloud_email_enable = loginToken.value.emailEnable || false;
			// user.cloud_email_mask = loginToken.value.emailMask || '';
			user.cloud_expired = loginToken.value.expired;
			await userStore.users!.items.update(user);
			await userStore.save();
			emit('loginSuccess');
		} else if (response.code == 507) {
			secondVerifyID = res.secondVerifyID;

			if (email2FA.value) {
				if (res.needEmail) {
					email2FAFinished.value = false;
				} else {
					email2FAFinished.value = true;
				}
			}

			if (phone2FA.value) {
				if (res.needPhone) {
					phone2FAFinished.value = false;
				} else {
					phone2FAFinished.value = true;
				}
			}

			if (ga2FA.value) {
				if (res.needGA) {
					ga2FAFinished.value = false;
				} else {
					ga2FAFinished.value = true;
				}
			}

			if (response.data.message != '' && response.data.message != null) {
				notifyFailed(response?.data.message);
			}
		}
	} catch (e: any) {
		Loading.hide();
		notifyFailed(e.message);
	}
};

const isSixDigit = (v: string) => {
	if (!v) {
		return false;
	}

	if (v.length != 6) {
		return false;
	}

	for (var i = 0; i < v.length; i++) {
		var charCode = v.charCodeAt(i);
		if (charCode < 48 || charCode > 57) {
			return false;
		}
	}

	return true;
};

const onReturn = () => {
	if (step.value > 1) {
		step.value -= 1;
		password.value = '';
		return;
	}
	router.go(-1);
};

const emit = defineEmits(['loginSuccess']);
</script>

<style lang="scss" scoped>
.resend {
	text-align: right;
	color: $blue-4;
}

.countdown {
	text-align: right;
	color: $grey-5;
}
</style>
