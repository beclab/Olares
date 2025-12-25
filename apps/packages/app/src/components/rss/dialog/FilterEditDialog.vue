<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="
			edit
				? t('dialog.edit_filtered_view')
				: title || t('dialog.create_a_new_view')
		"
		@onSubmit="onOK"
		:okLoading="isLoading"
		:okDisabled="okDisabled"
		:ok="t('base.confirm')"
		:cancel="t('base.cancel')"
	>
		<div class="column">
			<div class="prompt-name q-mb-xs text-body3">
				{{ t('base.name') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="filterName"
				borderless
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				no-error-icon
				placeholder=""
			/>

			<div
				v-if="edit || createWithQuery"
				class="prompt-name q-mb-xs text-body3 q-mt-lg"
			>
				{{ t('base.description') }}
			</div>
			<edit-view
				v-if="edit || createWithQuery"
				height="52px"
				class="prompt-style q-mt-xs"
				v-model="filterDesc"
			/>
			<div
				v-if="edit || createWithQuery"
				class="prompt-name q-mb-xs text-body3 q-mt-lg"
			>
				{{ t('base.query') }}
			</div>
			<q-input
				v-if="edit || createWithQuery"
				class="prompt-input text-body3"
				v-model="filterQuery"
				borderless
				no-error-icon
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				placeholder=""
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useFilterStore } from '../../../stores/rss-filter';
import { FilterInfo } from '../../../utils/rss-types';
import { PropType } from 'vue/dist/vue';
import EditView from '../EditView.vue';
import { useI18n } from 'vue-i18n';
import { onMounted, ref, computed } from 'vue';

const props = defineProps({
	data: {
		type: Object as PropType<FilterInfo>,
		require: false
	},
	title: {
		type: String,
		required: false
	},
	createWithQuery: {
		type: Boolean,
		required: false
	}
});

const { t } = useI18n();
const filterDesc = ref('');
const filterName = ref('');
const filterQuery = ref('');
const edit = ref(!!props.data);
const filterStore = useFilterStore();
const isLoading = ref(false);
const customRef = ref();
onMounted(() => {
	if (edit.value) {
		filterName.value = props.data.name;
		filterQuery.value = props.data.query;
		filterDesc.value = props.data.description;
	}
});

const okDisabled = computed(() => {
	if (!filterName.value) {
		return true;
	}

	if (edit.value) {
		const hasChanges =
			filterName.value !== props.data.name ||
			filterQuery.value !== props.data.query ||
			filterDesc.value !== props.data.description;
		return !hasChanges;
	}

	return false;
});

const onOK = async () => {
	if (edit.value) {
		const hasChanges =
			filterName.value !== props.data.name ||
			filterQuery.value !== props.data.query ||
			filterDesc.value !== props.data.description;

		if (!hasChanges) {
			console.log('not changed！');
			return;
		}
	}
	isLoading.value = true;

	if (edit.value) {
		console.log(props.data);
		filterStore
			.modifyFilter({
				...props.data,
				name: filterName.value,
				description: filterDesc.value,
				query: filterQuery.value
			})
			.then(() => {
				customRef.value.onDialogOK();
			})
			.finally(() => {
				isLoading.value = false;
			});
	} else {
		filterStore
			.addFilter(filterName.value, filterDesc.value, filterQuery.value)
			.then((filter) => {
				customRef.value.onDialogOK(filter);
			})
			.finally(() => {
				isLoading.value = false;
			});
	}
};
</script>

<style scoped lang="scss">
.prompt-name {
	color: $ink-3;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}
.prompt-style {
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;
}

.prompt-input {
	padding-left: 7px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;
	height: 32px;
}
</style>
