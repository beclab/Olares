<template>
	<div class="join-organization-root">
		<BindTerminusVCContent
			class="join-organization-vc-root"
			:has-btn="true"
			:btn-title="t('continue')"
			:btn-status="btnStatus"
			@onConfirm="queryDomainRule"
			descClasses="desc-more"
		>
			<template v-slot:desc>
				<div v-html="descRef" style="text-align: left" />
			</template>
			<template v-slot:content>
				<terminus-edit
					v-model="domainRef"
					:label="t('Olares ID')"
					class="q-mt-md"
					@update:model-value="domainUpdate"
					input-type="email"
				/>

				<terminus-edit
					v-model="passwordRef"
					:label="t('password')"
					:showPasswordImg="true"
					class="q-mt-lg"
					@update:model-value="domainUpdate"
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
import { ref, computed } from 'vue';

import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';
import { joinOrganization } from './BindVCBusiness';
import { useQuasar } from 'quasar';

const router = useRouter();

const { t } = useI18n();

const $q = useQuasar();

const descRef = computed(function () {
	return `${t('enter_domain_name_organization_want_to_join', {
		org: "<span class='text-blue-4'>member@organization.com</span>"
	})}`;
});

const domainRef = ref('');

const passwordRef = ref('');

const btnStatus = ref(ConfirmButtonStatus.disable);

const domainUpdate = () => {
	btnStatus.value =
		domainRef.value.length > 0 && passwordRef.value.length > 0
			? ConfirmButtonStatus.normal
			: ConfirmButtonStatus.disable;
};

async function queryDomainRule() {
	if (!domainRef.value || domainRef.value.split('@').length !== 2) {
		notifyFailed(
			t('Organization member names are not available, for example:') +
				"<br><span class='text-blue-4'>member@organization.com</span>"
		);
		return;
	}
	$q.loading.show();

	joinOrganization(domainRef.value, passwordRef.value, {
		onSuccess: async () => {
			$q.loading.hide();
			router.replace({
				path: '/bind_vc_success'
			});
		},
		onFailure: (message: string) => {
			$q.loading.hide();
			notifyFailed(message);
		}
	});
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
