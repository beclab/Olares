<template>
	<bt-popup style="width: 240px" padding="8px 12px">
		<div class="row">
			<template v-for="item in entryLabels" :key="item.id">
				<create-view
					class="q-mr-xs q-my-xs"
					:name="item.name"
					:edit="true"
					@on-remove-click="onSelected(item, false)"
				/>
			</template>
			<q-input
				dense
				borderless
				:class="entryLabels.length === 0 ? 'full-width' : ''"
				:placeholder="t('create_tag_hide_text')"
				class="q-mt-sm q-mx-sm"
				style="height: 16px"
				input-class="text-ink-2 text-body3"
				input-style="padding-left: 0px; height: 16px"
				v-model="labelName"
				@keyup.enter="createTag"
			/>
		</div>

		<q-separator class="q-mt-sm bg-separator" />

		<div class="column">
			<template v-for="item in rssStore.labels" :key="item.id">
				<bt-check-box
					:model-value="
						entry?.id && item.entries ? item.entries.includes(entry?.id) : false
					"
					:label="item.name"
					@update:model-value="
						(value: any, _: Event) => onSelected(item, value)
					"
				/>
			</template>

			<bt-popup-item
				v-if="labelName"
				:title="t('main.create_tag')"
				:selected="true"
				@on-item-click="createTag"
			>
				<template v-slot:after="{ hover }">
					<create-view
						max-width="100px"
						:selected="hover"
						class="q-ml-sm"
						:data="labelName"
						:edit="false"
					/>
				</template>
			</bt-popup-item>
		</div>
	</bt-popup>
</template>

<script lang="ts" setup>
import { useRssStore } from '../../stores/rss';
import { PropType, ref, watch } from 'vue';
import { Entry, Label } from '../../utils/rss-types';
import BtCheckBox from '../rss/BtCheckBox.vue';
import BtPopupItem from 'src/components/base/BtPopupItem.vue';
import BtPopup from 'src/components/base/BtPopup.vue';
import CreateView from './CreateView.vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	entry: {
		type: Object as PropType<Entry>,
		require: true
	}
});

const { t } = useI18n();
const rssStore = useRssStore();
const labelName = ref();
const entryLabels = ref(rssStore.getEntryLabels(props.entry));

watch(
	() => rssStore.labels,
	() => {
		entryLabels.value = rssStore.getEntryLabels(props.entry);
	},
	{
		deep: true
	}
);

const onSelected = (label: Label, selected: boolean) => {
	if (!props.entry) {
		return;
	}
	if (selected) {
		rssStore.setLabelOnEntry(label.id, props.entry?.id);
	} else {
		rssStore.removeLabelOnEntry(label.id, props.entry?.id);
	}
};

const createTag = async () => {
	if (!labelName.value) {
		return;
	}
	const label = await rssStore.addLabel(labelName.value);
	if (props.entry) {
		await rssStore.setLabelOnEntry(label.id, props.entry?.id);
		labelName.value = '';
	}
};
</script>

<style lang="scss"></style>
