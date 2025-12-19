<template>
	<div class="row justify-end items-center q-mr-sm no-wrap">
		<slot />
		<q-btn
			v-if="!configStore.rightDrawerOpen && rightDrawer"
			class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
			icon="sym_r_right_panel_open"
			color="ink-2"
			:disable="!readerStore.hoverEntry"
			outline
			no-caps
			@click="configStore.setRightDrawerOpen(true)"
		>
			<bt-tooltip :label="t('base.show_panel')" />
		</q-btn>
	</div>
</template>

<script setup lang="ts">
import { useConfigStore } from 'src/stores/rss-config';
import { useReaderStore } from 'src/stores/rss-reader';
import BtTooltip from './BtTooltip.vue';
import { useI18n } from 'vue-i18n';

const configStore = useConfigStore();
const readerStore = useReaderStore();
const { t } = useI18n();

defineProps({
	rightDrawer: {
		type: Boolean,
		default: true
	}
});
</script>

<style lang="scss"></style>
