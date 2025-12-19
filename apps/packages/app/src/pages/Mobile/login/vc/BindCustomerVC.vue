<template>
	<q-scroll-area class="bind-terminus-vc-scroll">
		<div
			class="bind-terminus-vc-page column justify-start items-center"
			style="position: relative"
		>
			<q-img
				class="bind-terminus-vc-page__image"
				:src="getRequireImage('login/create_terminus_name.svg')"
			/>
			<span class="bind-terminus-vc-page__desc login-ink-2">{{
				t(
					'Bind a verification credential (VC) via your social accounts to create an Olares ID'
				)
			}}</span>
			<div class="bind-terminus-vc-page__title text-ink-1 text-h6">
				{{ t('verify_using_gmail') }}
			</div>
			<terminus-v-c-card
				style="margin-top: 20px"
				:sub-text="
					t('please_use_a_gmail_email_address', {
						gmail: '@gmail.com'
					})
				"
				:type="TERMINUS_VC_TYPE.GOOGLE"
				@click="request(TERMINUS_VC_TYPE.GOOGLE)"
			/>
			<terminus-title-bar
				style="position: absolute; top: 0"
				:translate="true"
			/>
		</div>
	</q-scroll-area>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { getRequireImage } from '../../../../utils/imageUtils';
import { TERMINUS_VC_TYPE } from '../../../../utils/constants';
import { requestBindVC, readySubmitPresentation } from './BindVCBusiness';
import TerminusVCCard from '../../../../components/common/TerminusVCCard.vue';
import { useRouter } from 'vue-router';
import {
	BIND_STATUS,
	current_user_bind_status
} from '../../../../utils/terminusBindUtils';
import { useUserStore } from '../../../../stores/user';
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';

const $q = useQuasar();
const { t } = useI18n();
const router = useRouter();
const userStore = useUserStore();

async function request(request_type: TERMINUS_VC_TYPE) {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	$q.loading.show();
	await requestBindVC(request_type, null, {
		async onSuccess(data: any) {
			if (data) {
				const userBindStatus = current_user_bind_status();
				const terminusName = userStore.current_user?.name;
				const bindVCAndSelectedName =
					!terminusName && userBindStatus < BIND_STATUS.OS_NONE;
				if (bindVCAndSelectedName) {
					await readySubmitPresentation(data, null, {
						onSuccess() {
							router.replace({
								path: '/bind_vc_success'
							});
						},
						onFailure(message: string) {
							notifyFailed(message);
						}
					});
				} else {
					router.replace({ path: '/connectLoading' });
				}
			} else {
				notifyFailed(t('errors.bind_vc_error'));
			}
		},
		onFailure(message: string) {
			notifyFailed(message);
		}
	});
	$q.loading.hide();
}
</script>

<style lang="scss" scoped>
.bind-terminus-vc-scroll {
	width: 100%;
	height: 100%;

	.bind-terminus-vc-page {
		width: 100%;
		height: 100%;
		padding-left: 20px;
		padding-right: 20px;

		&__image {
			width: 160px;
			height: 160px;
			margin-top: 32px;
		}

		&__desc {
			width: 100%;
			margin-top: 20px;
		}

		&__title {
			margin-top: 32px;
			width: 100%;
		}
	}
}
</style>
