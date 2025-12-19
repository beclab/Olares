<template>
	<q-card class="env-container" flat>
		<q-card-section class="text-h6 text-ink-1 row items-center justify-between">
			<div class="row items-center justify-center">
				<span>{{ t('docker.storage_volume') }}</span>
				<!-- <q-icon class="text-ink-3 q-ml-xs" name="sym_r_help" size="16px">
					<q-tooltip> {{ t('docker.storage_volume') }} </q-tooltip>
				</q-icon> -->
			</div>
			<q-btn
				class="text-teal-pressed btn-size-sm"
				:label="t('config.add')"
				outline
				icon="sym_r_add"
				no-caps
				@click="createMountsConfig"
			/>
		</q-card-section>

		<q-card-section
			class="q-py-none q-pb-lg"
			v-if="mountsConfig && mountsConfig.length > 0"
		>
			<custom-table :columns="columns" :data="mountsConfig">
				<template v-slot:body-cell-actions="{ props }">
					<q-icon
						class="text-ink-2 cursor-pointer"
						name="sym_r_edit_square"
						size="24px"
						@click="edit(props)"
					/>
					<q-icon
						class="text-ink-2 q-ml-md cursor-pointer"
						name="sym_r_delete"
						size="24px"
						@click="remove(props)"
					/>
				</template>
			</custom-table>
		</q-card-section>
	</q-card>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { VariableType } from '@apps/studio/src/types/core';
import CustomTable from './../common/CustomTable.vue';
import DialogMountsAddress from './../dialog/DialogMountsAddress.vue';

const emits = defineEmits(['updateMounts']);

const $q = useQuasar();
const { t } = useI18n();

const columns = [
	{
		name: 'key',
		label: t('docker.host_path'),
		align: 'left'
	},
	{ name: 'value', label: t('docker.container_path'), align: 'left' },
	{ name: 'operate', label: t('operations'), align: 'right' }
];

const mountsConfig = ref<VariableType[]>([]);

const createMountsConfig = () => {
	$q.dialog({
		component: DialogMountsAddress,
		componentProps: {
			mountsConfig: mountsConfig.value
		}
	}).onOk((res) => {
		const id = Date.now();
		const data = { ...res, id };

		mountsConfig.value.push(data);
	});
};

const edit = (value) => {
	$q.dialog({
		component: DialogMountsAddress,
		componentProps: {
			defaultMounts: value,
			mountsConfig: mountsConfig.value
		}
	}).onOk((res) => {
		if (res.id) {
			mountsConfig.value[
				mountsConfig.value.findIndex((item) => item.id === res.id)
			] = res;
		}
	});
};

const remove = (value) => {
	const index = mountsConfig.value.findIndex((item) => item.id === value.id);
	mountsConfig.value.splice(index, 1);
};

const updateMemory = () => {
	const envMap = {};
	for (const env of mountsConfig.value) {
		envMap[env.key] = env.value;
	}

	emits('updateMounts', envMap);
};

watch(
	() => mountsConfig,
	() => {
		updateMemory();
	},
	{
		deep: true
	}
);
</script>

<style lang="scss" scoped>
.env-container {
	margin: 20px 20px 0 20px;
	padding: 4px;
	border-radius: 12px;
	background-color: $background-1;

	.header-row {
		height: 36px;
		padding: 0 12px;
		border-bottom: 1px solid rgba(0, 0, 0, 0.1);
	}

	.body-row {
		height: 64px;
		padding: 0 12px;
		border-bottom: 1px solid rgba(0, 0, 0, 0.1);

		&:hover {
			background-color: rgba($color: #000000, $alpha: 0.05);
		}
	}
}
</style>
../common/CustomTable.vue../dialog/DialogMountsAddress.vue
