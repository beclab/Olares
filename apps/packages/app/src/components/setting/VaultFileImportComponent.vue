<template>
	<TerminusSelectLocalFile
		@on-success="selectFiles"
		accept="text/plain,.csv,.pls,.set,.pbes2,.1pux,.json"
	>
		<div class="adminBtn q-mt-md text-body3 row items-center">
			<q-icon name="sym_r_upload_file" class="icon q-mr-xs" size="16px" />
			<span>
				{{ t('Import...') }}
			</span>
		</div>
	</TerminusSelectLocalFile>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import TerminusSelectLocalFile from '../common/TerminusSelectLocalFile.vue';
// import * as impl from '';
import * as imp from '@didvault/sdk/src/import';
import { ref } from 'vue';
import { VaultItem } from '@didvault/sdk/src/core';
import { useQuasar } from 'quasar';
import VaultImportDialog from './VaultImportDialog.vue';
import { useUserStore } from 'src/stores/user';
import { parseData } from 'src/utils/vault';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';
const { t } = useI18n();

const file = ref<File | undefined>();

const formatSelect = ref<imp.ImportFormat>();

const items = ref<VaultItem[]>([]);

const itemColumns = ref<imp.ImportCSVColumn[]>([]);

const userStore = useUserStore();

const $q = useQuasar();

const selectFiles = async (files: File[]) => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	if (files.length == 0) {
		return;
	}
	file.value = files[0];
	formatSelect.value = (await imp.guessFormat(files[0])) || imp.CSV;
	let result = {
		returnItemColumns: [] as imp.ImportCSVColumn[],
		items: [] as VaultItem[]
	};

	try {
		result = await parseData(
			file.value,
			formatSelect.value,
			[],
			$q,
			true,
			true
		);
	} catch (error) {
		console.log('error ===>', error);
	}

	itemColumns.value = result.returnItemColumns;
	items.value = result.items;
	if (items.value.length == 0) {
		notifyFailed(
			t(
				'Unable to read file content. Please ensure you are importing from a valid file.'
			)
		);
		return;
	}
	openImportVaultDialog();
};

const openImportVaultDialog = () => {
	$q.dialog({
		component: VaultImportDialog,
		componentProps: {
			pItems: items.value,
			file: file.value,
			formatSelect: formatSelect.value,
			pItemColumns: itemColumns.value
		}
	}).onOk((s: VaultItem[]) => {
		items.value = s;
	});
};
</script>

<style scoped lang="scss">
.adminBtn {
	border: 1px solid $yellow;
	background-color: $yellow;
	display: inline-block;
	color: $grey-10;
	padding: 7px 12px;
	border-radius: 8px;
	cursor: pointer;

	&:hover {
		background-color: $yellow-3;
	}
}
</style>
