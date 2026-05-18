<template>
	<page-title-component :show-back="true" :title="t('acls')" />

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<adaptive-layout>
			<template v-slot:pc>
				<q-list class="q-py-md q-list-class q-mt-md">
					<div
						v-if="rows && rows.length > 0"
						class="column item-margin-left item-margin-right"
					>
						<q-table
							flat
							tableHeaderStyle="height: 32px;"
							table-header-class="text-body3 text-ink-2"
							:bordered="false"
							:rows="rows"
							:columns="columns"
							row-key="Port"
							hide-pagination
							hide-selected-banner
							hide-bottom
							:rowsPerPageOptions="[0]"
						>
							<template v-slot:header="props">
								<q-th
									v-for="col in props.cols"
									:key="col.name"
									:props="props"
									class="text-body3 text-ink-2 q-py-sm"
								>
									{{ col.label }}
								</q-th>
							</template>
							<template v-slot:body-cell="props">
								<q-td :props="props">
									<div class="text-body1 text-ink-1">{{ props.value }}</div>
								</q-td>
							</template>
						</q-table>
					</div>
					<empty-component
						class="q-pb-xl"
						v-else
						:info="t('No image added')"
						:empty-image-top="40"
					/>
				</q-list>
			</template>
			<template v-slot:mobile>
				<template v-for="(item, index) in aclStore.appAclList" :key="index">
					<bt-list>
						<bt-form-item :title="t('dst')">
							<div class="dst-grid row justify-end items-center">
								<template v-for="dst in item.dst" :key="dst">
									<div class="acl-dst text-caption text-ink-2">
										{{ dst }}
									</div>
								</template>
							</div>
						</bt-form-item>
						<bt-form-item
							:width-separator="false"
							:title="t('protocol')"
							:data="item.proto"
						/>
					</bt-list>
				</template>
			</template>
		</adaptive-layout>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useAclStore } from 'src/stores/settings/acl';
import { format } from 'src/utils/format';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { computed, onMounted, ref, watch } from 'vue';

const { t } = useI18n();
const aclStore = useAclStore();
const route = useRoute();
const rows = ref([]);

onMounted(() => {
	if (route.params.name) {
		aclStore.getAppAclStatus(route.params.name as string);
	}
});

watch(
	() => aclStore.appAclList,
	() => {
		console.log(aclStore.appAclList);
		rows.value = [];
		aclStore.appAclList.forEach((item) => {
			const protocol = item.proto || 'all';
			item.dst.forEach((dstStr) => {
				const [destination, portStr] = dstStr.split(':');
				const ports = portStr.split(',');
				ports.forEach((port) => {
					rows.value.push({
						Protocol: protocol.toUpperCase(),
						Destination: destination,
						Port: port.trim()
					});
				});
			});
		});

		console.log(rows.value);
	}
);

const columns: any = [
	{
		name: 'protocol',
		align: 'left',
		label: t('Protocol'),
		field: 'Protocol'
	},
	{
		name: 'destination',
		align: 'left',
		label: t('Destination'),
		field: 'Destination'
	},
	{
		name: 'port',
		align: 'right',
		label: t('Port'),
		field: 'Port'
	}
];
</script>

<style scoped lang="scss">
.dst-grid {
	max-width: 100%;
	padding-top: 14px;
	padding-bottom: 14px;
	gap: 10px;
	text-align: right;

	.acl-dst {
		height: 20px;
		padding: 4px 12px;
		border-radius: 20px;
		border: 1px solid $separator;
	}
}
</style>
