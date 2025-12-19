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
			<my-card no-content-gap square flat animated>
				<template #title>
					<div>{{ t('DATA') }}</div>
				</template>
				<template #extra>
					<QButtonStyle size="sm">
						<q-btn
							color="grey-5"
							flat
							dense
							no-caps
							size="sm"
							:icon="
								secretValueVisible ? 'sym_r_visibility_off' : 'sym_r_visibility'
							"
							@click="secretValueVisibleHandler"
						>
						</q-btn>
					</QButtonStyle>
				</template>
				<DataDetail :data="secretObj"> </DataDetail>
			</my-card>
			<q-inner-loading :showing="loading"> </q-inner-loading>
		</MyPage>
	</MyContentPage>
	<Yaml
		ref="yamlRef"
		:title="t('EDIT_YAML')"
		module="secrets"
		:readonly="isStudio2"
	></Yaml>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import { computed, ref, watch } from 'vue';
import { getSecretsData } from '@apps/control-hub/src/network';
import { ObjectMapper } from '@apps/control-hub/src/utils/object.mapper';
import { isEmpty } from 'lodash-es';
import DetailPage from '@apps/control-panel-common/src/containers/DetailPage.vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import { getLocalTime } from '@apps/control-hub/src/utils';
import { SECRET_TYPES } from '@apps/control-hub/src/utils/constants';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';
import MyContentPage from '@apps/control-hub/src/components/MyContentPage.vue';
import { safeBtoa } from '@apps/control-panel-common/src/utils/base64';
import DataDetail from '@apps/control-panel-common/src/containers/DataDetail.vue';
import Yaml from '@apps/control-hub/src/pages/NamespacePods/Yaml3.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import MyCardHeader from '@apps/control-hub/src/components/MyCardHeader.vue';
import { useIsStudio, useIsStudio2 } from '@apps/control-hub/src/stores/hook';
import { selectedNodes } from '../treeStore';
const isStudio = useIsStudio();
const isStudio2 = useIsStudio2();

let loading = ref(false);
const secretValueVisible = ref(false);
const detail = ref();
const route = useRoute();
const secretsData = ref<{ [key: string]: string }>({});
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
			name: t('TYPE'),
			// eslint-disable-next-line @typescript-eslint/ban-ts-comment
			//@ts-ignore
			value: t(SECRET_TYPES[detail.type] || detail.type)
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

const secretObj = computed(() => {
	if (!secretValueVisible.value) {
		const obj = {};
		for (const key in secretsData.value) {
			obj[key] = safeBtoa(secretsData.value[key]);
		}
		return obj;
	}
	return secretsData.value;
});

const fetchDetail = () => {
	const { name, namespace }: any = route.params;
	loading.value = true;
	secretsData.value = {};
	detail.value = [];
	getSecretsData({ name, namespace })
		.then((res) => {
			const result = ObjectMapper.secrets(res.data);
			secretsData.value = result.data;
			detail.value = getAttrs(result);
		})
		.finally(() => {
			loading.value = false;
		});
};

const clickHandler = () => {
	yamlRef.value.show();
};

const secretValueVisibleHandler = () => {
	secretValueVisible.value = !secretValueVisible.value;
};

watch(
	() => route.params.pods_uid,
	async () => {
		fetchDetail();
	},
	{
		immediate: true
	}
);
</script>
