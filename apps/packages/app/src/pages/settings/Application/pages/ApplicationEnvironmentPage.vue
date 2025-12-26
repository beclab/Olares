<template>
	<page-title-component
		:show-back="true"
		:title="t('Manage Environment Variables')"
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
					:rowsPerPageOptions="[0]"
				>
					<template v-slot:body-cell-data="props">
						<q-td :props="props" style="height: 64px" no-hover>
							<div class="column justify-start">
								<div class="row justify-start items-center">
									<div
										class="text-body1 row"
										:class="props.row.editable ? 'text-ink-1' : 'text-ink-3'"
									>
										{{ props.row.envName }}
									</div>
									<q-icon
										v-if="props.row.valueFrom"
										size="20px"
										name="sym_r_info"
										class="text-ink-3"
									>
										<q-tooltip
											self="top left"
											class="text-body3"
											:offset="[0, 0]"
											style="width: 284px"
											>{{
												t(
													'This value is set by a system environment variable',
													{
														envName: props.row.valueFrom.envName
													}
												)
											}}</q-tooltip
										>
									</q-icon>
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
		<div class="row justify-end">
			<q-btn
				v-if="environmentList.length > 0"
				dense
				flat
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

import { getAppEnv, updateAppEnv } from 'src/api/settings/env';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { useQuasar } from 'quasar';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { BaseEnv } from 'src/constant';
import { useRoute } from 'vue-router';

const { t } = useI18n();
const $q = useQuasar();
const route = useRoute();
const appName = route.query.appName as string;

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

const environmentList = ref<BaseEnv[]>([]);

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
			}
			notifySuccess(t('Changes saved temporarily'));
		}
	});
};

const onSubmit = () => {
	updateAppEnv(
		appName,
		environmentList.value
			.filter((item) => item.editable)
			.map((item) => {
				return {
					envName: item.envName,
					value: item.value || ''
				};
			})
	)
		.then((updateEnvList: BaseEnv[]) => {
			notifySuccess(t('All changes saved successfully'));
			environmentList.value = updateEnvList;
		})
		.catch((err) => {
			console.error(err.message || err.response?.data?.message || err);
			notifyFailed(t('Failed to save changes') + (err.message || ''));
		});
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
