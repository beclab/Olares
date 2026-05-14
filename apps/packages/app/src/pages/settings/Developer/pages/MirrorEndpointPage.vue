<template>
	<page-title-component :show-back="true" :title="t('Mirror management')">
		<template v-slot:end>
			<div
				v-if="deviceStore.isMobile"
				class="row justify-center items-center"
				@click="addRegistry()"
			>
				<q-icon name="sym_r_add" color="ink-1" size="32px" />
				<div class="text-body3 add-title" v-if="!deviceStore.isMobile">
					{{ t('Add mirror') }}
				</div>
			</div>
		</template>
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<AdaptiveLayout>
			<template v-slot:pc>
				<bt-list>
					<bt-form-item
						:title="t('Repo name')"
						:margin-top="false"
						:chevron-right="false"
						:data="registry"
						:widthSeparator="false"
					/>
				</bt-list>
				<q-list class="q-mt-lg q-py-md q-list-class">
					<div
						v-if="endpoints && endpoints.length > 0"
						class="column item-margin-left item-margin-right"
					>
						<q-table
							tableHeaderStyle="height: 32px;"
							table-header-class="text-body3 text-ink-3"
							flat
							:bordered="false"
							:rows="endpoints"
							:columns="columns"
							row-key="id"
							hide-pagination
							hide-selected-banner
							hide-bottom
							wrap-cells
							:rowsPerPageOptions="[0]"
						>
							<template v-slot:body-cell-url="props">
								<q-td
									:props="props"
									class="text-ink-1 text-body1 ellipsis"
									style="max-width: 220px"
								>
									{{ props.row }}
								</q-td>
							</template>
							<template v-slot:body-cell-sorting="props">
								<q-td :props="props" class="text-ink-1 text-body1 items-center">
									<!-- {{ props.row }} -->
									<div class="row" style="min-width: 40px">
										<bt-action-icon
											name="sym_r_keyboard_arrow_up"
											:icon-size="16"
											:size="20"
											@click.stop="toFirst(props.row)"
										/>
										<bt-action-icon
											name="sym_r_keyboard_arrow_down"
											:icon-size="16"
											:size="20"
											@click.stop="toLast(props.row)"
										/>
									</div>
								</q-td>
							</template>
							<template v-slot:body-cell-actions="props">
								<q-td :props="props" class="text-ink-1 text-body1">
									<div class="row items-center justify-end">
										<bt-action-icon
											name="sym_r_delete"
											:icon-size="20"
											@click.stop="removeConfirm(props.row)"
										/>
									</div>
								</q-td>
							</template>
						</q-table>
					</div>
					<empty-component
						class="q-pb-xl"
						v-else
						:info="t('No endpoint added')"
						:empty-image-top="40"
					/>
				</q-list>

				<div class="row justify-end items-center q-my-lg">
					<div
						class="add-btn row justify-end items-center"
						@click="addRegistry()"
					>
						<q-icon name="sym_r_add" color="ink-1" size="20px" />
						<div class="text-body3 add-title" v-if="!deviceStore.isMobile">
							{{ t('Add mirror') }}
						</div>
					</div>
				</div>
			</template>
			<template v-slot:mobile>
				<div v-if="endpoints && endpoints.length > 0">
					<bt-grid
						class="mobile-items-list"
						:repeat-count="2"
						v-for="(endpoint, index) in endpoints"
						:key="index"
						:paddingY="12"
					>
						<template v-slot:title>
							<div
								class="text-subtitle3-m row justify-between items-center clickable-view q-mb-md"
							>
								<div>
									{{ endpoint }}
								</div>
								<q-icon
									name="sym_r_delete"
									color="ink-2"
									size="20px"
									@click.stop="removeConfirm(endpoint)"
								/>
							</div>
						</template>
						<template v-slot:grid>
							<bt-grid-item mobileTitleClasses="text-body3-m">
								<template v-slot:value>
									<q-btn
										class="btn-size-lg"
										icon="sym_r_arrow_upward_alt"
										text-color="ink-1"
										:label="t('Move to top')"
										@click="toFirst(endpoint)"
									>
									</q-btn>
								</template>
							</bt-grid-item>
							<bt-grid-item mobileTitleClasses="text-body3-m">
								<template v-slot:value>
									<q-btn
										class="btn-size-lg"
										icon="sym_r_arrow_downward_alt"
										text-color="ink-1"
										:label="t('Move to bottom')"
										@click="toLast(endpoint)"
									>
									</q-btn>
								</template>
							</bt-grid-item>
						</template>
					</bt-grid>
				</div>
				<empty-component
					class="q-pb-xl"
					v-else
					:info="t('No endpoint added')"
					:empty-image-top="40"
				/>
			</template>
		</AdaptiveLayout>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import ReminderDialogComponent from '../../../../components/settings/ReminderDialogComponent.vue';
import EditMirrorDialog from 'src/pages/settings/Developer/pages/dialog/EditMirrorDialog.vue';
import BtActionIcon from '../../../../components/settings/base/BtActionIcon.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtGridItem from '../../../../components/settings/base/BtGridItem.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import BtGrid from 'src/components/settings/base/BtGrid.vue';

import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { useDeviceStore } from 'src/stores/settings/device';
import { useMirrorStore } from 'src/stores/settings/mirror';
import { useRoute, useRouter } from 'vue-router';
import { onMounted, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import BtList from 'src/components/settings/base/BtList.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import cloneDeep from 'lodash/cloneDeep';

const { t } = useI18n();
const $q = useQuasar();
const route = useRoute();
const router = useRouter();
const deviceStore = useDeviceStore();
const mirrorStore = useMirrorStore();
const endpoints = ref([] as string[]);
const registry = ref((route.query.registry as string) || '');

onMounted(async () => {
	endpoints.value = await mirrorStore.getRegistryEndpoint(registry.value);
});

const removeConfirm = (item: string) => {
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('Confirm deletion?'),
			message: t('Are you sure you want to delete image source {source}?', {
				source: item
			}),
			useCancel: true,
			confirmText: t('confirm'),
			cancelText: t('cancel')
		}
	}).onOk(async () => {
		const items = endpoints.value.filter((e) => e !== item);
		updateEndpoints(items);
	});
};

const toFirst = async (item: string) => {
	const items = endpoints.value.filter((e) => e !== item);
	updateEndpoints([item, ...items]);
};
const toLast = (item: string) => {
	const items = endpoints.value.filter((e) => e !== item);
	updateEndpoints([...items, item]);
};

const updateEndpoints = async (items: string[]) => {
	try {
		endpoints.value = await mirrorStore.putRegistryEndpoint(
			registry.value,
			items
		);
		notifySuccess(t('successful'));
		if (endpoints.value.length == 0) {
			setTimeout(() => {
				router.back();
			}, 1000);
		}
	} catch (error) {
		console.log(error);
	}
};

const columns: any = [
	{
		name: 'url',
		align: 'left',
		label: t('endpoint'),
		field: '',
		sortable: false
	},
	{
		name: 'sorting',
		align: 'left',
		label: t('Sorting'),
		field: '',
		sortable: false
	},
	{
		name: 'actions',
		align: 'right',
		label: t('action'),
		sortable: false
	}
];

const addRegistry = () => {
	$q.dialog({
		component: EditMirrorDialog,
		componentProps: {}
	}).onOk(async (data: { endpoint: string }) => {
		const currentEndpoints = endpoints.value || [];
		if (currentEndpoints.includes(data.endpoint)) {
			notifySuccess(t('successful'));
			return;
		}
		const newEndpoints = cloneDeep(currentEndpoints);
		newEndpoints.push(data.endpoint);
		try {
			endpoints.value = await mirrorStore.putRegistryEndpoint(
				registry.value,
				newEndpoints
			);
			notifySuccess(t('successful'));
		} catch (error) {
			// notifyFailed(error);
			if (
				error.response &&
				error.response.data &&
				error.response.data.message
			) {
				notifyFailed(error.response.data.message);
			} else {
				notifyFailed(error);
			}
		}
	});
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

::v-deep(.q-table tbody td) {
	font-size: 16px;
}
</style>
