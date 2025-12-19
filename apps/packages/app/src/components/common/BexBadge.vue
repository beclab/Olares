<template>
	<div>
		<div class="text-h5 text-ink-1">{{ $t('bex.badge') }}</div>
		<div class="q-mt-md">
			<div class="badge-card-wrapper">
				<div class="row items-center justify-between q-py-lg q-pl-lg q-pr-sm">
					<div class="text-subtitle2 text-ink-1">{{ t('enable_badges') }}</div>
					<bt-switch
						class="custom-toggle-wrapper"
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="isExtensionBadge"
						@update:model-value="setExtensionBadge"
					/>
				</div>
				<template v-if="isExtensionBadge">
					<q-separator color="separator-2" />
					<div
						class="q-py-lg q-pl-lg q-pr-sm column no-wrap flex-gap-y-lg content"
					>
						<div class="row items-center justify-between q-py-xs">
							<div class="text-subtitle2 text-ink-1">
								{{ t('enable_approval_badge') }}
							</div>

							<bt-switch
								class="custom-toggle-wrapper"
								size="sm"
								truthy-track-color="light-blue-default"
								v-model="approvalBadgeEnableRef"
								@update:model-value="setApprovalBadgeEnable"
							/>
						</div>
						<div class="row items-center justify-between q-py-xs">
							<div class="text-subtitle2 text-ink-1">
								{{ t('enable_autofill_badge') }}
							</div>

							<bt-switch
								class="custom-toggle-wrapper"
								size="sm"
								truthy-track-color="light-blue-default"
								v-model="autofillBadgeEnableRef"
								@update:model-value="setAutofillBadgeEnable"
							/>
						</div>
						<div class="row items-center justify-between q-py-xs">
							<div class="text-subtitle2 text-ink-1">
								{{ t('enable_rss_badge') }}
							</div>

							<bt-switch
								class="custom-toggle-wrapper"
								size="sm"
								truthy-track-color="light-blue-default"
								v-model="rssBadgeEnableRef"
								@update:model-value="setRssBadgeEnable"
							/>
						</div>
					</div>
				</template>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useBexStore } from '../../stores/bex';
import { onMounted, ref } from 'vue';
import { app } from '../../globals';
import TerminusSettingsModuleItem from '../../components/common/TerminusSettingsModuleItem.vue';

const { t } = useI18n();
const bexStore = useBexStore();
const isExtensionBadge = ref(true);

const isBex = ref(process.env.IS_BEX || process.env.DEV_PLATFORM_BEX);

const autofillBadgeEnableRef = ref(true);
const rssBadgeEnableRef = ref(true);
const approvalBadgeEnableRef = ref(true);

const setAutofillBadgeEnable = async (enable: boolean) => {
	return bexStore.controller.setAutofillBadgeEnable(enable);
};

const setRssBadgeEnable = async (enable: boolean) => {
	return bexStore.controller.setRssBadgeEnable(enable);
};

const setApprovalBadgeEnable = async (enable: boolean) => {
	return bexStore.controller.setApprovalBadgeEnable(enable);
};

const setExtensionBadge = async (enable) => {
	await app.setSettings({ extensionBadge: enable });
};

onMounted(async () => {
	if (isBex.value) {
		isExtensionBadge.value = app.settings.extensionBadge;
		autofillBadgeEnableRef.value =
			await bexStore.controller.getAutofillBadgeEnable();
		rssBadgeEnableRef.value = await bexStore.controller.getRssBadgeEnable();
		approvalBadgeEnableRef.value =
			await bexStore.controller.getApprovalBadgeEnable();
	}
});
</script>

<style lang="scss" scoped>
.badge-card-wrapper {
	border: 1px solid $separator-2;
	border-radius: 12px;
}
.custom-toggle-wrapper {
	::v-deep(.q-toggle__inner--truthy .q-toggle__thumb:after) {
		background-color: $ink-on-brand !important;
	}
}
</style>
