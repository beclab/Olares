<template>
	<div class="join-organization-root">
		<BindTerminusVCContent
			class="join-organization-vc-root"
			:has-btn="true"
			:btn-title="t('continue')"
			:btn-status="btnStatus"
			@onConfirm="queryDomainRule"
		>
			<template v-slot:desc>
				<div v-html="descRef" />
			</template>
			<template v-slot:content>
				<terminus-edit
					v-model="domainRef"
					:label="t('Domain name')"
					class="q-mt-lg"
					@update:model-value="domainUpdate"
					input-type="email"
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
import { useRouter } from 'vue-router';
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import BindTerminusVCContent from './BindTerminusVCContent.vue';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import TerminusEdit from '../../../../components/common/TerminusEdit.vue';
import { useCloudStore } from '../../../../stores/cloud';
import { ref, computed } from 'vue';

import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';

const router = useRouter();

const cloudStore = useCloudStore();

const { t } = useI18n();

const descRef = computed(function () {
	return (
		`${t('enter_domain_name_organization_want_to_join')}` +
		"<span class='text-blue-4'>g.com</span>"
	);
});

const domainRef = ref('');

const btnStatus = ref(ConfirmButtonStatus.disable);

const domainUpdate = () => {
	btnStatus.value =
		domainRef.value.length > 0
			? ConfirmButtonStatus.normal
			: ConfirmButtonStatus.disable;
};

async function queryDomainRule() {
	try {
		const result = await cloudStore.getDomainRule(domainRef.value);
		if (result) {
			// rule.value = result;
			router.replace({
				path: '/OrganizationBindVC/' + domainRef.value
			});
		} else {
			notifyFailed(
				t(
					'the_current_domain_name_is_unavailable_please_contact_the_administrator_to_obtain_a_valid_domain_name'
				)
			);
		}
	} catch (error) {
		if (error && error.message) {
			notifyFailed(error.message);
		}
	}
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
