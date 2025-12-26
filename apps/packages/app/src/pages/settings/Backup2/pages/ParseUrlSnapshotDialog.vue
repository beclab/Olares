<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('snapshots')"
		:ok="false"
		size="medium"
		:platform="deviceStore.platform"
	>
		<adaptive-layout>
			<template v-slot:pc>
				<div class="border-radius-8 q-pa-lg bg-background-6">
					<div class="full-width full-height relative-position">
						<q-table
							v-model:pagination="pagination"
							:rows="rows"
							:columns="columns"
							row-key="name"
							flat
							class="bg-transparent"
							:loading="loading"
							@request="onRequest"
						>
							<template #body-cell-size="props">
								<td>
									{{ props.value }}
								</td>
							</template>
							<template #body-cell-snapshotTime="props">
								<td>
									{{ props.value }}
								</td>
							</template>
							<template #body-cell-action="props">
								<q-td
									:props="props"
									class="snapshot-body text-subtitle2 text-orange-default cursor-pointer"
									@click="onItemClick(props.row)"
								>
									{{ t('restore') }}
								</q-td>
							</template>
							<template #no-data>
								<div class="row justify-center full-width relative-position">
									<empty-view />
								</div>
							</template>
							<template #loading>
								<div
									class="row justify-center items-center full-width"
									style="min-height: 400px; position: absolute"
								>
									<bt-loading :loading="true" size="50px" />
								</div>
							</template>
						</q-table>
					</div>
				</div>
			</template>

			<template v-slot:mobile>
				<div class="details-mobile q-pa-lg">
					<q-infinite-scroll
						@load="onLoad"
						:offset="350"
						:disable="!mobileLoading"
					>
						<div
							class="column flex-gap-lg table-card-container full-width"
							style="overflow: hidden"
						>
							<div
								v-for="item in rows"
								:key="item.id"
								class="bg-background-6 border-radius-12 full-width"
							>
								<div class="q-mx-lg q-my-md row justify-between items-center">
									<div class="column justify-start">
										<div class="text-subtitle3 text-grey-10">
											{{ calculateTime(item.createAt) }}
										</div>
										<div class="text-ink-3 q-mt-xs text-overline">
											{{ calculateSize(item.size) }}
										</div>
									</div>
									<q-icon
										size="24px"
										class="cursor-pointer text-orange-default"
										name="sym_r_downloading"
										@click="onItemClick(item)"
									/>
								</div>
							</div>
						</div>
						<template v-slot:loading>
							<div class="row justify-center q-my-md">
								<q-spinner-dots color="primary" size="40px" />
							</div>
						</template>
					</q-infinite-scroll>
				</div>
			</template>
		</adaptive-layout>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import { date, useQuasar } from 'quasar';
import { useDeviceStore } from 'src/stores/device';
import { useBackupStore } from 'src/stores/settings/backup';
import { getSuitableValue } from 'src/utils/settings/monitoring';
import {
	BackupResourcesType,
	SnapshotInfo,
	RestoreSnapshotInfo
} from 'src/constant';
import BtLoading from '../../../../components/base/BtLoading.vue';
import EmptyView from '../../../../components/rss/EmptyView.vue';
import AdaptiveLayout from '../../../../components/settings/AdaptiveLayout.vue';

const { t } = useI18n();
const CustomRef = ref();
const $q = useQuasar();
const deviceStore = useDeviceStore();
const backupStore = useBackupStore();
const loading = ref(false);
const rows = ref<SnapshotInfo[]>([]);
const backupPath = ref();
const resourceType = ref(BackupResourcesType.files);
const mobileLoading = ref(true);
const pagination = ref({
	page: 1,
	rowsPerPage: 20,
	rowsNumber: 0
});

const columns: any = [
	{
		name: 'snapshotTime',
		align: 'left',
		label: t('create_time'),
		field: 'createAt',
		format: (val) => calculateTime(val)
	},
	{
		name: 'size',
		align: 'left',
		label: t('backup_size'),
		field: 'size',
		format: (val, row) => calculateSize(row.size)
	},
	{
		name: 'action',
		align: 'right',
		label: t('action'),
		field: 'action'
	}
];

const props = defineProps({
	url: String,
	pwd: String
});

function getData() {
	loading.value = true;
	backupStore
		.parseUrl(
			props.url,
			props.pwd,
			pagination.value.rowsPerPage * (pagination.value.page - 1),
			pagination.value.rowsPerPage
		)
		.then((data: any) => {
			if (data) {
				pagination.value.rowsNumber = data.totalCount;
				rows.value = data.snapshots ? data.snapshots : [];
				backupPath.value = data.backupPath;
				resourceType.value = data.backupType;
			}
		})
		.finally(() => {
			loading.value = false;
		});
}

function onRequest(props) {
	const { page, rowsPerPage } = props.pagination;
	pagination.value = {
		...pagination.value,
		page: page,
		rowsPerPage: rowsPerPage
	};
	getData();
}

const calculateTime = (time: number) => {
	return time === 0
		? '-'
		: date.formatDate(Number(time * 1000), 'YYYY-MM-DD HH:mm');
};

const calculateSize = (size: number) => {
	return getSuitableValue(size.toString(), 'disk');
};

const onItemClick = (data) => {
	CustomRef.value.onDialogOK({
		data: {
			...data,
			backupPath: backupPath.value
		} as RestoreSnapshotInfo,
		type: resourceType.value
	});
};

const onLoad = () => {
	if (!mobileLoading.value) {
		return;
	}
	loading.value = true;
	backupStore
		.parseUrl(
			props.url,
			props.pwd,
			pagination.value.rowsPerPage * (pagination.value.page - 1),
			pagination.value.rowsPerPage
		)
		.then((data: any) => {
			pagination.value.rowsNumber = data.totalCount;
			rows.value = data.snapshots ? data.snapshots : [];
			backupPath.value = data.backupPath;
			mobileLoading.value =
				pagination.value.rowsNumber >
				pagination.value.page * pagination.value.rowsPerPage +
					rows.value.length;
			if (mobileLoading.value) {
				pagination.value = {
					...pagination.value,
					page: pagination.value.page + 1
				};
			}
		})
		.finally(() => {
			loading.value = false;
		});
};

onMounted(async () => {
	if (!$q.platform.is.mobile) {
		getData();
	}
});
</script>

<style scoped lang="scss">
.date-of-day {
	height: 36px;
	color: $ink-2;
	border-radius: 8px;
	border: 1px solid $input-stroke;
	width: var(--selectedWidth);

	.time-clock {
		border-radius: 4px;
		margin-right: 8px;
	}
}
</style>
