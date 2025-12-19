<template>
	<div class="account_box">
		<div class="q-pa-md wrap account_center" style="word-wrap: break-word">
			<div class="boot_justify1">
				<div :style="{ width: '124px', height: '124px' }">
					<img :src="tokenStore.avatar_url" style="width: 100%; height: 100%" />
				</div>
			</div>
			<p class="get-text">{{ username }}</p>
			<p class="account_please_text">
				{{ tokenStore.olaresId }}
			</p>
			<div class="Account_input_from">
				<p class="Password_top">{{ t('step_label_password') }}</p>
				<q-input
					type="password"
					v-model="tokenStore.wizard.password"
					dense
					:disable="tokenStore.raw_login_loading"
					class="Account_input Account_Password_text"
					:placeholder="t('step_input_placeholder')"
					inputStyle="text-indent: 6px;"
				/>
				<p class="failFlag" v-if="failFlag">
					{{ t(failInfo.message) }}
					<a
						v-if="failInfo.showLink"
						href="https://docs.olares.xyz/manual/get-started/activate-olares.html"
						target="_blank"
						>{{ t('err_learn_more') }}.</a
					>
				</p>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useTokenStore } from '../../../stores/wizard-step';
import { AccountErrMessage } from '../../../utils/constants';

const { t } = useI18n();
const tokenStore = useTokenStore();
const failFlag = ref(false);
const failInfo = ref(AccountErrMessage.CLOUD_ADMIN);
const username =
	tokenStore.olaresId &&
	tokenStore.olaresId.slice(0, tokenStore.olaresId.indexOf('@'));

const click = async (): Promise<boolean> => {
	if (!tokenStore.wizard.password) {
		failFlag.value = true;
		return false;
	}

	failFlag.value = false;
	tokenStore.raw_login_loading = true;
	try {
		tokenStore.wizard.username = tokenStore.olaresId;
		const password = tokenStore.wizard.password || '';

		await tokenStore.raw_login(tokenStore.olaresId, password);
		tokenStore.raw_login_loading = false;
		tokenStore.setStep(2);

		return true;
	} catch (err) {
		failFlag.value = true;
		return false;
	} finally {
		tokenStore.raw_login_loading = false;
	}
};

const checkFailInfo = () => {
	const origin = window.location.origin;
	if (tokenStore.user.selfhosted) {
		if (origin.indexOf('30180') > -1) {
			failInfo.value = AccountErrMessage.SELFHOSTED_ADMIN;
		} else {
			failInfo.value = AccountErrMessage.SELFHOSTED_SUB;
		}
	} else {
		if (origin.indexOf('30180') > -1) {
			failInfo.value = AccountErrMessage.CLOUD_ADMIN;
		} else {
			failInfo.value = AccountErrMessage.CLOUD_SUB;
		}
	}
};

defineExpose({
	click
});

onMounted(() => {
	tokenStore.wizard.password = '';
	checkFailInfo();
});
</script>
