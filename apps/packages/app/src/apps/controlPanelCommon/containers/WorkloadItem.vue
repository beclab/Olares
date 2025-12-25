<template>
	<div
		class="workload-item-container row items-center q-col-gutter-x-md no-wrap q-pa-lg"
	>
		<div class="col-4 q-pa-none">
			<ListItem size="sm" icon-size="lg">
				<template #avatar>
					<q-img :src="deploymentsIcon" width="32px"> </q-img>
				</template>
				<template #title>
					<!-- <a class="workload-item-link" :href="pathFormat(detail)">
            {{ getDisplayName(detail) }}
          </a> -->

					<a class="workload-item-link" @click="pathFormat(detail)">
						{{ getDisplayName(detail) }}
					</a>
				</template>
				<template #subTitle>
					<WorkloadsDescription :detail="detail"></WorkloadsDescription>
				</template>
			</ListItem>
		</div>
		<div class="col-4">
			<ListItem size="sm" icon-size="lg">
				<template #title>
					<WorkloadStatus
						:data="detail"
						:module="detail.module"
					></WorkloadStatus>
				</template>
				<template #subTitle>
					{{ t('STATUS') }}
				</template>
			</ListItem>
		</div>
		<div class="col-4">
			<ListItem
				size="sm"
				icon-size="lg"
				:title="`#${currentRevision}`"
				:sub-title="t('REVISION_RECORD')"
				class="relative-position"
				v-if="!loading"
			>
			</ListItem>
		</div>
		<SimpleLoading :showing="loading"> </SimpleLoading>
	</div>
</template>

<script setup lang="ts">
import { defineProps, computed, ref } from 'vue';
import { get } from 'lodash';
import { MODULE_KIND_MAP } from '@apps/control-panel-common/src/utils/constants';
import { useRoute, useRouter } from 'vue-router';
import { getRevisions } from '@apps/control-panel-common/src/network';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import { getDisplayName } from '@apps/control-panel-common/src/utils';
import WorkloadStatus from './WorkloadStatus.vue';
import WorkloadsDescription from './WorkloadsDescription.vue';
import { getCurrentRevision } from '../utils/workload';
import SimpleLoading from '../components/SimpleLoading.vue';
import deploymentsIcon from '@apps/control-hub/assets/Deployments.svg';
import ListItem from '../components/MyListItem/ListItem.vue';
import { useI18n } from 'vue-i18n';
import { componentName } from '@apps/control-hub/src/router/const';

const { t } = useI18n();
interface Props {
	prefix: string;
	detail: any;
}
const route = useRoute();
const router = useRouter();
const props = withDefaults(defineProps<Props>(), {});
const loading = ref(false);
const revisionsData = ref();
const currentRevision = ref();
const kind = computed(() => get(props.detail, 'module'));

const fetchRevisions = () => {
	const { namespace, kind }: Record<string, any> = route.params;
	loading.value = true;
	getRevisions(props.detail.module, props.detail)
		.then((res: any) => {
			const list = res.data.items
				.map(ObjectMapper.revisions)
				.filter(
					(revision: any) =>
						revision.ownerName === props.detail.name &&
						revision.ownerKind === MODULE_KIND_MAP[kind]
				);

			currentRevision.value = getCurrentRevision(
				props.detail,
				list,
				props.detail.module
			);
		})
		.finally(() => {
			loading.value = false;
		});
};

const pathFormat = (data: any) => {
	router.push({
		name: componentName.WORKLOAD_POD_TOP,
		params: {
			kind: data.module,
			namespace: data.namespace,
			pods_name: data.name,
			pods_uid: data.uid
		},
		query: {
			type: 'workload'
		}
	});
};

fetchRevisions();
</script>

<style lang="scss" scoped>
.workload-item-container {
	border: 1px solid $separator;
	border-radius: 8px;
}
.workload-item-link {
	color: $ink-1;
	text-decoration: none;
	cursor: pointer;
	&:hover {
		color: $blue-default;
	}
}
</style>
