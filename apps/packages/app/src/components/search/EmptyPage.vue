<template>
	<div class="text-grey-7 text-center q-pa-md flex flex-center">
		<div>
			<div style="font-size: 30vh">404</div>

			<div class="text-h2" style="opacity: 0.4">{{ t('page_404') }}</div>

			<q-btn
				class="q-mt-xl"
				color="white"
				text-color="blue"
				label="Go Home"
				no-caps
				@click="goBack"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

const emits = defineEmits(['openCommand']);

const { t } = useI18n();

const goBack = () => {
	emits('openCommand', '');
};
</script>
