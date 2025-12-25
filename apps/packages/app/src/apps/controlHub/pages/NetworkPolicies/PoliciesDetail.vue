<template>
	<MyContentPage>
		<template #extra>
			<div class="col-auto">
				<QButtonStyle v-permission>
					<q-btn
						dense
						flat
						icon="sym_r_preview"
						color="ink-2"
						@click="clickHandler"
					>
						<q-tooltip>
							<div style="white-space: nowrap">
								{{ $t('VIEW_YAML') }}
							</div>
						</q-tooltip>
					</q-btn>
				</QButtonStyle>
			</div>
		</template>
		<MyPage2>
			<template v-for="(item, index) in tabs" :key="index">
				<MyCar4 square flat :title="item.name">
					<QTableStyle2 stickyLast>
						<q-table
							:rows="tableData(detail, item.value)"
							:columns="columns"
							row-key="name"
							hide-pagination
							flat
							:loading="
								item.value === 'ingress' ? ingress_loading : egress_loading
							"
							:pagination="pagination"
							@refresh="fetchData(true, false)"
						>
							<template v-slot:header="props">
								<q-tr :props="props">
									<q-th
										v-for="item in columns"
										:key="item.name"
										:props="props"
										>{{ item.label }}</q-th
									>
								</q-tr>
							</template>
							<template v-slot:body-cell-specPodSelector="props">
								<q-td :props="props">
									<span v-if="isEmpty(props.value)">-</span>
									<Labels v-else :data="props.value" no-empty></Labels>
								</q-td>
							</template>
							<template v-slot:body-cell-separator="props">
								<q-td :props="props">
									<span
										class="text-overline text-grey-5 row items-center justify-center"
										>→</span
									>
								</q-td>
							</template>

							<template v-slot:body-cell-namespaceSelector="props">
								<q-td :props="props">
									<div v-if="isEmpty(get(props.row, 'ipBlock'))">
										<div v-if="get(props.value, 'matchLabels')">
											<Labels :data="props.value.matchLabels"></Labels>
										</div>
										<div v-else>
											{{ get(props.row, 'namespace') }}
										</div>
									</div>
								</q-td>
							</template>
							<template v-slot:body-cell-podSelector="props">
								<q-td :props="props">
									<div v-if="isEmpty(get(props.row, 'ipBlock'))">
										<div v-if="get(props.row, 'matchLabels')">
											<Labels :data="props.value.matchLabels"></Labels>
										</div>
										<div v-else>Any</div>
									</div>
								</q-td>
							</template>
							<template v-slot:body-cell-ipBlock="props">
								<q-td :props="props">
									<span v-if="isEmpty(props.value)">-</span>
									<span v-else>{{ renderIpBlock(props.value) }}</span>
								</q-td>
							</template>
							<template v-slot:body-cell-ports="props">
								<q-td :props="props">
									<div v-if="isEmpty(props.value)">Any</div>
									<div v-else>
										<div v-for="(port, index) in props.value" :key="index">
											{{ port.protocol }}: {{ port.port }}
										</div>
									</div>
								</q-td>
							</template>
							<template #loading>
								<MyLoading2> </MyLoading2>
							</template>
							<template #no-data>
								<div class="full-width relative-position" style="height: 280px">
									<Empty3
										v-show="
											!(item.value === 'ingress'
												? ingress_loading
												: egress_loading)
										"
										@click="fetchData(false, true)"
									></Empty3>
								</div>
							</template>
						</q-table>
					</QTableStyle2>
				</MyCar4>
			</template>
		</MyPage2>
	</MyContentPage>
	<Yaml3
		ref="yamlRef"
		:title="t('VIEW_YAML')"
		:name="detail.name"
		:namespace="detail.namespace"
		module="networkpolicies"
		readonly
	>
	</Yaml3>
</template>

<script setup lang="ts">
import { get, isEmpty, toPairs } from 'lodash-es';
import { t } from '@apps/control-hub/src/boot/i18n';
import { getNetworkpoliciesDetail } from '@apps/control-hub/src/network';
import { NetworkPoliciesMapper } from '@apps/control-hub/src/utils/object.mapper';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import Labels from '@apps/control-panel-common/src/containers/Labels.vue';
import MyPage2 from '@apps/control-panel-common/src/containers/MyPage2.vue';
import Empty3 from '@apps/control-panel-common/src/components/Empty3.vue';
import MyContentPage from '@apps/control-hub/src/components/MyContentPage.vue';
import MyCar4 from '@apps/control-panel-common/src/components/MyCard2.vue';
import QTableStyle2 from '@apps/control-panel-common/src/components/QTableStyle2.vue';
import MyLoading2 from '@apps/control-panel-common/src/components/MyLoading2.vue';
import Yaml3 from '../NamespacePods/Yaml3.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';

const tabs = [
	{ name: t('INGRESS_RULES'), value: 'ingress' },
	{ name: t('EGRESS_RULES'), value: 'egress' }
];

const detail = ref<any>({});
const ingress_loading = ref(false);
const egress_loading = ref(false);
const route = useRoute();
const pagination = ref({
	rowsNumber: 0
});
const yamlRef = ref();
const tableData = (detail: any, direction: string) => {
	const originData = get(detail, '_originData');
	const namespace = get(originData, 'metadata.namespace');
	const specPodSelectorData = get(originData, 'spec.podSelector');
	const specPodSelector = toPairs(
		get(specPodSelectorData, 'matchLabels', [])
	).map((kv) => ({ [kv[0]]: kv[1] }));
	const directions = get(originData, `spec.${direction}`, []);
	const fromto = direction === 'ingress' ? 'from' : 'to';
	const data: any = [];

	directions.forEach((direction: any) => {
		const ports = get(direction, 'ports', []);
		const rules = get(direction, `${fromto}`, []);
		rules.forEach((rule: any) => {
			data.push({
				namespace,
				specPodSelector,
				ports,
				direction: direction,
				...rule
			});
		});
	});
	return data;
};

const columns: any = [
	{
		name: 'namespace',
		required: true,
		label: t('TARGET_PROJECT'),
		align: 'left',
		field: 'namespace'
	},
	{
		name: 'specPodSelector',
		align: 'left',
		label: t('LABEL'),
		field: 'specPodSelector'
	},
	{
		name: 'separator',
		align: 'left',
		label: '',
		field: 'namespace'
	},
	{
		name: 'namespaceSelector',
		label: t('DESTINATION_PROJECT'),
		align: 'left',
		field: 'namespaceSelector'
	},
	{
		name: 'podSelector',
		label: t('LABEL'),
		align: 'left',
		field: 'podSelector'
	},
	{
		name: 'ipBlock',
		label: t('CIDR'),
		align: 'left',
		field: 'ipBlock'
	},
	{
		name: 'ports',
		label: t('PORT'),
		align: 'left',
		field: 'ports'
	}
];

const renderIpBlock = (data: any) =>
	toPairs(data)
		.map((kv) => (kv[0] === 'cidr' ? `${kv[1]}` : kv.join(': ')))
		.reduce((acc: any, x: any) => (acc === null ? x : `${acc} ${x}`), null);

const fetchData = (ingress_loading_show = true, egress_loading_show = true) => {
	const { namespace, name }: { [key: string]: any } = route.params;
	const params = {
		namespace,
		name
	};
	if (ingress_loading_show) {
		ingress_loading.value = true;
	}
	if (egress_loading_show) {
		egress_loading.value = true;
	}
	getNetworkpoliciesDetail(params)
		.then((res) => {
			detail.value = NetworkPoliciesMapper(res.data);
		})
		.finally(() => {
			ingress_loading.value = false;
			egress_loading.value = false;
		});
};

const clickHandler = () => {
	yamlRef.value.show();
};

watch(
	() => route.params.uid,
	() => {
		fetchData();
	}
);

onMounted(() => {
	fetchData();
});
</script>
