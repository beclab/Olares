<template>
	<div class="row items-center no-wrap full-width">
		<q-btn
			padding="0px"
			flat
			:disable="disable"
			:loading="loading"
			icon="sym_r_sell"
			size="9px"
			color="ink-2"
			class="q-mr-xs"
			@click.stop
		>
			<bt-tooltip
				sty
				:label="entry?.exist_entry_id ? $t('add_tag') : $t('add_tag_info')"
				anchor="top middle"
				self="bottom middle"
				style="max-width: 240px; white-space: wrap; text-align: left"
			/>
			<bt-popup style="width: 240px; max-height: 50%" padding="8px 12px">
				<div class="row items-center flex-gap-xs">
					<create-view
						v-for="item in entryLabels"
						:key="item.id"
						:name="item.name"
						:edit="true"
						@on-remove-click="onSelected(item, false)"
					/>
				</div>
				<q-input
					dense
					borderless
					:placeholder="t('create_tag_hide_text')"
					class="q-mt-sm q-mx-sm full-width"
					style="height: 16px"
					input-class="text-ink-2 text-body3"
					input-style="padding-left: 0px; height: 16px"
					v-model="labelName"
					@keyup.enter="createTag"
					@focus.stop.prevent
				/>

				<q-separator class="q-mt-sm bg-separator" />

				<div class="column">
					<template v-for="item in labels" :key="item.exist_entry_id">
						<bt-check-box
							:model-value="
								entry?.exist_entry_id && item.entries
									? item.entries.includes(entry?.exist_entry_id)
									: false
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
				<q-inner-loading :showing="loading"> </q-inner-loading>
			</bt-popup>
		</q-btn>

		<div class="ellipsis no-wrap" style="flex: 1">
			<create-view
				v-for="item in entryLabels"
				:key="item.id"
				:name="item.name"
				:selected="false"
				class="q-mr-xs"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useRssStore } from '../../stores/rss';
import { computed, onMounted, ref, watch } from 'vue';
import { Label } from '../../utils/rss-types';
import BtCheckBox from '../rss/BtCheckBox2.vue';
import BtPopupItem from 'src/components/base/BtPopupItem.vue';
import BtPopup from 'src/components/base/BtPopup.vue';
import CreateView from './CreateView.vue';
import { useI18n } from 'vue-i18n';
import {
	createLabel,
	setLabelOnEntry,
	removeLabelOnEntry,
	syncLabels,
	getLabel
} from 'src/api/wise';
import { CollectEntry } from 'src/types/commonApi';
import axios, { CancelTokenSource } from 'axios';
import BtTooltip from '../base/BtTooltip.vue';

const CancelToken = axios.CancelToken;
let CollectSearchCancelTokenSource: undefined | CancelTokenSource = undefined;
interface Props {
	entry: CollectEntry;
	disabled?: boolean;
}
const props = defineProps<Props>();

const { t } = useI18n();
const rssStore = useRssStore();
const labelName = ref();
const entryLabels = ref();
const labels = ref<Label[]>([]);
const loading = ref(false);

const disable = computed(() => {
	return props.disabled || !props.entry?.exist_entry_id || loading.value;
});

const onSelected = async (label: Label, selected: boolean) => {
	if (!props.entry?.exist_entry_id) {
		return;
	}
	loading.value = true;
	try {
		const entry_id = props.entry.exist_entry_id;
		if (!entry_id) {
			return;
		}
		if (selected) {
			await setLabelOnEntry(label.id, entry_id);
			addEntryLabelsLocal(label, entry_id);
		} else {
			await removeLabelOnEntry(label.id, entry_id);
			deleteEntryLabelsLocal(label, entry_id);
		}
		syncLabelInWise();
	} catch (error) {
		console.error('Error setting label on entry:', error);
	}
	loading.value = false;
};

const createTag = async () => {
	if (!labelName.value) {
		return;
	}
	loading.value = true;
	try {
		const label = await createLabel(labelName.value);
		labels.value.push(label);
		if (props.entry.exist_entry_id) {
			labelName.value = '';
		}
		syncLabelInWise();
	} catch (error) {
		console.error('Error creating label:', error);
	}
	loading.value = false;
};

const getWiseLabels = async () => {
	labels.value = await syncLabels(0);
};

const addEntryLabelsLocal = (label: Label, entry_id: string) => {
	const index = labels.value.findIndex((item: Label) => item.id === label.id);
	if (index > -1) {
		labels.value[index].entries?.push(entry_id);
	}

	entryLabels.value.push(label);
};

const deleteEntryLabelsLocal = (label: Label, entry_id: string) => {
	const index = labels.value.findIndex((item: Label) => item.id === label.id);
	if (index > -1) {
		labels.value[index].entries = labels.value[index].entries?.filter(
			(item) => item !== entry_id
		);
	}

	entryLabels.value = entryLabels.value.filter(
		(item: Label) => item.id !== label.id
	);
};

const getLabelById = async () => {
	const entry_id = props.entry?.exist_entry_id;
	if (entry_id) {
		loading.value = true;
		try {
			CollectSearchCancelTokenSource && CollectSearchCancelTokenSource.cancel();
			CollectSearchCancelTokenSource = CancelToken.source();
			const res = await getLabel(
				entry_id,
				CollectSearchCancelTokenSource.token
			);
			entryLabels.value = res.items;
		} catch (error) {
			console.error('Error fetching labels for entry:', error);
		}
		loading.value = false;
	}
};

const syncLabelInWise = () => {
	if (process.env.APPLICATION === 'WISE') {
		rssStore.syncLabels();
	}
};
const exist_entry_id = computed(() => props.entry?.exist_entry_id || '');
watch(exist_entry_id, (newValue) => {
	if (newValue) {
		getLabelById();
	}
});

onMounted(() => {
	getWiseLabels();
	getLabelById();
});
</script>

<style lang="scss"></style>
