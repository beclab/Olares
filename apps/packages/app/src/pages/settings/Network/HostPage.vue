<template>
	<page-title-component :show-back="true" :title="t('hosts')">
		<template v-slot:end>
			<div
				v-if="rows.length > 0"
				class="add-btn row justify-center items-center"
				@click="editHost(-1)"
			>
				<q-icon size="20px" name="sym_r_add" color="ink-1" />
				<div class="text-body3 add-title">{{ t('add_hosts') }}</div>
			</div>
		</template>
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<AdaptiveLayout>
			<template v-slot:pc>
				<app-menu-empty
					v-if="rows.length === 0"
					:title="t('Flexible Hosts Management')"
					:button-label="t('Add hosts')"
					:message="
						t(
							'Add and manage custom hosts rules to easily optimize your network access and DNS resolution.'
						)
					"
					image="settings/imgs/root/host.svg"
					@click="editHost(-1)"
				/>
				<div v-else class="q-list-class q-pa-md column">
					<q-table
						:rows="currentPageData"
						flat
						:columns="columns"
						row-key="host"
						:loading="loading2"
						@request="onRequest"
						v-model:pagination="pagination"
						tableHeaderStyle="height: 32px;"
						table-header-class="text-body3 text-ink-2"
						:bordered="false"
						hide-pagination
						hide-selected-banner
						hide-bottom
					>
						<template v-slot:body-cell-action="props">
							<q-td :props="props">
								<div class="row justify-end items-center">
									<q-btn
										class="q-mr-xs btn-size-sm btn-no-text btn-no-border"
										icon="sym_r_edit_square"
										color="ink-2"
										outline
										@click.stop="editHost(props.rowIndex)"
										no-caps
									/>
									<q-btn
										class="btn-size-sm btn-no-text btn-no-border"
										icon="sym_r_delete"
										color="ink-2"
										outline
										@click.stop
										no-caps
										@click="deleteHost(props.rowIndex)"
									/>
								</div>
							</q-td>
						</template>
					</q-table>
					<div class="row justify-center">
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
				<q-infinite-scroll @load="onLoad" :offset="250">
					<app-menu-empty
						v-if="rows.length === 0"
						:title="t('Flexible Hosts Management')"
						:button-label="t('Add hosts')"
						:message="
							t(
								'Add and manage custom hosts rules to easily optimize your network access and DNS resolution.'
							)
						"
						image="settings/imgs/root/host.svg"
						@click="editHost(-1)"
					/>
					<div
						class="mobile-items-list row justify-between q-px-lg q-py-md"
						v-for="(item, index) in rows"
						:key="item.host"
					>
						<div class="column justify-start">
							<div class="text-body2 text-ink-1">
								{{ item.host }}
							</div>
							<div class="text-overline text-ink-3 q-mt-xs">
								{{ item.ip }}
							</div>
						</div>

						<div class="row justify-end items-center">
							<q-btn
								class="q-mr-xs btn-size-sm btn-no-text btn-no-border"
								icon="sym_r_edit_square"
								color="ink-2"
								outline
								@click.stop="editHost(index)"
								no-caps
							/>
							<q-btn
								class="btn-size-sm btn-no-text btn-no-border"
								icon="sym_r_delete"
								color="ink-2"
								outline
								@click.stop
								no-caps
								@click="deleteHost(index)"
							/>
						</div>
					</div>
				</q-infinite-scroll>
			</template>
		</AdaptiveLayout>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import EditHostDialog from './EditHostDialog.vue';
import { useTerminusDStore } from 'src/stores/settings/terminusd';
import { useDeviceStore } from 'src/stores/settings/device';
import { Loading, useQuasar } from 'quasar';
import { HostItem, MENU_TYPE } from 'src/constant';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';

const { t } = useI18n();
const $q = useQuasar();

const columns: any = [
	{
		name: 'host',
		align: 'left',
		label: t('host_name'),
		field: 'host'
	},
	{
		name: 'ip',
		align: 'left',
		label: t('ip'),
		field: 'ip'
	},
	{
		name: 'action',
		align: 'right',
		label: t('action')
	}
];

const pagination = ref({
	sortBy: 'desc',
	page: 1,
	rowsPerPage: 10,
	rowsNumber: 0
});

const loading2 = ref(false);
const rows = ref<HostItem[]>([]);
const terminusDStore = useTerminusDStore();

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
	// const rowsPerPageFormat = rowsPerPage ? rowsPerPage : undefined;
	// const params = {
	// 	page: page,
	// 	limit: rowsPerPageFormat
	// };
	pagination.value.page = page;
	pagination.value.rowsPerPage = rowsPerPage;
	fetchHistory();
};

const fetchLoginRecords = async () => {
	// const params = {
	// 	page: pagination.value.page,
	// 	limit: pagination.value.rowsPerPage
	// };
	await fetchHistory();
};

const deviceStore = useDeviceStore();

onMounted(async () => {
	if (!deviceStore.isMobile) {
		fetchLoginRecords();
	}
});

const fetchHistory = async () => {
	loading2.value = true;
	const res = await terminusDStore.getHostsList();
	loading2.value = false;
	if (res && res.length > 0) {
		rows.value = res.map((item: any) => ({
			ip: item.ip,
			host: item.host
		}));
	}

	pagination.value.rowsNumber = res ? res.length : 0;
};

const currentPageData = computed(() => {
	const start = (pagination.value.page - 1) * pagination.value.rowsPerPage;
	const end = start + pagination.value.rowsPerPage;
	return rows.value.slice(start, end);
});

const onLoad = async (index: number, done: (stop?: boolean) => void) => {
	pagination.value.page = index;
	try {
		await fetchLoginRecords();
		done(pagination.value.rowsNumber == rows.value.length);
	} catch (error) {
		done(pagination.value.rowsNumber == rows.value.length);
	}
};

const pagesNumber = computed(() =>
	Math.ceil(rows.value.length / pagination.value.rowsPerPage)
);

const editHost = (index: number) => {
	$q.dialog({
		component: EditHostDialog,
		componentProps: {
			itemList: rows.value,
			index: index
		}
	}).onOk((list: HostItem[]) => {
		rows.value = list;
		console.log(list);
	});
};

const deleteHost = (index: number) => {
	Loading.show();
	try {
		const list = rows.value.filter((_, i) => index !== i);
		terminusDStore.updateHostsList(list);
		rows.value = list;
	} catch (e) {
		console.log(e);
	} finally {
		Loading.hide();
	}
};
</script>

<style scoped lang="scss">
.add-btn {
	border-radius: 8px;
	border: 1px solid $separator;
	cursor: pointer;
	padding: 5px 8px;
	text-decoration: none;

	.add-title {
		color: $ink-2;
	}
}

.add-btn-padding-mobile {
	padding: 8px 10px;
}

.add-btn:hover {
	background-color: $background-3;
}
</style>
