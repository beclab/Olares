<template>
	<bt-popup style="width: 240px" padding="8px 12px">
		<div class="column">
			<template v-for="item in filterStore.customList" :key="item.id">
				<bt-check-box
					:model-value="
						viewList.filter((view) => view.id === item.id).length > 0
					"
					:label="item.name"
					@update:model-value="
						(value: any, _: Event) => onSelected(item, value)
					"
				/>
			</template>

			<bt-popup-item
				:title="t('main.create_a_new_view')"
				icon="sym_r_add_circle"
				:selected="true"
				:selected-icon="false"
				@on-item-click="createTag"
			>
				<template v-slot:after="{ hover }">
					<create-view
						max-width="100px"
						:selected="hover"
						class="q-ml-sm"
						:name="viewName"
						:edit="false"
					/>
				</template>
			</bt-popup-item>
		</div>
	</bt-popup>
</template>

<script lang="ts" setup>
import FilterEditDialog from './dialog/FilterEditDialog.vue';
import { useFilterStore } from '../../stores/rss-filter';
import { FilterInfo } from '../../utils/rss-types';
import BtPopupItem from '../base/BtPopupItem.vue';
import BtCheckBox from '../rss/BtCheckBox.vue';
import { computed, PropType, ref } from 'vue';
import BtPopup from '../base/BtPopup.vue';
import CreateView from './CreateView.vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';

const props = defineProps({
	data: {
		type: Object as PropType<any>,
		require: true
	},
	type: {
		type: String,
		required: true
	}
});

const { t } = useI18n();
const viewName = ref();
const filterStore = useFilterStore();
const $q = useQuasar();

const viewList = computed(() => {
	if (!props.data) {
		return [];
	}
	let list = [];
	if (props.type === 'feed_id') {
		const value = filterStore.feedMap.get(props.data.id);
		if (value && value.size > 0) {
			list = Array.from(value);
		}
	} else {
		const value = filterStore.labelMap.get(props.data.id);
		if (value && value.size > 0) {
			list = Array.from(value);
		}
	}

	console.log('====>ssss', filterStore.feedMap);
	console.log('====>ssss', filterStore.labelMap);
	console.log('====>ssss', list);
	return list;
});

const onSelected = (filter: FilterInfo, selected: boolean) => {
	if (!props.data) {
		return;
	}
	console.log('====>ssss333', selected);
	if (selected) {
		filterStore.addToQuery(filter, props.type, props.data.id);
	} else {
		filterStore.removeFromQuery(filter, props.type, props.data.id);
	}
};

const createTag = async () => {
	if (!props.data) {
		return;
	}

	$q.dialog({
		component: FilterEditDialog,
		componentProps: {
			title:
				props.type === 'feed_id'
					? t('dialog.create_a_new_view_from_feed')
					: t('dialog.create_a_new_view_from_tag')
		}
	}).onOk((filter) => {
		if (filter) {
			filterStore.addToQuery(filter, props.type, props.data.id);
		}
	});
};
</script>

<style lang="scss"></style>
