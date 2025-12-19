<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('activation_third_party_domain')"
		:skip="false"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		@onSubmit="onOKClick"
	>
		<div class="text-ink-1 text-body2">
			{{
				t(
					'in_order_to_apply_third_party_domain_to_this_app_you_need_to_add_a_dns_cname_record_domain_to_your_domain_target',
					{
						domain: thirdPartyDomain,
						target: cnameTarget
					}
				)
			}}
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from 'src/stores/settings/device';

defineProps({
	thirdPartyDomain: {
		type: String,
		require: true
	},
	cnameTarget: {
		type: String,
		require: true
	}
});
const { t } = useI18n();
const deviceStore = useDeviceStore();

const CustomRef = ref();
async function onOKClick() {
	CustomRef.value.onDialogOK();
}
</script>
