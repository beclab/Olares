<template>
	<FullPageWithBack :title="$t('DISK_DETAILS')">
		<div
			v-for="(item, index) in diskData"
			:key="index"
			:class="{ 'q-mt-xl': index != 0 }"
		>
			<MyCard>
				<div class="row flex-gap-lg">
					<q-img
						:src="deviceIcon"
						:ratio="1"
						spinner-size="0px"
						width="32px"
						height="32px"
					/>
					<div class="text-body3" style="flex: 1">
						<div class="row justify-between items-center">
							<div class="text-h6">
								<span>{{ item.headerData.device }}&nbsp;</span>
								<span>({{ item.headerData.rotational }})</span>
							</div>
						</div>
						<div class="row justify-between items-center q-mt-sm">
							<div class="text-body3 text-ink-2">
								<span>{{ $t('DISK_OP.STORAGE_STATUS') }}:</span>
								<DiskStatus
									:label="item.headerData.health_ok"
									:status="item.headerData.health_ok_status"
								></DiskStatus>
							</div>
						</div>
					</div>
					<div class="row justify-between self-center">
						<div
							class="text-body3 text-light-blue-default"
							@click="dialogShow(item.headerData.device, item.headerData.node)"
						>
							<span class="cursor-pointer">{{
								$t('DISK_OP.OCCUPANCY_ANALYSIS')
							}}</span>
							<q-popup-proxy
								:offset="[0, -20]"
								style="padding: 0px; border-radius: 12px; max-width: 650px"
							>
								<q-card flat class="q-pa-lg" style="width: 650px">
									<DiskLsblkTable
										:rows="rows"
										:columns="columns"
										:pagination="pagination"
									/>
								</q-card>
							</q-popup-proxy>
						</div>
					</div>
				</div>
				<q-separator class="q-mt-lg q-mb-xl" color="separator" />
				<Descriptions :data="item.contentData" colWidth="310px"> </Descriptions>
			</MyCard>
		</div>
		<q-inner-loading :showing="loading"> </q-inner-loading>
		<Empty
			class="absolute-center"
			v-show="!loading && diskData.length == 0"
			@click="fetchData"
		></Empty>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import { getNodeMonitoring } from '@apps/dashboard/src/network';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import { getResult } from '@apps/dashboard/src/utils/monitoring';
import {
	getDiskOptions,
	getDiskPartitionRows,
	MetricTypes,
	getLsblkColumns,
	type LsblkFlatRow
} from './config';
import deviceIcon from '@apps/dashboard/assets/device.svg';
import DiskStatus from './DiskStatus.vue';
import Descriptions from '@apps/control-panel-common/components/Descriptions.vue';
import Empty from '@apps/control-panel-common/components/Empty3.vue';
import DiskLsblkTable from './DiskLsblkTable.vue';

const diskResult = ref({});

const diskData = ref(getDiskOptions({}, MetricTypes));
const loading = ref(false);
const pagination = ref({
	rowsNumber: 0
});

const rows = ref<LsblkFlatRow[]>([]);

const { locale } = useI18n();
const columns = computed(() => {
	locale.value;
	return getLsblkColumns();
});

const fetchData = async () => {
	const filters = {
		metrics: Object.values(MetricTypes),
		step: '0s'
	};
	loading.value = true;
	try {
		const { metrics_filter } = getParams(filters);
		const params = {
			metrics_filter
		};
		const res = await getNodeMonitoring(params);
		const result = getResult(res.data.results);
		const data = fillEmptyMetrics(params, result);
		diskResult.value = data;
		diskData.value = getDiskOptions(data, MetricTypes);
	} catch (error) {
		loading.value = false;
	}
	loading.value = false;
};

const dialogShow = (name: string, node: string) => {
	rows.value = getDiskPartitionRows(diskResult.value, name, node);
	pagination.value.rowsNumber = rows.value.length;
};

onMounted(() => {
	fetchData();
});
</script>

<style lang="scss" scoped>
.progress-wrapper {
	::v-deep(.q-linear-progress__track) {
		opacity: 1;
	}
}
.header-right2 {
	width: 360px;
}
</style>
