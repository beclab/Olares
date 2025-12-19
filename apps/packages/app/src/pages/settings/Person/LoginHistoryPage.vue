<template>
	<page-title-component :show-back="true" :title="t('login_history')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<AdaptiveLayout>
			<template v-slot:pc>
				<div
					v-if="userInfo?.wizard_complete === true"
					class="q-list-class q-pa-md column q-mt-md"
				>
					<q-table
						:rows="rows"
						flat
						:columns="columns"
						row-key="name"
						:loading="loading2"
						@request="onRequest"
						v-model:pagination="pagination"
						style="margin-top: 12px"
						tableHeaderStyle="height: 32px;"
						table-header-class="text-body3 text-ink-2"
						:bordered="false"
						hide-pagination
						hide-selected-banner
						hide-bottom
					>
						<template v-slot:body-cell-status="props">
							<q-td :props="props">
								<div class="row items-center">
									<div
										class="login-token-base-status q-mr-sm"
										:class="props.row.status ? 'bg-info' : 'bg-red'"
									></div>
									<span>{{
										props.row.status ? t('successful') : t('failed')
									}}</span>
								</div>
							</q-td>
						</template>

						<template v-slot:body-cell-reason="props">
							<q-td :props="props">
								<div
									style="
										max-width: 200px;
										word-wrap: break-word;
										white-space: normal;
									"
								>
									{{ props.row.reason }}
								</div>
							</q-td>
						</template>
					</q-table>
					<div class="row justify-center q-mt-md">
						<q-pagination
							v-model="pagination.page"
							:max="pagesNumber"
							input
							icon-first="sym_r_keyboard_double_arrow_left"
							icon-last="sym_r_keyboard_double_arrow_right"
							icon-next="sym_r_keyboard_arrow_right"
							icon-prev="sym_r_keyboard_arrow_left"
							color="ink-3"
							input-class="text-ink-2 text-body-3"
						/>
					</div>
				</div>
			</template>
			<template v-slot:mobile>
				<!-- <q-infinite-scroll @load="onLoad" :offset="250"> -->
				<bt-grid
					class="mobile-items-list q-pb-lg q-mt-md"
					:repeat-count="2"
					v-for="(token, index) in rows"
					:key="index"
					:paddingY="12"
				>
					<template v-slot:title>
						<div
							class="text-subtitle3-m row justify-between items-center clickable-view q-mb-md"
						>
							<div>{{ token.source_ip }}</div>
							<div class="row items-center justify-end">
								<div
									class="login-token-base-status"
									:class="token.status ? 'bg-info' : 'bg-red'"
								></div>
								<div class="q-ml-xs text-body3-m text-ink-1">
									{{ token.status ? t('successful') : t('failed') }}
								</div>
							</div>
						</div>
					</template>
					<template v-slot:grid>
						<bt-grid-item
							:label="t('time.time')"
							mobileTitleClasses="text-body3-m"
							:value="getLocalTime(token.time).format('YYYY-MM-DD HH:mm')"
						/>
						<bt-grid-item
							:label="t('reason')"
							mobileTitleClasses="text-body3-m"
							:value="token.reason"
						/>
					</template>
				</bt-grid>
				<!-- <template v-slot:loading>
						<div class="row justify-center q-my-md">
							<q-spinner-dots color="primary" size="40px" />
						</div>
					</template> -->
				<!-- </q-infinite-scroll> -->
			</template>
		</AdaptiveLayout>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import BtGridItem from 'src/components/settings/base/BtGridItem.vue';
import BtGrid from 'src/components/settings/base/BtGrid.vue';
import { AccountInfo, UsersParam } from 'src/constant/global';
import { getLocalTime } from 'src/utils/settings/utils';
import { useUserStore } from 'src/stores/settings/user';
import { useDeviceStore } from 'src/stores/settings/device';
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const columns: any = [
	{
		name: 'createTime',
		align: 'left',
		label: t('time.time'),
		field: 'createTime',
		format: (time: string) => getLocalTime(time).format('YYYY-MM-DD HH:mm:ss')
	},
	{
		name: 'status',
		align: 'left',
		label: t('status'),
		field: (row: any) => row.name
	},
	{
		name: 'source_ip',
		align: 'left',
		label: t('source_ip_address'),
		field: 'source_ip'
	},
	{
		name: 'reason',
		align: 'right',
		label: t('reason'),
		field: 'reason'
	}
];

const pagination = ref({
	sortBy: 'desc',
	page: 1,
	rowsPerPage: 10,
	rowsNumber: 0
});

const loading2 = ref(false);
const rows = ref<any[]>([]);

const onRequest = (props: {
	pagination: {
		sortBy: string;
		page: number;
		rowsPerPage: number;
	};
	filter?: any;
	getCellValue: (col: any, row: any) => any;
}) => {
	const { page, rowsPerPage } = props.pagination;
	const rowsPerPageFormat = rowsPerPage ? rowsPerPage : undefined;
	const params = {
		page: page,
		limit: rowsPerPageFormat,
		user: userInfo.value?.name
	};
	pagination.value.page = page;
	pagination.value.rowsPerPage = rowsPerPage;
	fetchHistory(params);
};

const fetchLoginRecords = async (name: string) => {
	const params = {
		page: pagination.value.page,
		limit: pagination.value.rowsPerPage,
		user: name
	};
	await fetchHistory(params);
};

const accountStore = useUserStore();
const Route = useRoute();
const deviceStore = useDeviceStore();

const userInfo = ref<AccountInfo | undefined>(
	accountStore.getUserByName(Route.params.name as string)
);

onMounted(async () => {
	userInfo.value = await accountStore.get_account_info(
		Route.params.name as string
	);
	// if (!deviceStore.isMobile) {
	fetchLoginRecords(Route.params.name as string);
	// }
});

const fetchHistory = async (params: UsersParam) => {
	loading2.value = true;
	const res = await accountStore.getLoginrecords(params);
	loading2.value = false;
	if (deviceStore.isMobile) {
		rows.value = rows.value.concat(
			res.items.map((item: any) => ({
				createTime:
					String(item.login_time).length > 10
						? item.login_time
						: item.login_time * 1000,
				status: item.success,
				source_ip: item.sourceIP,
				reason: item.reason
			}))
		);
	} else {
		rows.value = res.items.map((item: any) => ({
			createTime:
				String(item.login_time).length > 10
					? item.login_time
					: item.login_time * 1000,
			status: item.success,
			source_ip: item.sourceIP,
			reason: item.reason
		}));
	}

	pagination.value.rowsNumber = res.totalItems;
};

// const onLoad = async (index: number, done: (stop?: boolean) => void) => {
// 	pagination.value.page = index;
// 	try {
// 		await fetchLoginRecords();
// 		done(pagination.value.rowsNumber == rows.value.length);
// 	} catch (error) {
// 		done(pagination.value.rowsNumber == rows.value.length);
// 	}
// };

const pagesNumber = computed(() =>
	Math.ceil(rows.value.length / pagination.value.rowsPerPage)
);
</script>

<style scoped lang="scss">
.login-token-base-status {
	width: 8px;
	height: 8px;
	border-radius: 4px;
}
</style>
