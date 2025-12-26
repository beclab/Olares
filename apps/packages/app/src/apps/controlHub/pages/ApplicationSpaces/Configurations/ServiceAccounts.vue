<template>
	<MyContentPage>
		<template #extra>
			<div class="col-auto">
				<QButtonStyle v-permission>
					<q-btn dense flat icon="sym_r_edit_square" @click="clickHandler">
						<q-tooltip>
							<div style="white-space: nowrap">
								{{ $t('EDIT_YAML') }}
							</div>
						</q-tooltip>
					</q-btn>
				</QButtonStyle>
			</div>
		</template>
		<MyPage>
			<my-card square flat animated>
				<template #title>
					<MyCardHeader
						:title="isStudio ? $route.params.name : t('DETAILS')"
						:img="selectedNodes?.img"
					/>
				</template>
				<template #extra v-if="isStudio">
					<QButtonStyle v-permission>
						<q-btn
							dense
							flat
							size="16px"
							:icon="isStudio2 ? 'sym_r_preview' : 'sym_r_edit_square'"
							@click="clickHandler"
						>
							<q-tooltip>
								<div style="white-space: nowrap">
									{{ isStudio2 ? $t('VIEW_YAML') : $t('EDIT_YAML') }}
								</div>
							</q-tooltip>
						</q-btn>
					</QButtonStyle>
				</template>
				<DetailPage :data="detail"></DetailPage>
			</my-card>
			<Detail
				v-for="item in secrets"
				:key="item.name"
				:secret="item.name"
				yaml
				:serviceAccountName="serviceAccountName"
			></Detail>
			<q-inner-loading :showing="loading"> </q-inner-loading>
		</MyPage>
		<Yaml
			ref="yamlRef"
			:title="t('EDIT_YAML')"
			module="serviceaccounts"
			:readonly="isStudio2"
		></Yaml>
	</MyContentPage>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import { onMounted, ref, watch } from 'vue';
import { getServiceaccountsItem } from '@apps/control-hub/src/network';
import { ObjectMapper } from '@apps/control-hub/src/utils/object.mapper';
import Detail from './Detail.vue';
import { get, isEmpty } from 'lodash-es';
import DetailPage from '@apps/control-panel-common/src/containers/DetailPage.vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import { getLocalTime } from '@apps/control-hub/src/utils';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';
import MyContentPage from '@apps/control-hub/src/components/MyContentPage.vue';
import Yaml from '@apps/control-hub/src/pages/NamespacePods/Yaml3.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import { useIsStudio, useIsStudio2 } from '@apps/control-hub/src/stores/hook';
import MyCardHeader from '@apps/control-hub/src/components/MyCardHeader.vue';
import { selectedNodes } from '../treeStore';
const isStudio = useIsStudio();
const isStudio2 = useIsStudio2();
const loading = ref(false);
const secrets = ref();
const data = ref();
const detail = ref();
const route = useRoute();
const serviceAccountName = ref();
const yamlRef = ref();

const getAttrs = (detail: any) => {
	const { cluster, namespace } = route.params;

	if (isEmpty(detail)) {
		return;
	}

	return [
		{
			name: t('CLUSTER'),
			value: cluster
		},
		{
			name: t('PROJECT'),
			value: namespace
		},
		{
			name: t('ROLE'),
			value: detail.role
		},
		{
			name: t('CREATION_TIME_TCAP'),
			value: getLocalTime(detail.createTime).format('YYYY-MM-DD HH:mm:ss')
		},
		{
			name: t('CREATOR'),
			value: detail.creator
		}
	];
};

const fetchList = () => {
	const { namespace, name }: any = route.params;
	detail.value = [];
	data.value = {};
	loading.value = true;
	getServiceaccountsItem(namespace, name)
		.then((res) => {
			secrets.value = get(res.data, 'secrets', []);

			const result = ObjectMapper.serviceaccounts(res.data);
			data.value = result;
			detail.value = getAttrs(result);
			serviceAccountName.value = result.name;
		})
		.finally(() => {
			loading.value = false;
		});
};

const clickHandler = () => {
	yamlRef.value.show();
};

watch(
	() => route.params.pods_uid,
	async () => {
		fetchList();
	},
	{
		immediate: true
	}
);
</script>
