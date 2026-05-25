<template>
	<page-title-component :show-back="true" :title="t('Repository management')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<AdaptiveLayout>
			<template v-slot:pc>
				<q-list class="q-py-md q-list-class q-mt-md">
					<div
						v-if="mirrorStore.registries && mirrorStore.registries.length > 0"
						class="column item-margin-left item-margin-right"
					>
						<q-table
							tableHeaderStyle="height: 32px;"
							table-header-class="text-body3 text-ink-2"
							flat
							:bordered="false"
							:rows="mirrorStore.registries"
							:columns="columns"
							row-key="id"
							hide-pagination
							hide-selected-banner
							hide-bottom
							:rowsPerPageOptions="[0]"
						>
							<template v-slot:body-cell-actions="props">
								<q-td
									:props="props"
									style="height: 64px"
									class="text-ink-1 text-body1 row items-center justify-end"
									no-hover
								>
									<!--									<bt-action-icon-->
									<!--										name="sym_r_box_edit"-->
									<!--										:icon-size="20"-->
									<!--										@click.stop="enterEndpoint(props.row)"-->
									<!--									/>-->
									<q-btn
										class="btn-size-xs btn-no-text text-grey-8"
										icon="sym_r_keyboard_arrow_right"
										text-color="ink-2"
										@click.stop="enterEndpoint(props.row)"
									/>
									<!-- <bt-action-icon
										name="sym_r_feed"
										:icon-size="15"
										@click.stop="enterImages(props.row)"
									/> -->
								</q-td>
							</template>
							<template v-slot:body-cell-port="props">
								<q-td
									class="text-ink-1 text-body1"
									style="height: 64px"
									:props="props"
									no-hover
								>
									{{ props.row.name }}
								</q-td>
							</template>

							<template v-slot:body-cell-count="props">
								<q-td
									class="text-ink-1 text-body1 cursor-pointer"
									style="height: 64px"
									:props="props"
									no-hover
									@click="enterImages(props.row)"
								>
									{{ props.row.image_count }}
								</q-td>
							</template>

							<template v-slot:body-cell-size="props">
								<q-td
									class="text-ink-1 text-body1"
									style="height: 64px"
									:props="props"
									no-hover
								>
									{{ format.humanStorageSize(props.row.image_size) }}
								</q-td>
							</template>
						</q-table>
					</div>
					<empty-component
						class="q-pb-xl"
						v-else
						:info="t('No image repository is used')"
						:empty-image-top="40"
					/>
				</q-list>
			</template>
			<template v-slot:mobile>
				<div>
					<bt-grid
						class="mobile-items-list"
						:repeat-count="2"
						v-for="(port, index) in mirrorStore.registries"
						:key="index"
						:paddingY="12"
					>
						<template v-slot:title>
							<div
								class="text-subtitle3-m row justify-between items-center clickable-view q-mb-md"
							>
								<div>
									{{ port.name }}
								</div>
								<div class="row items-center justify-end">
									<bt-action-icon
										name="sym_r_keyboard_arrow_right"
										:icon-size="15"
										@click.stop="enterEndpoint(port)"
									/>
									<!-- <bt-action-icon
										name="sym_r_feed"
										:icon-size="15"
										@click.stop="enterImages(port)"
									/> -->
								</div>
							</div>
						</template>
						<template v-slot:grid>
							<bt-grid-item
								:label="t('Image count')"
								mobileTitleClasses="text-body3-m"
								:value="port.image_count"
								@click="enterImages(port)"
							/>
							<bt-grid-item
								:label="t('Image size')"
								mobileTitleClasses="text-body3-m"
								:value="format.humanStorageSize(port.image_size)"
							/>
						</template>
					</bt-grid>
				</div>
			</template>
		</AdaptiveLayout>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtActionIcon from '../../../../components/settings/base/BtActionIcon.vue';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import BtGridItem from 'src/components/settings/base/BtGridItem.vue';
import BtGrid from 'src/components/settings/base/BtGrid.vue';

import { useMirrorStore, RegistryMirror } from 'src/stores/settings/mirror';
import { format } from 'src/utils/format';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { onMounted } from 'vue';

const { t } = useI18n();
const mirrorStore = useMirrorStore();
const router = useRouter();

onMounted(async () => {
	mirrorStore.getRegistryMirrors().then((mirrors) => {});
});

const enterEndpoint = (item: RegistryMirror) => {
	router.push({
		path: '/developer/mirror/endpoint',
		query: {
			registry: item.name
		}
	});
};

const enterImages = (item: RegistryMirror) => {
	router.push({
		path: '/developer/images',
		query: {
			registry: item.name
		}
	});
};

const columns: any = [
	{
		name: 'port',
		align: 'left',
		label: t('Repo name'),
		field: 'name',
		format: (val: any) => {
			return val;
		},
		sortable: false
	},
	{
		name: 'count',
		align: 'left',
		label: t('Image count'),
		field: 'image_count',
		sortable: false
	},
	{
		name: 'size',
		align: 'left',
		label: t('Image size'),
		field: 'image_size',
		format: (val: any) => {
			return format.humanStorageSize(val);
		},
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
