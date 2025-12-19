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
							<div v-if="item.headerData.capacity_show">
								<q-linear-progress
									size="8px"
									:value="item.headerData.disk_size_ratio"
									color="light-blue-default"
									track-color="background-3"
									class="header-right progress-wrapper"
								/>
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
							<div
								class="row justify-between header-right"
								v-if="item.headerData.capacity_show"
							>
								<div class="row">
									<span class="text-body3 text-light-blue-default">{{
										item.headerData.used_size
									}}</span>
									<q-separator class="q-mx-md" vertical color="ink-3" />
									<span>{{ item.headerData.capacity_size }}</span>
								</div>
								<div
									class="text-body3 text-light-blue-default"
									@click="
										dialogShow(item.headerData.name, item.headerData.node)
									"
								>
									<span class="cursor-pointer">{{
										$t('DISK_OP.OCCUPANCY_ANALYSIS')
									}}</span>
									<q-popup-proxy
										:offset="[0, -20]"
										style="padding: 0px; border-radius: 12px; max-width: 800px"
									>
										<q-card flat class="q-pa-lg" style="width: 800px">
											<div class="row justify-between items-center">
												<span class="text-h6 text-ink-1">{{
													$t('DISK_OP.STORAGE_USAGE')
												}}</span>
												<QButtonStyle style="margin-right: -8px">
													<q-btn flat dense no-caps>
														<q-icon
															name="close"
															v-close-popup
															color="ink-2"
															class="cursor-pointer"
														/>
													</q-btn>
												</QButtonStyle>
											</div>
											<div class="row flex-gap-lg q-mt-lg">
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
															<span>{{ item.headerData.device }}</span>
															<span>({{ item.headerData.rotational }})</span>
														</div>
														<div>
															<q-linear-progress
																size="8px"
																:value="item.headerData.disk_size_ratio"
																color="light-blue-default"
																track-color="background-3"
																class="header-right progress-wrapper"
															/>
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
														<div class="row justify-between header-right">
															<div class="row">
																<span
																	class="text-body3 text-light-blue-default"
																	>{{ item.headerData.used_size }}</span
																>
																<q-separator
																	class="q-mx-md"
																	vertical
																	color="ink-3"
																/>
																<span>{{ item.headerData.capacity_size }}</span>
															</div>
														</div>
													</div>
												</div>
											</div>
											<q-separator class="q-my-lg" color="separator" />
											<QTableStyle2>
												<q-table
													:rows="rows"
													:columns="columns"
													row-key="name"
													flat
													:pagination="pagination"
													hide-pagination
												/>
											</QTableStyle2>
										</q-card>
									</q-popup-proxy>
								</div>
							</div>
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
import { onMounted, ref } from 'vue';
import {
	fillEmptyMetrics,
	getParams
} from '@apps/control-panel-common/src/containers/Monitoring/config';
import { getResult } from '@apps/dashboard/src/utils/monitoring';
import {
	getDiskOptions,
	getDiskPartitionRows,
	MetricTypes,
	columns
} from './config';
import QButtonStyle from '@apps/control-panel-common/components/QButtonStyle.vue';
import deviceIcon from '@apps/dashboard/assets/device.svg';
import DiskStatus from './DiskStatus.vue';
import Descriptions from '@apps/control-panel-common/components/Descriptions.vue';
import QTableStyle2 from '@apps/control-panel-common/components/QTableStyle2.vue';
import Empty from '@apps/control-panel-common/components/Empty3.vue';

const diskResult = ref({});

const diskData = ref(getDiskOptions({}, MetricTypes));
const loading = ref(false);
const pagination = ref({
	rowsNumber: 0
});

const rows = ref([]);

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
.header-right {
	width: 412px;
}
.header-right2 {
	width: 360px;
}
</style>
