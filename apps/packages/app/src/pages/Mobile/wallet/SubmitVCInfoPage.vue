<template>
	<q-scroll-area class="connect-scroll">
		<div class="connect-root">
			<connect-header
				:title="t('notification')"
				:dappLogo="dappLogo"
				:dappUrl="dappUrl"
				:action="t('submit_vc_info')"
				:showDappInfo="true"
			/>
			<v-c-card-container
				v-if="cardItemRef"
				:show-details="true"
				class="vc-item"
				:item="cardItemRef"
			/>
			<div class="error text-body3" v-if="errorRef">
				{{ errorRef }}
			</div>
			<div
				class="row justify-center items-center"
				style="margin-top: 20px; width: 100%"
			>
				<confirm-button
					style="width: 45%"
					:btn-title="t('ok')"
					@onConfirm="approveAction"
				/>
			</div>
		</div>
	</q-scroll-area>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useApproval } from './approval';
import ConfirmButton from '../../../components/common/ConfirmButton.vue';
import ConnectHeader from './ConnectHeader.vue';
import { useBexStore } from '../../../stores/bex';
import VCCardContainer from '../items/VCCardContainer.vue';
import { createVCItem } from '../vc/vcutils';
import { useI18n } from 'vue-i18n';
import { VCCardInfo, convertVault2CardItem } from '../../../utils/vc';

const $q = useQuasar();
const $router = useRouter();
const store = useBexStore();
const { t } = useI18n();

const dappLogo = ref('');
const dappUrl = ref('');
const cardItemRef = ref();
const errorRef = ref();

store.controller.getApproval().then(async (approval) => {
	if (approval) {
		const origin = approval.data.session.origin;
		const icon = approval.data.session.icon;
		if (origin && typeof origin === 'string') {
			dappUrl.value = origin;
		}
		if (icon && typeof icon === 'string') {
			dappLogo.value = icon;
		}
		$q.loading.show();
		try {
			const vcInfo: VCCardInfo = approval.data.params.vcInfo as VCCardInfo;
			const vcItem = await createVCItem(vcInfo);
			if (!vcItem) {
				$q.loading.hide();
				throw new Error(t('errors.create_vc_error'));
			}
			cardItemRef.value = convertVault2CardItem(vcItem);
		} catch (e) {
			errorRef.value = e.message;
		} finally {
			$q.loading.hide();
		}
	}
});

const approveAction = async () => {
	const { resolveApproval } = useApproval($router);
	await resolveApproval({ code: cardItemRef.value ? 200 : 401 });
};
</script>

<style lang="scss" scoped>
.connect-scroll {
	width: 100%;
	height: 100%;

	.connect-root {
		width: 100%;
		height: 100%;

		.vc-item {
			width: calc(100% - 48px);
			margin: 24px;
		}

		.error {
			width: 100%;
			padding-left: 24px;
			padding-right: 24px;
			color: $negative;
		}
	}
}
</style>
