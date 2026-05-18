<template>
	<page-title-component
		:show-back="true"
		:title="t('Manage environment variables')"
	>
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<div
				v-if="environmentList.length > 0"
				class="column item-margin-left item-margin-right"
			>
				<q-table
					flat
					:bordered="false"
					:rows="environmentList"
					:columns="columns"
					row-key="id"
					hide-header
					hide-pagination
					hide-selected-banner
					hide-bottom
					wrap-cells
					:rowsPerPageOptions="[0]"
				>
					<template v-slot:body-cell-data="props">
						<q-td :props="props" class="data-cell" no-hover>
							<div class="column justify-start">
								<div class="row justify-start items-center">
									<div
										class="text-body1 row"
										:class="props.row.editable ? 'text-ink-1' : 'text-ink-3'"
									>
										{{ props.row.envName }}
									</div>
									<q-icon
										v-if="!!props.row.valueFrom || props.row.description"
										size="20px"
										name="sym_r_info"
										class="text-ink-3"
									>
										<q-tooltip
											self="top left"
											class="text-body3"
											:offset="[0, 0]"
										>
											<div style="max-width: 284px">
												<div v-if="props.row.description">
													{{ props.row.description }}
												</div>
												<div v-if="!!props.row.valueFrom">
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
								<div class="text-ink-3 text-body3">
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
			<empty-component
				class="q-pb-xl"
				v-else
				:info="t('No available environment variable configurations')"
				:empty-image-top="40"
			/>
		</bt-list>
		<div class="row justify-end">
			<q-btn
				v-if="environmentList.length > 0"
				dense
				flat
				:disabled="!hasChanges"
				:loading="isLoading"
				class="confirm-btn q-px-md q-my-lg"
				:label="t('apply')"
				@click="onSubmit"
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
import { getAppEnv, updateAppEnv } from 'src/api/settings/env';
import { onBeforeRouteLeave, useRoute } from 'vue-router';
import { BtDialog, useColor } from '@bytetrade/ui';
import { BaseEnv } from 'src/constant';
import { onMounted, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const $q = useQuasar();
const route = useRoute();
const isLoading = ref(false);
const appName = route.query.appName as string;
const environmentList = ref<BaseEnv[]>([]);
const showPasswordList = ref<string[]>([]);

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

const hasChanges = ref(false);
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

onMounted(async () => {
	try {
		environmentList.value = await getAppEnv(appName);
	} catch (error) {
		notifyFailed(error.message || error.response?.data?.message || error);
	}
});

const editEnvironment = (item: BaseEnv) => {
	$q.dialog({
		component: EditEnvironmentDialog,
		componentProps: {
			data: item
		}
	}).onOk(async (data: any) => {
		if (data) {
			const environment = environmentList.value.find((item) => {
				return item.envName === data.key;
			});
			if (environment) {
				environment.value = data.value;
				if (data.applyOnChange !== undefined) {
					environment.applyOnChange = data.applyOnChange;
				}
				if (data.valueFrom === null || data.valueFrom === undefined) {
					delete environment.valueFrom;
				} else {
					environment.valueFrom = data.valueFrom;
				}
				hasChanges.value = true;
			}
			notifySuccess(t('Changes saved temporarily'));
		}
	});
};

const onSubmit = () => {
	isLoading.value = true;
	updateAppEnv(
		appName,
		environmentList.value
			.filter((item) => item.editable)
			.map((item) => {
				const row: {
					envName: string;
					value: string;
					applyOnChange?: boolean;
					valueFrom?: BaseEnv['valueFrom'];
				} = {
					envName: item.envName,
					value: item.value || ''
				};
				if (item.applyOnChange !== undefined) {
					row.applyOnChange = item.applyOnChange;
				}
				if (item.valueFrom !== undefined) {
					row.valueFrom = item.valueFrom;
				}
				return row;
			})
	)
		.then((updateEnvList: BaseEnv[]) => {
			hasChanges.value = false;
			notifySuccess(t('All changes saved successfully'));
			environmentList.value = updateEnvList;
		})
		.catch((err) => {
			console.error(err.message || err.response?.data?.message || err);
			notifyFailed(t('Failed to save changes') + (err.message || ''));
		})
		.finally(() => {
			isLoading.value = false;
		});
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
		// return '•'.repeat(baseValue.length);
		return '••••••••';
	}
	return realValue ? realValue : '(empty)';
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
</style>
