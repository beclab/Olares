<template>
	<div class="row justify-center items-center no-wrap">
		<q-input
			class="page-input text-body3"
			v-model="pdfStore.pageNum"
			@update:model-value="pageUpdate"
			borderless
			debounce="500"
			input-class="text-ink-2 text-body3"
			dense
			input-style="height: 24px"
			no-error-icon
			placeholder="https://"
		/>
		<span
			class="text-center q-ml-xs text-body3 text-ink-2"
			style="text-align: center; white-space: nowrap"
		>
			/ {{ pdfStore.numPages }}
		</span>

		<div class="title-line bg-separator" />

		<q-btn
			class="btn-size-sm btn-no-text btn-no-border no-padding"
			color="ink-2"
			outline
			no-caps
			icon="sym_r_remove"
			@click="pdfStore.pageZoomOut()"
		>
			<bt-tooltip :label="t('pdf.zoom_out')" />
		</q-btn>

		<div class="scale-input text-center text-body3 text-ink-2 q-py-xs">
			{{ formattedZoom(pdfStore.scale) }}
		</div>

		<q-btn
			class="btn-size-sm btn-no-text btn-no-border no-padding"
			color="ink-2"
			outline
			no-caps
			icon="sym_r_add"
			@click="pdfStore.pageZoomIn()"
		>
			<bt-tooltip :label="t('pdf.zoom_in')" />
		</q-btn>

		<div class="title-line bg-separator" />

		<q-btn
			class="btn-size-sm btn-no-text btn-no-border no-padding"
			color="ink-2"
			outline
			no-caps
			icon="sym_r_rotate_90_degrees_cw"
			@click="pdfStore.pageRotate()"
		>
			<bt-tooltip :label="t('pdf.rotate_clockwise')" />
		</q-btn>

		<q-btn
			class="btn-size-sm btn-no-text btn-no-border no-padding"
			color="ink-2"
			outline
			no-caps
			icon="sym_r_rotate_90_degrees_ccw"
			@click="pdfStore.pageCounterRotate()"
		>
			<bt-tooltip :label="t('pdf.rotate_counterclockwise')" />
		</q-btn>
	</div>
</template>

<script setup lang="ts">
import BtTooltip from '../../../../components/base/BtTooltip.vue';
import { usePDfStore } from '../../../../stores/pdf';
import { useI18n } from 'vue-i18n';

const pdfStore = usePDfStore();
const { t } = useI18n();
const pageUpdate = () => {
	pdfStore.skipPage(pdfStore.pageNum);
};

function formattedZoom(pdfScale: number) {
	return (pdfScale * 100).toFixed(0) + '%';
}
</script>

<style scoped lang="scss">
.page-input {
	border: 1px solid $separator;
	min-width: 24px;
	max-width: 28px;
	width: auto;
	height: 24px;
	border-radius: 4px;
	padding-left: 4px;
	padding-right: 4px;
	text-align: center;
}

.no-padding {
	padding: 0 !important;
}

.scale-input {
	border: 1px solid $separator;
	width: 46px;
	height: 24px;
	border-radius: 4px;
}

.title-line {
	margin-left: 16px;
	margin-right: 16px;
	height: 20px;
	width: 1px;
}
::v-deep(.q-field__control) {
	height: 24px !important;
}
</style>
