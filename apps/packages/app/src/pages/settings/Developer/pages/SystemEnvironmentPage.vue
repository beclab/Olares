<template>
	<page-title-component
		:show-back="true"
		:title="t('System Environment Variables')"
	>
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<div
				v-if="allEnv.length > 0"
				class="column item-margin-left item-margin-right"
			>
				<q-table
					flat
					:bordered="false"
					:rows="allEnv"
					:columns="columns"
					row-key="id"
					hide-header
					hide-pagination
					hide-selected-banner
					hide-bottom
					:rowsPerPageOptions="[0]"
				>
					<template v-slot:body-cell-data="props">
						<q-td :props="props" style="height: 64px" no-hover>
							<div class="column justify-start">
								<div
									class="text-body1"
									:class="props.row.editable ? 'text-ink-1' : 'text-ink-3'"
								>
									{{ props.row.envName }}
								</div>
								<div class="text-ink-3 text-body3">
									{{ getDisplayValue(props.row) }}
								</div>
							</div>
						</q-td>
					</template>
					<template v-slot:body-cell-actions="props">
						<q-td
							:props="props"
							style="height: 64px"
							class="text-ink-2 row items-center justify-end"
							no-hover
						>
							<q-btn
								v-if="props.row.editable"
								class="btn-size-sm btn-no-text btn-no-border"
								icon="sym_r_edit_square"
								color="ink-2"
								outline
								@click.stop
								no-caps
								@click="editEnvironment(props.row)"
							>
								<bt-tooltip :label="t('base.edit')" />
							</q-btn>
						</q-td>
					</template>
				</q-table>
			</div>
			<empty-component
				class="q-pb-xl"
				v-else
				:info="t('No available environment variable configurations')"
				:empty-image-top="40"
			/>
		</bt-list>
		<div v-if="systemEnv.length > 0" class="row justify-end q-mt-lg">
			<q-btn
				dense
				:disable="!hasChanges"
				flat
				class="confirm-btn q-px-md q-mb-lg"
				:label="t('apply')"
				@click="submitAllChanges"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import EditEnvironmentDialog from 'src/pages/settings/Developer/pages/dialog/EditEnvironmentDialog.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import BtTooltip from 'src/components/base/BtTooltip.vue';

import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { useAdminStore } from 'src/stores/settings/admin';
import { onBeforeRouteLeave } from 'vue-router';
import { computed, onMounted, ref } from 'vue';
import { BaseEnv, UpdateEnvBody } from 'src/constant';
import { BtDialog, useColor } from '@bytetrade/ui';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import {
	getSystemEnvList,
	getUserEnvList,
	updateSystemEnv,
	updateUserEnv
} from 'src/api/settings/env';

const { t } = useI18n();
const $q = useQuasar();
const adminStore = useAdminStore();

const columns: any = [
	{
		name: 'data',
		align: 'left',
		sortable: false
	},
	{
		name: 'actions',
		align: 'right',
		sortable: false
	}
];

const systemEnv = ref<BaseEnv[]>([]);
const userEnv = ref<BaseEnv[]>([]);

const pendingChanges = ref<
	Map<string, { value: string; type: 'system' | 'user' }>
>(new Map());

onMounted(async () => {
	const systemData = await getSystemEnvList();
	if (adminStore.isNormal) {
		systemEnv.value = systemData.map((item) => {
			item.editable = false;
			return item;
		});
	} else {
		systemEnv.value = systemData;
	}
	userEnv.value = await getUserEnvList();
});

const allEnv = computed(() => {
	return systemEnv.value
		.concat(userEnv.value)
		.map((env) => {
			const change = pendingChanges.value.get(env.envName);
			return change ? { ...env, value: change.value } : env;
		})
		.sort((a, b) =>
			a.envName.localeCompare(b.envName, undefined, { sensitivity: 'base' })
		);
});

const hasChanges = computed(() => pendingChanges.value.size > 0);
const { color: blue } = useColor('blue-default');
const { color: textInk } = useColor('ink-2');

onBeforeRouteLeave((to, from, next) => {
	if (hasChanges.value) {
		BtDialog.show({
			title: t('Warning'),
			message: t('You have unsaved changes. Leave without saving?'),
			okStyle: {
				background: blue.value,
				color: textInk.value
			},
			okText: t('Leave'),
			cancelText: t('base.cancel')
		})
			.then((res) => {
				if (res) {
					next();
				} else {
					next(false);
				}
			})
			.catch((err) => {
				console.log('click error', err);
			});
	} else {
		next();
	}
});

const editEnvironment = (item: BaseEnv) => {
	$q.dialog({
		component: EditEnvironmentDialog,
		componentProps: {
			data: { ...item }
		}
	}).onOk((data: any) => {
		if (data && data.value !== item.value) {
			const isSystem = systemEnv.value.some((e) => e.envName === item.envName);
			pendingChanges.value.set(data.key, {
				value: data.value,
				type: isSystem ? 'system' : 'user'
			});
			notifySuccess(t('Changes saved temporarily'));
		}
	});
};

const submitAllChanges = async () => {
	if (pendingChanges.value.size === 0) return;

	$q.loading.show();
	try {
		const systemUpdates: UpdateEnvBody = [];
		const userUpdates: UpdateEnvBody = [];

		pendingChanges.value.forEach((change, envName) => {
			const item = { envName, value: change.value };
			if (change.type === 'system') {
				systemUpdates.push(item);
			} else {
				userUpdates.push(item);
			}
		});

		const updatePromises: Promise<BaseEnv[]>[] = [];
		if (systemUpdates.length > 0) {
			updatePromises.push(updateSystemEnv(systemUpdates));
		}
		if (userUpdates.length > 0) {
			updatePromises.push(updateUserEnv(userUpdates));
		}

		const results = await Promise.all(updatePromises);

		const allUpdatedEnvs = results.flat();
		allUpdatedEnvs.forEach((updatedEnv) => {
			const systemIndex = systemEnv.value.findIndex(
				(e) => e.envName === updatedEnv.envName
			);
			if (systemIndex !== -1) {
				systemEnv.value[systemIndex] = updatedEnv;
				return;
			}
			const userIndex = userEnv.value.findIndex(
				(e) => e.envName === updatedEnv.envName
			);
			if (userIndex !== -1) {
				userEnv.value[userIndex] = updatedEnv;
			}
		});

		pendingChanges.value.clear();
		notifySuccess(t('All changes saved successfully'));
	} catch (err) {
		console.error('Batch submit failed:', err);
		notifyFailed(t('Failed to save changes') + (err.message || ''));
	} finally {
		$q.loading.hide();
	}
};

const getDisplayValue = (row) => {
	const baseValue = row.value || row.default;

	if (row.type === 'password' && baseValue) {
		return '•'.repeat(baseValue.length);
	}
	return baseValue ? baseValue : '(empty)';
};
</script>

<style scoped lang="scss">
.add-btn {
	border-radius: 8px;
	padding: 6px 12px;
	border: 1px solid $separator;
	cursor: pointer;
	text-decoration: none;

	.add-title {
		color: $ink-2;
	}
}

.add-btn:hover {
	background-color: $background-3;
}
</style>
