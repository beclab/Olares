<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('prompts.replace')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="small"
		@onSubmit="(event) => store.showConfirm(event, 'replace')"
		@onCancel="onHide"
		@onHide="onHide"
	>
		<div :style="{ textAlign: isMobile ? 'center' : 'left' }">
			{{ t('prompts.replaceMessage') }}
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useQuasar } from 'quasar';
import { useDataStore } from '../../../stores/data';

import { useI18n } from 'vue-i18n';

const $q = useQuasar();
const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);
const store = useDataStore();

const { t } = useI18n();
const onHide = () => {
	store.closeHovers;
};
</script>
