<template>
	<div class="vc-manager-root">
		<terminus-title-bar :title="t('vc_management')" />
		<terminus-scroll-area class="vc-scroll">
			<template v-slot:content>
				<q-list>
					<q-item class="q-mt-md" v-if="!isBex">
						<div class="column" style="width: 100%">
							<template v-for="(item, index) in VCCardItemList" :key="item.id">
								<VCCard
									v-if="!!item.verifiable_credential"
									:style="index > 0 ? 'margin-top: 20px' : ''"
									:data="item"
									:id="item.id"
								/>
							</template>
							<terminus-v-c-card
								:style="VCCardItemList.length > 0 ? 'margin-top: 20px' : ''"
								:type="TERMINUS_VC_TYPE.FACEBOOK"
								v-if="showFacebook"
								@click="request(TERMINUS_VC_TYPE.FACEBOOK)"
							/>

							<terminus-v-c-card
								style="margin-top: 20px"
								:type="TERMINUS_VC_TYPE.TWITTER"
								v-if="showTwitter"
								@click="request(TERMINUS_VC_TYPE.TWITTER)"
							/>
							<terminus-v-c-card
								style="margin-top: 20px"
								:type="TERMINUS_VC_TYPE.GOOGLE"
								v-if="showGoogle"
								@click="request(TERMINUS_VC_TYPE.GOOGLE)"
							/>
						</div>
					</q-item>
				</q-list>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import TerminusVCCard from '../../../components/common/TerminusVCCard.vue';
import VCCard from '../vc/card/VCCard.vue';
import { getVCCardItemList } from '../items/VCItemUtil';
import { useMenuStore } from '../../../stores/menu';
import { debounce } from '@didvault/sdk/src/core';
import { TERMINUS_VC_TYPE } from '../../../utils/constants';
import { useRouter } from 'vue-router';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import { busOn, busOff } from '../../../utils/bus';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';

import { useQuasar } from 'quasar';
import {
	BIND_STATUS,
	current_user_bind_status
} from '../../../utils/terminusBindUtils';
import { useI18n } from 'vue-i18n';
import {
	readySubmitPresentation,
	requestBindVC
} from '../login/vc/BindVCBusiness';
import { useUserStore } from '../../../stores/user';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { VCCardItem } from '../../../utils/vc';

const $q = useQuasar();
const router = useRouter();
const bindVCAndSelectedName = ref<boolean>(false);
const { t } = useI18n();
const userStore = useUserStore();
const isBex = ref(process.env.IS_BEX);

const userBindStatus = current_user_bind_status();
const terminusName = userStore.current_user?.name;

bindVCAndSelectedName.value =
	!terminusName && userBindStatus < BIND_STATUS.OS_NONE;

async function request(request_type: TERMINUS_VC_TYPE) {
	$q.loading.show();
	await requestBindVC(request_type, null, {
		async onSuccess(data: any) {
			$q.loading.hide();
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
				notifyFailed(t('bind_vc_error'));
			}
		},
		onFailure(message: string) {
			$q.loading.hide();
			notifyFailed(message);
		}
	});
	$q.loading.hide();
}

const VCCardItemList = ref<VCCardItem[]>([]);
const menuStore = useMenuStore();

const showFacebook = ref(true);
const showTwitter = ref(true);
const showGoogle = ref(true);

onMounted(() => {
	stateUpdate();
	busOn('appSubscribe', stateUpdate);
	menuStore.$subscribe(() => {
		updateItems();
	});
});

let updateItems = debounce(() => {
	stateUpdate();
}, 50);

onUnmounted(() => {
	busOff('appSubscribe', stateUpdate);
});

function stateUpdate() {
	VCCardItemList.value = getVCCardItemList();

	VCCardItemList.value.forEach((item) => {
		if (item.type === TERMINUS_VC_TYPE.TWITTER) {
			showTwitter.value = false;
		} else if (item.type === TERMINUS_VC_TYPE.FACEBOOK) {
			showFacebook.value = false;
		} else if (item.type === TERMINUS_VC_TYPE.GOOGLE) {
			showGoogle.value = false;
		}
	});
}
</script>

<style scoped lang="scss">
.vc-manager-root {
	width: 100%;
	height: 100%;
}

.vc-scroll {
	width: 100%;
	height: calc(100% - 56px);
}
</style>
