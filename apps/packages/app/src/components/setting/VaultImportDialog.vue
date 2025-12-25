<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Import Data')"
		:ok="
			t('Import {num} items', {
				num: items.length
			})
		"
		:cancel="t('cancel')"
		size="medium"
		platform="web"
		:okDisabled="items.length <= 0"
		@onSubmit="importVaults"
	>
		<div class="card-content">
			<div class="login-label q-mb-xs">
				{{ t('format') }}
			</div>

			<bt-select
				v-model="format"
				:options="formatOptions"
				:border="true"
				:disable="!hasChosenCsvSupportedFormat"
				color="text-blue-default"
				@update:model-value="handleFormatSelect"
			/>

			<div v-if="format === imp.CSV.value">
				<div class="text-body2 text-ink-2 q-mt-lg">
					{{
						t(
							'Choose the correct column names and types for each column below.'
						)
					}}
				</div>

				<div class="row items-center justify-between q-mt-lg">
					<div class="text-subtitle1 text-ink-2">
						{{ t('First row contains field names') }}
					</div>

					<bt-switch
						class="custom-toggle-wrapper"
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="csvHasColumnsOnFirstRow"
						@update:model-value="reloadData(true)"
					/>
				</div>

				<div class="login-label q-mb-xs q-mt-lg">
					{{ t('Name column') }}
				</div>

				<bt-select
					v-model="nameColumnSelect"
					:options="nameColumnSelectOptions"
					:border="true"
					color="text-blue-default"
					@update:modelValue="handleNameSelect()"
				/>

				<div class="login-label q-mb-xs q-mt-lg">
					{{ t('Tag column') }}
				</div>

				<bt-select
					v-model="tagColumnSelect"
					:options="tagColumnSelectOptions"
					:border="true"
					color="text-blue-default"
					@update:modelValue="handleTagSelect()"
				/>

				<template
					v-for="(itemColumn, itemColumnIndex) in itemColumns"
					:key="itemColumn.displayName + itemColumnIndex"
				>
					<div
						v-if="itemColumn.type !== 'name' && itemColumn.type !== 'tags'"
						class="item-column q-mt-lg"
					>
						<div class="row items-center justify-between">
							<div class="text-subtitle1 text-ink-1">
								{{
									csvHasColumnsOnFirstRow
										? itemColumn.name
										: $l('Column {column}', {
												column: (itemColumnIndex + 1).toString()
										  })
								}}
							</div>
							<div v-if="csvHasColumnsOnFirstRow" class="text-body1 text-ink-3">
								{{
									$l('Column {column}', {
										column: (itemColumnIndex + 1).toString()
									})
								}}
							</div>
						</div>

						<div class="text-body2 text-ink-2 q-mt-xs">
							{{
								itemColumn.values
									.filter((value) => value !== '')
									.slice(0, 20)
									.map((value) => (value.includes(',') ? `"${value}"` : value))
									.join(', ')
							}}
						</div>

						<terminus-edit
							v-model="itemColumn.displayName"
							:label="$l('Field Name')"
							:show-password-img="false"
							class="q-mt-md"
							:inputHeight="40"
							:emitKey="`${itemColumnIndex}`"
							@onTextChange="handleFieldNameChane"
							@onBlur="fieldNameonBlur(itemColumnIndex)"
						/>

						<div class="login-label q-mb-xs q-mt-lg">
							{{ $l('Field type') }}
						</div>

						<bt-select
							v-model="itemColumn.type"
							:options="fieldTypeOptions"
							:border="true"
							color="text-blue-default"
							@update:modelValue="handleFieldTypeChange(itemColumn)"
						/>
					</div>
				</template>
			</div>

			<div class="login-label q-mb-xs q-mt-lg">
				{{ t('Target vault') }}
			</div>

			<bt-select
				v-model="targetVault"
				:options="vaultsOption"
				:border="true"
				color="text-blue-default"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { computed, PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import TerminusEdit from '../common/TerminusEdit.vue';
import { notifyFailed, notifySuccess } from '../../utils/notifyRedefinedUtil';
import * as imp from '@didvault/sdk/src/import';
import { ImportFormat } from '@didvault/sdk/src/import';
import BtSelect from '../base/BtSelect.vue';
import { FIELD_DEFS, FieldType, VaultItem } from '@didvault/sdk/src/core';
import { app } from 'src/globals';
import { parseData } from 'src/utils/vault';
import { useQuasar } from 'quasar';
import { translate as $l } from '@didvault/sdk/src/util';

const props = defineProps({
	pItems: {
		type: Array as PropType<VaultItem[]>,
		required: true
	},
	formatSelect: {
		type: Object as PropType<ImportFormat>,
		required: true
	},
	file: {
		type: File,
		required: true
	},
	pItemColumns: {
		type: Array as PropType<imp.ImportCSVColumn[]>,
		required: true
	}
});

const { t } = useI18n();
const CustomRef = ref();

const $q = useQuasar();

const itemColumns = ref<imp.ImportCSVColumn[]>(props.pItemColumns);

const csvHasColumnsOnFirstRow = ref(true);

const format = ref(props.formatSelect.value);

const csvSupportedFormatValues = imp.csvSupportedFormats.map(
	(importFormat) => importFormat.value
);

const hasChosenCsvSupportedFormat = Boolean(
	props.formatSelect.value &&
		csvSupportedFormatValues.includes(props.formatSelect.value)
);

const formatOptions = imp.supportedFormats.map((importFormat) => ({
	...importFormat,
	disable:
		hasChosenCsvSupportedFormat &&
		!csvSupportedFormatValues.includes(importFormat.value)
}));

const targetVault = ref(app.mainVault?.id);

const vaultsOption = app.vaults.map((vault) => ({
	disable: !app.isEditable(vault),
	value: vault.id,
	label: vault.name
}));

const items = ref(props.pItems);

const reloadData = async (resetCSVColumns = false) => {
	const formatV = formatOptions.find((e) => e.value == format.value)![0];
	const result = await parseData(
		props.file,
		formatV,
		itemColumns.value,
		$q,
		resetCSVColumns,
		csvHasColumnsOnFirstRow.value
	);
	itemColumns.value = result.returnItemColumns;
	items.value = result.items;
	setDefaultValue();
};

const importVaults = async () => {
	try {
		if (targetVault.value) {
			await app.addItems(items.value as VaultItem[], {
				id: targetVault.value
			});
			notifySuccess(t('success'));
			CustomRef.value.onDialogOK();
		}
	} catch (error) {
		notifyFailed(error);
	}
};

const nameColumnSelect = ref(0);

const nameColumnSelectOptions = computed(() => {
	return itemColumns.value.map((itemColumn, itemColumnIndex) => ({
		label: csvHasColumnsOnFirstRow.value
			? `${itemColumn.displayName} (${$l('Column {column}', {
					column: (itemColumnIndex + 1).toString()
			  })})`
			: $l('Column {column}', {
					column: (itemColumnIndex + 1).toString()
			  }),
		value: itemColumnIndex
	}));
});

const tagColumnSelect = ref(-1);

const tagColumnSelectOptions = computed(() => {
	return [
		{ label: $l('None'), value: -1 },
		...itemColumns.value.map((itemColumn, itemColumnIndex) => ({
			label: csvHasColumnsOnFirstRow.value
				? `${itemColumn.displayName} (${$l('Column {column}', {
						column: (itemColumnIndex + 1).toString()
				  })})`
				: $l('Column {column}', {
						column: (itemColumnIndex + 1).toString()
				  }),
			value: itemColumnIndex
		}))
	];
});

const fieldTypeOptions = Object.keys(FIELD_DEFS).map((fieldType) => ({
	label: $l(FIELD_DEFS[fieldType].name as string),
	value: fieldType
}));

const handleFieldTypeChange = (item: imp.ImportCSVColumn) => {
	reloadData();
};

const handleFieldNameChane = (index: string, value: string) => {
	(itemColumns.value[Number(index)] as any).displayNameCopy = value;
};

const handleFormatSelect = () => {
	reloadData();
};

const handleNameSelect = () => {
	// reloadData();
	const currentNameColumnIndex = itemColumns.value.findIndex(
		(itemColumn) => itemColumn.type === 'name'
	);
	const nameColumnIndex = nameColumnSelect.value || 0;

	itemColumns.value[nameColumnIndex].type = 'name';

	if (currentNameColumnIndex !== -1) {
		itemColumns.value[currentNameColumnIndex].type = FieldType.Text;
	}
	reloadData();
};

const handleTagSelect = () => {
	const currentTagsColumnIndex = itemColumns.value.findIndex(
		(itemColumn) => itemColumn.type === 'tags'
	);
	itemColumns.value[tagColumnSelect.value].type = 'tags';
	if (currentTagsColumnIndex !== -1) {
		itemColumns.value[currentTagsColumnIndex].type = FieldType.Text;
	}
	reloadData();
};

const fieldNameonBlur = (index: number) => {
	if (
		(itemColumns.value[Number(index)] as any).displayNameCopy &&
		(itemColumns.value[Number(index)] as any).displayNameCopy.length > 0 &&
		(itemColumns.value[Number(index)] as any).displayNameCopy !=
			itemColumns.value[index].displayName
	) {
		itemColumns.value[index].displayName = (
			itemColumns.value[Number(index)] as any
		).displayNameCopy;
		delete (itemColumns.value[Number(index)] as any).displayNameCopy;
		reloadData();
	}
};

const setDefaultValue = () => {
	nameColumnSelect.value = itemColumns.value.findIndex(
		({ type }) => type === 'name'
	);
	tagColumnSelect.value = itemColumns.value.findIndex(
		({ type }) => type === 'tags'
	);
};
setDefaultValue();
</script>

<style lang="scss" scoped>
.card-content {
	padding: 0 0px;
	.input {
		border-radius: 5px;
		border: 1px solid $input-stroke;
		background-color: transparent;
		&:focus {
			border: 1px solid $yellow-disabled;
		}
	}

	.item-column {
		border: 1px solid $separator;
		border-radius: 12px;
		padding: 20px;
	}
}
</style>
