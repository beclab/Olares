<template>
	<div class="join-organization-root">
		<BindTerminusVCContent class="join-organization-vc-root" :has-btn="false">
			<template v-slot:desc>
				<div v-html="descRef" class="q-mb-md" />
			</template>
			<template v-slot:content>
				<terminus-v-c-card
					:type="TERMINUS_VC_TYPE.GOOGLE"
					@click="request(TERMINUS_VC_TYPE.GOOGLE)"
				/>
			</template>
		</BindTerminusVCContent>
		<terminus-title-bar
			style="position: absolute; top: 0"
			:translate="true"
			:hookBackAction="false"
		/>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import BindTerminusVCContent from './BindTerminusVCContent.vue';
import TerminusVCCard from '../../../../components/common/TerminusVCCard.vue';
import { useQuasar } from 'quasar';
import { ref, computed } from 'vue';
import { TERMINUS_VC_TYPE } from '../../../../utils/constants';
import { requestBindVC, readySubmitPresentation } from './BindVCBusiness';
import { useUserStore } from '../../../../stores/user';
import {
	BIND_STATUS,
	current_user_bind_status
} from '../../../../utils/terminusBindUtils';

import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';

const router = useRouter();

const route = useRoute();

const { t } = useI18n();

const $q = useQuasar();

const userStore = useUserStore();

const domainRef = ref(route.params.domain as string);

const descRef = computed(function () {
	return (
		`${t('bind_organization_vc_last_step_desc')}` +
		'<br><br>' +
		`<span class='text-blue-4'>${t(
			'bind_organization_vc_last_step_desc_more_note'
		)}</span>`
	);
});

async function request(request_type: TERMINUS_VC_TYPE) {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	$q.loading.show();
	await requestBindVC(request_type, domainRef.value, {
		async onSuccess(data: any) {
			if (data) {
				const userBindStatus = current_user_bind_status();
				const terminusName = userStore.current_user?.name;
				const bindVCAndSelectedName =
					!terminusName && userBindStatus < BIND_STATUS.OS_NONE;
				if (bindVCAndSelectedName) {
					await readySubmitPresentation(data, domainRef.value, {
						// data no use
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
.join-organization-root {
	width: 100%;
	height: 100%;
	position: relative;

	.join-organization-vc-root {
		width: 100%;
		height: 100%;
	}
}
</style>
