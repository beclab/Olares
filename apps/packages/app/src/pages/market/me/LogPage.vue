<template>
	<adaptive-layout>
		<template v-slot:pc>
			<page-container :title-height="56">
				<template v-slot:title>
					<title-bar :show="true" @onReturn="router.back()" />
				</template>
				<template v-slot:page>
					<div class="log-scroll" :style="{ '--paddingX': '44px' }">
						<app-store-body :title="t('logs')" :title-separator="true">
							<template v-slot:right>
								<div class="row justify-end items-center">
									<bt-label
										name="sym_r_page_info"
										:label="t('View Processing Workflow')"
										@click="
											openNewLink('app-store/api/v2/runtime/dashboard-app')
										"
									/>
									<bt-label
										class="q-ml-lg"
										name="sym_r_space_dashboard"
										:label="t('View Dashboard')"
										@click="openNewLink('app-store/api/v2/runtime/dashboard')"
									/>
									<bt-label
										class="q-ml-lg"
										name="sym_r_download"
										:label="t('Download Raw Log')"
										@click="downloadLogs"
									/>
								</div>
							</template>
							<template v-slot:body>
								<q-table
									v-if="rows.length > 0 || loading"
									:rows="rows"
									flat
									:columns="columns"
									row-key="name + time"
									:loading="loading"
									v-model:pagination="pagination"
									hide-pagination
									style="margin-top: 32px"
								>
									<template v-slot:header="props">
										<q-tr :props="props" style="height: 32px">
											<q-th
												v-for="col in props.cols"
												:key="col.name"
												:props="props"
												class="text-body3 text-grey-5"
											>
												{{ col.label }}
											</q-th>
											<q-th auto-width />
										</q-tr>
									</template>

									<template v-slot:body="props">
										<q-tr :props="props">
											<q-td
												v-for="col in props.cols"
												:key="col.name"
												:props="props"
											>
												{{ col.value }}
											</q-td>
											<q-td>
												<q-icon
													v-show="props.row.extended"
													size="20px"
													class="q-mr-md"
													round
													dense
													@click="props.row.expand = !props.row.expand"
													:name="
														props.row.expand
															? 'sym_r_keyboard_arrow_up'
															: 'sym_r_keyboard_arrow_down'
													"
												/>
											</q-td>
										</q-tr>
										<q-tr v-show="props.row.expand" :props="props">
											<q-td colspan="100%">
												<pre class="log-extend">{{
													coverJSONData(props.row.extended)
												}}</pre>
											</q-td>
										</q-tr>
									</template>
									<template v-slot:body-cell-message="props">
										<q-td :props="props">
											<div class="log-message">
												{{ props.row.message }}
											</div>
										</q-td>
									</template>
								</q-table>

								<div
									v-if="rows.length > 0 || loading"
									class="row justify-end q-mt-md"
								>
									<q-pagination
										v-model="pagination.page"
										color="grey-8"
										input
										:max="pagesNumber"
										size="sm"
									/>
								</div>

								<empty-view
									v-else
									:label="t('my.no_logs')"
									class="empty-view"
								/>
							</template>
						</app-store-body>
					</div>
				</template>
			</page-container>
		</template>

		<template v-slot:mobile>
			<page-container :title-height="56">
				<template v-slot:title>
					<title-bar
						:title="t('logs')"
						:show="true"
						:offset="48"
						@onReturn="router.back()"
					>
						<template v-slot:right>
							<div class="row justify-end items-center">
								<bt-label
									name="sym_r_page_info"
									label=""
									@click="openNewLink('app-store/api/v2/runtime/dashboard-app')"
								/>
								<bt-label
									class="q-ml-xs"
									name="sym_r_space_dashboard"
									label=""
									@click="openNewLink('app-store/api/v2/runtime/dashboard')"
								/>
								<bt-label
									class="q-ml-xs"
									name="sym_r_download"
									label=""
									@click="downloadLogs"
								/>
							</div>
						</template>
					</title-bar>
				</template>
				<template v-slot:page>
					<div class="log-scroll" :style="{ '--paddingX': '20px' }">
						<div
							class="log-item bg-background-6 column"
							v-for="item in rows"
							:key="item.id"
						>
							<div class="full-width row justify-between items-center">
								<div class="text-subtitle2-m text-ink-1">
									{{ item.app }}
								</div>

								<div class="text-body2-m text-ink-1">
									{{ formattedDate(item.time) }}
								</div>
							</div>

							<div class="full-width row justify-between items-center q-mt-md">
								<div class="text-body3-m text-ink-3">
									{{ t('account') }}
								</div>

								<div class="text-body3-m text-ink-1">
									{{ item.account }}
								</div>
							</div>

							<div class="full-width row justify-between items-center q-mt-sm">
								<div class="text-body3-m text-ink-3">
									{{ t('type') }}
								</div>
								<div class="text-body3-m text-ink-1">
									{{ item.type }}
								</div>
							</div>

							<div class="full-width row justify-between items-center q-mt-sm">
								<div class="text-body3-m text-ink-3">
									{{ t('message') }}
								</div>
								<div class="text-body3-m text-ink-1">
									{{ item.message ? item.message : '-' }}
								</div>
							</div>

							<q-separator color="separator" class="q-my-md full-width" />

							<pre class="text-body3-m text-ink-1 log-extend">{{
								coverJSONData(item.extended)
							}}</pre>
						</div>

						<app-store-body :title="t('logs')" :title-separator="true">
							<template v-slot:right>
								<bt-label
									name="sym_r_download"
									:label="deviceStore.isMobile ? '' : t('Download Raw Log')"
									@click="downloadLogs"
								/>
							</template>
							<template v-slot:body> </template>
						</app-store-body>
					</div>
				</template>
			</page-container>
		</template>
	</adaptive-layout>
</template>

<script lang="ts" setup>
import AdaptiveLayout from '../../../components/settings/AdaptiveLayout.vue';
import PageContainer from '../../../components/base/PageContainer.vue';
import AppStoreBody from '../../../components/base/AppStoreBody.vue';
import EmptyView from '../../../components/base/EmptyView.vue';
import TitleBar from '../../../components/base/TitleBar.vue';
import BtLabel from '../../../components/base/BtLabel.vue';
import { marketLogs } from '../../../api/market/private/operations';
import { useDeviceStore } from '../../../stores/settings/device';
import { useTerminusStore } from '../../../stores/terminus';
import { onMounted, ref, computed } from 'vue';
import { bus } from '../../../utils/bus';
import { useRouter } from 'vue-router';
import { saveAs } from 'file-saver';
import { useI18n } from 'vue-i18n';
import { date } from 'quasar';

const { t } = useI18n();
const router = useRouter();
const deviceStore = useDeviceStore();

const columns: any = [
	{
		name: 'createTime',
		align: 'left',
		label: t('base.time'),
		field: 'time',
		format: (time: number) => formattedDate(time)
	},
	{
		name: 'account',
		align: 'left',
		label: t('account'),
		field: 'account'
	},
	{
		name: 'name',
		align: 'left',
		label: t('base.app'),
		field: 'app'
	},
	{
		name: 'type',
		align: 'left',
		label: t('base.type'),
		field: 'type'
	},
	{
		name: 'message',
		align: 'left',
		label: t('base.message'),
		field: 'message'
	}
];

const formattedDate = (datetime: number) => {
	const originalDate = new Date(datetime * 1000);
	return date.formatDate(originalDate, 'YYYY-MM-DD HH:mm:ss');
};

const pagination = ref({
	sortBy: 'desc',
	page: 1,
	rowsPerPage: 10
});

const pagesNumber = computed(() =>
	Math.ceil(rows.value.length / pagination.value.rowsPerPage)
);

const loading = ref(false);
const rows = ref([]);

onMounted(async () => {
	loading.value = true;
	marketLogs(200)
		.then((data) => {
			console.log(data);
			if (data) {
				rows.value = data.records.map((item) => {
					return { ...item, expand: !!item.extended };
				});
				console.log(rows.value);
			}
		})
		.finally(() => {
			loading.value = false;
		});
});

const coverJSONData = (extended: string, indent = 4): string => {
	if (!extended) {
		return '';
	}
	try {
		const parsed = JSON.parse(extended);
		const processed = processNestedObjects(parsed);
		return JSON.stringify(processed, null, indent);
	} catch (error) {
		console.log('json:', extended);
		console.error('JSON parse error:', error);
		return extended;
	}
};

const processNestedObjects = (obj: any): any => {
	if (obj === null || typeof obj !== 'object') {
		return obj;
	}

	if (Array.isArray(obj)) {
		return obj.map((item) => processNestedObjects(item));
	}

	const result: Record<string, any> = {};
	for (const key in obj) {
		if (Object.prototype.hasOwnProperty.call(obj, key)) {
			const value = obj[key];
			if (typeof value === 'string') {
				try {
					const parsedValue = JSON.parse(value);
					result[key] = processNestedObjects(parsedValue);
				} catch {
					result[key] = value;
				}
			} else {
				result[key] = processNestedObjects(value);
			}
		}
	}
	return result;
};

const downloadLogs = async () => {
	marketLogs(500)
		.then((result) => {
			console.log(result);
			if (result && result.records && result.records.length > 0) {
				const terminusStore = useTerminusStore();
				const logContent = JSON.stringify(result.records, null, 2);
				const blob = new Blob([logContent], {
					type: 'text/plain;charset=utf-8'
				});
				saveAs(
					blob,
					`${terminusStore.olaresId}-${Date.now()}-market-frontend.log`
				);
			} else {
				bus.emit('app_backend_error', 'get log records error');
			}
		})
		.catch((err) => {
			bus.emit('app_backend_error', err.message || `get logs failure ${err}`);
		});
};

const openNewLink = (url: string) => {
	window.open(url, '_blank');
};
</script>

<style scoped lang="scss">
.log-scroll {
	width: 100%;
	height: 100%;
	padding: 0 var(--paddingX);

	.log-message {
		padding-left: 4px;
		max-width: 300px !important;
		white-space: pre-line;
		overflow: hidden;
		text-overflow: ellipsis;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
	}

	.log-extend {
		max-width: 100%;
		overflow: auto;
		white-space: pre-wrap;
		word-wrap: break-word;
		word-break: break-all;
	}

	.empty-view {
		width: 100%;
		height: 600px;
	}

	.log-item {
		margin-bottom: 12px;
		border-radius: 12px;
		padding: 12px 20px;
	}
}

::v-deep(.q-btn.text-grey-8:before) {
	border: unset !important;
}
</style>
