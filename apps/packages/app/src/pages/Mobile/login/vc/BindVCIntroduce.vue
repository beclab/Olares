<template>
	<div class="account-create-terminus">
		<q-img
			class="account-create-terminus__img"
			:src="getRequireImage('login/mobile_login_background.svg')"
		/>
		<terminus-wizard-view
			:btn-status="btnStatus"
			:btn-title="btnTitleRef"
			@on-confirm="enterVC"
		>
			<template v-slot:content>
				<div class="text-h5 account-create-terminus__title q-mb-md">
					{{ title }}
				</div>
				<div
					class="text-body1 account-create-terminus__content"
					v-html="reminderInfoRef"
				></div>
			</template>
		</terminus-wizard-view>
	</div>
</template>

<script setup lang="ts">
import TerminusWizardView from '../../../../components/common/TerminusWizardView.vue';
import { getRequireImage } from '../../../../utils/imageUtils';
import { ref, onMounted, onUnmounted } from 'vue';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import { useI18n } from 'vue-i18n';
import { useRouter, useRoute } from 'vue-router';
const { t } = useI18n();

const $router = useRouter();

const Route = useRoute();

const isPersonal = Route.params.isPersonal
	? Route.params.isPersonal == 'true'
		? true
		: false
	: false;

const reminderInfoRef = ref(
	isPersonal
		? t('create_olares_id_info')
		: t('create_organzation_olares_id_info')
);

const title = ref(
	isPersonal ? t('about_creating_olares_id') : t('about_organzation_olares_id')
);

const btnStatus = ref(ConfirmButtonStatus.disable);
const btnTitle = t('i_have_read');
let leftTime = 1;
const btnTitleRef = ref(btnTitle + ' ' + leftTime + 's');

let left_time_interval: any = null;

onMounted(() => {
	left_time_interval = setInterval(async () => {
		leftTime = leftTime - 1;
		if (leftTime > 0) {
			btnTitleRef.value = btnTitle + ' ' + leftTime + 's';
		} else {
			btnTitleRef.value = btnTitle;
			btnStatus.value = ConfirmButtonStatus.normal;
		}
	}, 1 * 1000);
});

onUnmounted(() => {
	if (left_time_interval) {
		clearInterval(left_time_interval);
	}
});

const enterVC = async () => {
	if (isPersonal) {
		$router.replace({ path: '/bind_customer_vc' });
		return;
	}
	$router.replace({ path: '/bind_org_vc' });
};
</script>

<style scoped lang="scss">
.account-create-terminus {
	width: 100%;
	height: 100%;
	position: relative;

	&__img {
		width: 276px;
		height: 322px;
		position: absolute;
		left: 0;
	}

	&__title {
		// padding-top: 115px;
		margin-top: 115px;
	}

	&__content {
		white-space: pre-wrap;
		word-wrap: break-word;
		// padding-bottom: 40px;
	}
}
</style>
