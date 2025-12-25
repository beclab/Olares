<template>
	<PageTitleComponent :title="titleMap[step]" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div v-if="step == AwsAddStep.start" class="q-mt-lg">
			<!-- <terminus-edit
				v-model="accessKeyID"
				:label="t('access_key_id')"
				:show-password-img="false"
				style="width: 100%"
			/>
			<terminus-edit
				v-model="accessKeySecret"
				class="q-mt-lg"
				:label="t('access_key_secret')"
				:show-password-img="false"
				style="width: 100%"
			/>
			<terminus-edit
				v-model="endpoint"
				class="q-mt-lg"
				:label="t('endpoint')"
				:show-password-img="false"
				style="width: 100%"
			/>
			<terminus-edit
				v-model="bucket"
				class="q-mt-lg"
				:label="t('bucket')"
				:show-password-img="false"
				:hint-text="t('optional')"
				style="width: 100%"
			/> -->
			<IntegrationAddInputs
				ref="integrationAddInputs"
				:account-type="accountType"
				v-model:button-status="enableCreate"
			/>
		</div>
		<div v-else-if="step == AwsAddStep.display">
			<div
				class="text-body3 text-ink-2 q-mt-lg"
				:class="{
					'text-body3': !deviceStore.isMobile,
					'text-body3-m': deviceStore.isMobile
				}"
			>
				{{
					t(
						'Congratulations, you have successfully mounted the object storage service!'
					)
				}}
			</div>
			<div class="row item-center justify-between item-content q-px-md q-mt-md">
				<div
					class="text-subtitle2 text-ink-2 row items-center"
					style="height: 100%"
				>
					{{ t('Object Storage') }}
				</div>
				<div class="row items-center justify-end">
					<q-img
						:src="getRequireImage(`integration/${accountInfo.icon}`)"
						width="32px"
						height="32px"
					/>
					<div class="text-subtitle2 text-ink-1 q-ml-sm">
						{{ accountInfo.name }}
					</div>
				</div>
			</div>
		</div>
		<div class="bottom-buttons">
			<confirm-button
				:btn-title="stepMap[step]"
				bgClasses="bg-blue-default"
				bgDisabledClasses="bg-blue-2"
				textClasses="text-white"
				:btn-status="btnStatusRef"
				@onConfirm="onConfirm"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import integrationService from 'src/services/integration/index';
import { useIntegrationStore } from 'src/stores/settings/integration';
import { notifyFailed } from 'src/utils/settings/btNotify';
import ConfirmButton from 'src/components/settings/ConfirmButton.vue';
import { ConfirmButtonStatus } from 'src/utils/constants';
import { getRequireImage } from 'src/utils/settings/helper';
import { useDeviceStore } from 'src/stores/settings/device';
import { AccountType } from '@bytetrade/core';
import IntegrationAddInputs from '../components/IntegrationAddInputs.vue';

enum AwsAddStep {
	start = 1,
	display
}

const { t } = useI18n();

const deviceStore = useDeviceStore();

const route = useRoute();

const stepMap: Record<AwsAddStep, string> = {
	[AwsAddStep.start]: t('next'),
	[AwsAddStep.display]: t('confirm')
};

const titleMap: Record<AwsAddStep, string> = {
	[AwsAddStep.start]: t('mount_network_drive'),
	[AwsAddStep.display]: t('add_account')
};

const accountType = ref(route.query.accountType as AccountType);

const accountInfo = ref(
	integrationService.supportAuthList.find((e) => e.type == accountType.value)!
		.detail
);

const $q = useQuasar();

const enableCreate = ref(false);

const integrationStore = useIntegrationStore();
const router = useRouter();

const btnStatusRef = computed(() => {
	return step.value == AwsAddStep.start
		? enableCreate.value
			? ConfirmButtonStatus.normal
			: ConfirmButtonStatus.disable
		: ConfirmButtonStatus.normal;
});

const integrationAddInputs = ref();

const onConfirm = async () => {
	if (step.value == AwsAddStep.start) {
		$q.loading.show();
		const inputs = integrationAddInputs.value.allAccountValues();
		try {
			await integrationStore.createAccount(inputs);
			step.value = AwsAddStep.display;
			$q.loading.hide();
		} catch (error) {
			$q.loading.hide();
			notifyFailed(error.message);
		}
	} else {
		router.go(-2);
	}
};

const step = ref(AwsAddStep.start);
</script>

<style scoped lang="scss">
.bottom-buttons {
	position: absolute;
	left: 20px;
	bottom: 30px;
	width: calc(100% - 40px);
	// height: 40px;
}

.item-content {
	height: 56px;
	width: 100%;
	border: 1px solid $separator;
	border-radius: 12px;
}
</style>
