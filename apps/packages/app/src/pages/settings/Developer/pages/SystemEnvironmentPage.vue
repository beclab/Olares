<template>
	<page-title-component
		:show-back="true"
		:title="t('System environment variables')"
	>
	</page-title-component>

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<app-menu-feature
			image="settings/imgs/root/env.svg"
			:label="t('Environment variables')"
			:description="t('Manage shared settings for your apps.')"
			:button="t('Add environment variables')"
			@on-button-click="addEnv"
		/>

		<bt-list v-if="systemEnvTemp.length > 0" :label="t('System config')">
			<div class="column item-margin-left item-margin-right">
				<q-table
					flat
					:bordered="false"
					:rows="systemEnvTemp"
					:columns="columns"
					row-key="id"
					hide-header
					hide-pagination
					hide-selected-banner
					wrap-cells
					hide-bottom
					:rowsPerPageOptions="[0]"
				>
					<template v-slot:body-cell-data="props">
						<q-td :props="props" class="data-cell" no-hover>
							<div class="column justify-start env-data-content">
								<div class="row justify-start items-center env-name-row">
									<div
										class="text-body1 env-name-text"
										:class="props.row.editable ? 'text-ink-1' : 'text-ink-3'"
									>
										{{ props.row.envName }}
									</div>
									<q-icon
										v-if="!!props.row.valueFrom || props.row.description"
										size="16px"
										name="sym_r_help"
										class="text-ink-3 q-ml-xs"
									>
										<q-tooltip
											self="top left"
											class="text-body3"
											:offset="[0, 0]"
										>
											<div style="max-width: 284px">
												<div v-if="props.row.description" class="tootip-text">
													{{ props.row.description }}
												</div>
												<div v-if="!!props.row.valueFrom" class="tootip-text">
													{{
														t(
															'This value is set by a system environment variable',
															{
																envName: props.row.valueFrom.envName
															}
														)
													}}
												</div>
											</div>
										</q-tooltip>
									</q-icon>
								</div>
								<div class="text-ink-3 text-body3 env-value-text">
									{{ getDisplayValue(props.row) }}
								</div>
							</div>
						</q-td>
					</template>
					<template v-slot:body-cell-actions="props">
						<q-td :props="props" class="actions-cell" no-hover>
							<div class="actions-wrapper">
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
								<q-btn
									v-else-if="props.row.type === 'password'"
									class="btn-size-sm btn-no-text btn-no-border"
									:icon="
										showPasswordList.includes(props.row.key)
											? 'sym_r_visibility_off'
											: 'sym_r_visibility'
									"
									color="ink-2"
									outline
									@click.stop
									no-caps
									@click="changePwdShowType(props.row)"
								>
									<bt-tooltip
										:label="
											showPasswordList.includes(props.row.key)
												? t('Hide')
												: t('View')
										"
									/>
								</q-btn>
							</div>
						</q-td>
					</template>
				</q-table>
			</div>
		</bt-list>
		<bt-list v-if="userEnvTemp.length > 0" :label="t('User information')">
			<div class="column item-margin-left item-margin-right">
				<q-table
					flat
					:bordered="false"
					:rows="userEnvTemp"
					:columns="columns"
					wrap-cells
					row-key="id"
					hide-header
					hide-pagination
					hide-selected-banner
					hide-bottom
					:rowsPerPageOptions="[0]"
				>
					<template v-slot:body-cell-data="props">
						<q-td :props="props" class="data-cell" no-hover>
							<div class="column justify-start env-data-content">
								<div class="row justify-start items-center env-name-row">
									<div
										class="text-body1 env-name-text"
										:class="props.row.editable ? 'text-ink-1' : 'text-ink-3'"
									>
										{{ props.row.envName }}
									</div>
									<q-icon
										v-if="!!props.row.valueFrom || props.row.description"
										size="16px"
										name="sym_r_help"
										class="text-ink-3 q-ml-xs"
									>
										<q-tooltip
											self="top left"
											class="text-body3"
											:offset="[0, 0]"
										>
											<div style="max-width: 284px">
												<div v-if="props.row.description" class="tootip-text">
													{{ props.row.description }}
												</div>
												<div v-if="!!props.row.valueFrom" class="tootip-text">
													{{
														t(
															'This value is set by a system environment variable',
															{
																envName: props.row.valueFrom.envName
															}
														)
													}}
												</div>
											</div>
										</q-tooltip>
									</q-icon>
								</div>
								<div class="text-ink-3 text-body3 env-value-text">
									{{ getDisplayValue(props.row) }}
								</div>
							</div>
						</q-td>
					</template>
					<template v-slot:body-cell-actions="props">
						<q-td :props="props" class="actions-cell" no-hover>
							<div class="actions-wrapper">
								<q-btn
									v-if="props.row.editable && !props.row.required"
									class="btn-size-md btn-no-text btn-no-border"
									icon="sym_r_delete"
									color="ink-2"
									outline
									@click.stop
									no-caps
									@click="deleteEnvironment(props.row)"
								>
									<bt-tooltip :label="t('base.delete')" />
								</q-btn>
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
								<q-btn
									v-else-if="props.row.type === 'password'"
									class="btn-size-sm btn-no-text btn-no-border"
									:icon="
										showPasswordList.includes(props.row.key)
											? 'sym_r_visibility_off'
											: 'sym_r_visibility'
									"
									color="ink-2"
									outline
									@click.stop
									no-caps
									@click="changePwdShowType(props.row)"
								>
									<bt-tooltip
										:label="
											showPasswordList.includes(props.row.key)
												? t('Hide')
												: t('View')
										"
									/>
								</q-btn>
							</div>
						</q-td>
					</template>
				</q-table>
			</div>
		</bt-list>
		<empty-component
			class="q-pb-xl"
			v-if="userEnvTemp.length === 0 && systemEnvTemp.length === 0"
			:info="t('No available environment variable configurations')"
			:empty-image-top="40"
		/>
		<div
			v-if="userEnvTemp.length > 0 || systemEnvTemp.length > 0"
			class="row justify-end q-mt-lg"
		>
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
import AddEnvironmentDialog from 'src/pages/settings/Developer/pages/dialog/AddEnvironmentDialog.vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import BtTooltip from 'src/components/base/BtTooltip.vue';
import { BtDialog, BtNotify, NotifyDefinedType, useColor } from '@bytetrade/ui';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { useAdminStore } from 'src/stores/settings/admin';
import { BaseEnv, UpdateEnvBody } from 'src/constant';
import { onBeforeRouteLeave } from 'vue-router';
import { computed, onMounted, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import {
	deleteUserEnv,
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
		sortable: true
	},
	{
		name: 'actions',
		align: 'right',
		sortable: false
	}
];

const systemEnv = ref<BaseEnv[]>([]);
const userEnv = ref<BaseEnv[]>([]);
const showPasswordList = ref<string[]>([]);

const systemEnvTemp = computed(() => {
	return systemEnv.value
		.map((env) => {
			const change = pendingChanges.value.get(env.envName);
			return change ? { ...env, value: change.value } : env;
		})
		.sort((a, b) =>
			a.envName.localeCompare(b.envName, undefined, { sensitivity: 'base' })
		);
});

const userEnvTemp = computed(() => {
	return userEnv.value
		.map((env) => {
			const change = pendingChanges.value.get(env.envName);
			return change ? { ...env, value: change.value } : env;
		})
		.sort((a, b) =>
			a.envName.localeCompare(b.envName, undefined, { sensitivity: 'base' })
		);
});

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

const addEnv = () => {
	$q.dialog({
		component: AddEnvironmentDialog
	}).onOk((result) => {
		if (result.envName) {
			const userIndex = userEnv.value.findIndex(
				(e) => e.envName === result.envName
			);
			if (userIndex === -1) {
				userEnv.value.push(result);
			}
		}
	});
};

const editEnvironment = (item: BaseEnv) => {
	$q.dialog({
		component: EditEnvironmentDialog,
		componentProps: {
			data: { ...item },
			enableEnvRefPicker: false
		}
	}).onOk((data: any) => {
		if (data && data.value !== item.value) {
			const isSystem = systemEnv.value.some((e) => e.envName === item.envName);
			pendingChanges.value.set(item.envName, {
				value: data.value,
				type: isSystem ? 'system' : 'user'
			});
			notifySuccess(t('Changes saved temporarily'));
		}
	});
};

const deleteEnvironment = (item: BaseEnv) => {
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('Delete environment variable'),
			message: t(
				'Are you sure to delete the environment variable? It cannot be recovered after deletion.',
				{
					envName: item.envName
				}
			),
			confirmText: t('confirm'),
			cancelText: t('cancel')
		}
	}).onOk(async () => {
		deleteUserEnv(item.envName)
			.then(() => {
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('success')
				});

				const userIndex = userEnv.value.findIndex(
					(e) => e.envName === item.envName
				);
				if (userIndex !== -1) {
					userEnv.value.splice(userIndex, 1);
				}
				pendingChanges.value.delete(item.envName);
			})
			.catch((e) => {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: e.response.data.message || e.message
				});
			});
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
		const errorMessage = err instanceof Error ? err.message : String(err);
		notifyFailed(t('Failed to save changes') + errorMessage);
	} finally {
		$q.loading.hide();
	}
};

const changePwdShowType = (row) => {
	const index = showPasswordList.value.indexOf(row.key);
	if (index > -1) {
		showPasswordList.value.splice(index, 1);
	} else {
		showPasswordList.value.push(row.key);
	}
};

const getDisplayValue = (row) => {
	let realValue;
	if (row.value?.toString().trim() === undefined) {
		realValue = row.default?.toString().trim();
	} else {
		realValue = row.value?.toString().trim();
	}

	if (row.type === 'password' && realValue) {
		if (showPasswordList.value.includes(row.key)) {
			return realValue;
		}
		return '••••••••';
	}
	return realValue || '(empty)';
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

.data-cell {
	min-height: 64px;
	vertical-align: middle;
	min-width: 0;
}

.env-data-content {
	min-width: 0;
}

.env-name-row {
	min-width: 0;
}

.env-name-text,
.env-value-text {
	min-width: 0;
	white-space: normal;
	overflow-wrap: anywhere;
	word-break: break-word;
}

.actions-cell {
	vertical-align: middle;
}

.actions-wrapper {
	display: flex;
	flex-direction: row;
	align-items: center;
	justify-content: flex-end;
	height: 100%;
	color: $ink-2;
}

.tootip-text {
	min-width: 0;
	white-space: normal;
	overflow-wrap: anywhere;
	word-break: break-word;
}
</style>
