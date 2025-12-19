<template>
	<page-title-component :show-back="true" :title="t('Endpoint management')">
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<AdaptiveLayout>
			<template v-slot:pc>
				<q-list class="q-py-md q-list-class">
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
							:rowsPerPageOptions="[0]"
						>
							<template v-slot:body-cell-url="props">
								<q-td :props="props" class="text-ink-1 text-body1">
									{{ props.row }}
								</q-td>
							</template>
							<template v-slot:body-cell-sorting="props">
								<q-td :props="props" class="text-ink-1 text-body1 items-center">
									<!-- {{ props.row }} -->
									<div class="row">
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
								<q-td
									:props="props"
									class="text-ink-1 text-body1 row items-center justify-end"
								>
									<bt-action-icon
										name="sym_r_delete"
										:icon-size="20"
										@click.stop="removeConfirm(props.row)"
									/>
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
import { onMounted, ref } from 'vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { useI18n } from 'vue-i18n';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import BtGridItem from '../../../../components/settings/base/BtGridItem.vue';
import BtGrid from 'src/components/settings/base/BtGrid.vue';

import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import { useMirrorStore } from '../../../../stores/settings/mirror';
import BtActionIcon from '../../../../components/settings/base/BtActionIcon.vue';

import { useRoute, useRouter } from 'vue-router';
import ReminderDialogComponent from '../../../../components/settings/ReminderDialogComponent.vue';
import { useQuasar } from 'quasar';
import { notifySuccess } from '../../../../utils/settings/btNotify';

const { t } = useI18n();

const mirrorStore = useMirrorStore();

const route = useRoute();

const registry = ref((route.query.registry as string) || '');

const endpoints = ref([] as string[]);

const $q = useQuasar();

const router = useRouter();

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
</script>

<style scoped lang="scss">
::v-deep(.q-table tbody td) {
	font-size: 16px;
}
</style>
