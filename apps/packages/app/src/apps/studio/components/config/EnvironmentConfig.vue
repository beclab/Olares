<template>
	<q-card class="env-container" flat>
		<q-card-section class="text-h6 text-ink-1 row items-center justify-between">
			<div class="row items-center justify-center">
				<span>{{ t('docker.env_variables') }}</span>
				<!-- <q-icon class="text-ink-3 q-ml-xs" name="sym_r_help" size="16px">
					<q-tooltip> {{ t('docker.env_variables') }} </q-tooltip>
				</q-icon> -->
			</div>
			<q-btn
				class="text-teal-pressed btn-size-sm"
				:label="t('config.add')"
				outline
				icon="sym_r_add"
				no-caps
				@click="createEnvConfig"
			/>
		</q-card-section>

		<q-card-section
			class="q-py-none q-mb-lg"
			v-if="envConfig && envConfig.length > 0"
		>
			<custom-table :columns="columns" :data="envConfig">
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
import DialogEnvironmentConfig from './../dialog/DialogEnvironmentConfig.vue';

const emits = defineEmits(['updateEnv']);

const $q = useQuasar();
const { t } = useI18n();

const columns = [
	{
		name: 'key',
		label: t('docker.key'),
		align: 'left'
	},
	{ name: 'value', label: t('docker.value'), align: 'left' },
	{ name: 'operate', label: t('operations'), align: 'right' }
];

const envConfig = ref<VariableType[]>([]);

const createEnvConfig = () => {
	$q.dialog({
		component: DialogEnvironmentConfig,
		componentProps: {
			envConfig: envConfig.value
		}
	}).onOk((res) => {
		const id = Date.now();
		const data = { ...res, id };

		envConfig.value.push(data);
	});
};

const edit = (value) => {
	$q.dialog({
		component: DialogEnvironmentConfig,
		componentProps: {
			defaultEnv: value,
			envConfig: envConfig.value
		}
	}).onOk((res) => {
		if (res.id) {
			envConfig.value[envConfig.value.findIndex((item) => item.id === res.id)] =
				res;
		}
	});
};

const remove = (value) => {
	const index = envConfig.value.findIndex((item) => item.id === value.id);
	envConfig.value.splice(index, 1);
};

const updateEnv = () => {
	const envMap = {};
	for (const env of envConfig.value) {
		envMap[env.key] = env.value;
	}

	emits('updateEnv', envMap);
};

watch(
	() => envConfig,
	() => {
		updateEnv();
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
}
</style>
../common/CustomTable.vue../dialog/DialogEnvironmentConfig.vue
