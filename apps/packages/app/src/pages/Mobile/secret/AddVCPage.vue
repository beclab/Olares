<template>
	<div class="add-vc">
		<back-title-bar :title="t('add_vc')" />
		<q-scroll-area class="add-vc__scroll">
			<div class="add-vc__layout">
				<terminus-v-c-card
					style="margin-top: 20px"
					:type="TERMINUS_VC_TYPE.FACEBOOK"
					v-if="!bindVCAndSelectedName"
					@click="request(TERMINUS_VC_TYPE.FACEBOOK)"
				/>

				<terminus-v-c-card
					style="margin-top: 20px"
					:type="TERMINUS_VC_TYPE.TWITTER"
					@click="request(TERMINUS_VC_TYPE.TWITTER)"
				/>
				<terminus-v-c-card
					style="margin-top: 20px"
					:type="TERMINUS_VC_TYPE.GOOGLE"
					@click="request(TERMINUS_VC_TYPE.GOOGLE)"
				/>
			</div>
		</q-scroll-area>
	</div>
</template>

<script setup lang="ts">
import BackTitleBar from '../../../components/common/BackTitleBar.vue';
import TerminusVCCard from '../../../components/common/TerminusVCCard.vue';

import { useQuasar } from 'quasar';
import { ref } from 'vue';
import {
	BIND_STATUS,
	current_user_bind_status
} from '../../../utils/terminusBindUtils';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { TERMINUS_VC_TYPE } from '../../../utils/constants';
import {
	readySubmitPresentation,
	requestBindVC
} from '../login/vc/BindVCBusiness';
import { useUserStore } from '../../../stores/user';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';

const $q = useQuasar();
const router = useRouter();
const bindVCAndSelectedName = ref<boolean>(false);
const { t } = useI18n();
const userStore = useUserStore();

const userBindStatus = current_user_bind_status();
const terminusName = userStore.current_user?.name;

bindVCAndSelectedName.value =
	!terminusName && userBindStatus < BIND_STATUS.OS_NONE;

async function request(request_type: TERMINUS_VC_TYPE) {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	$q.loading.show();
	await requestBindVC(request_type, null, {
		async onSuccess(data: any) {
			if (data) {
				if (bindVCAndSelectedName.value) {
					await readySubmitPresentation(data, null, {
						onSuccess() {
							router.back();
						},
						onFailure(message: string) {
							notifyFailed(message);
						}
					});
				} else {
					router.back();
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

<style scoped lang="scss">
.add-vc {
	height: 100%;
	width: 100%;

	&__scroll {
		height: 100%;
		width: 100%;
		padding-left: 24px;
		padding-right: 24px;
	}

	&__item {
		margin-top: 16px;
	}

	&__layout {
		width: 100%;
		height: 100%;
	}
}
</style>
