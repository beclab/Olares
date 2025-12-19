<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('main.split_view')"
		@onSubmit="onOK"
		:cancel="t('base.cancel')"
		:ok="t('base.confirm')"
		:okDisabled="filter.splitview === splitView"
	>
		<div class="text-ink-2 text-body3 q-mb-xs">
			{{ t('main.categorizes_content_based_on') }}
		</div>

		<bt-select
			v-model="splitView"
			:options="selectOptions"
			:border="true"
			color="text-orange-default"
		/>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import BtSelect from '../../base/BtSelect.vue';
import { useFilterStore } from '../../../stores/rss-filter';
import { FilterInfo, SPLIT_TYPE } from '../../../utils/rss-types';
import { PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const filterStore = useFilterStore();
const customRef = ref();
const { t } = useI18n();

const props = defineProps({
	filter: {
		type: Object as PropType<FilterInfo>,
		required: true
	}
});

const splitView = ref(props.filter.splitview);

const selectOptions = [
	{
		label: t('main.location'),
		value: SPLIT_TYPE.LOCATION,
		enable: true
	},
	{
		label: t('main.seen_unseen'),
		value: SPLIT_TYPE.SEEN,
		enable: true
	},
	{
		label: t('main.none'),
		value: SPLIT_TYPE.NONE,
		enable: true
	}
];

const onOK = () => {
	customRef.value.onDialogOK(splitView.value);
};
</script>

<style scoped lang="scss"></style>
