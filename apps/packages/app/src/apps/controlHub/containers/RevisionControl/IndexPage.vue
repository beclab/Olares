<template>
	<div class="relative-position">
		<QSectionStyle2>
			<q-select outlined v-model="model" :options="options">
				<template v-slot:selected>
					<q-item>
						<q-item-section avatar center>
							<q-icon size="md" name="history" />
						</q-item-section>
						<q-item-section>
							<q-item-label>
								<div style="line-height: 30px">
									<span>{{ labelFormat(model) }}</span>
									<q-chip
										v-if="model && model.revision === curRevision"
										dense
										square
										color="primary"
										text-color="white"
										>{{ t('RUNNING') }}
									</q-chip>
								</div>
							</q-item-label>
							<q-item-label caption>
								{{
									t('CREATED_TIME', {
										diff: getLocalTime(model.createTime).format(
											'YYYY-MM-DD HH:mm:ss'
										)
									})
								}}
							</q-item-label>
						</q-item-section>
					</q-item>
				</template>

				<template v-slot:option="scope">
					<q-item v-bind="scope.itemProps">
						<q-item-section avatar>
							<q-icon size="md" name="history" />
						</q-item-section>
						<q-item-section>
							<q-item-label>
								<span> {{ labelFormat(scope.opt) }}</span>
								<q-chip
									v-if="scope.opt.revision === curRevision"
									dense
									square
									color="primary"
									text-color="white"
									>{{ t('RUNNING') }}
								</q-chip>
							</q-item-label>
							<q-item-label caption>
								{{
									t('CREATED_TIME', {
										diff: getLocalTime(scope.opt.createTime).format(
											'YYYY-MM-DD HH:mm:ss'
										)
									})
								}}
							</q-item-label>
						</q-item-section>
					</q-item>
				</template>
			</q-select>
		</QSectionStyle2>
		<DiffYaml :datas="renderDiff"></DiffYaml>
		<q-inner-loading :showing="loading"> </q-inner-loading>
	</div>
</template>

<script setup lang="ts">
import { getRevisions } from '@apps/control-hub/src/network';
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { t } from '@apps/control-hub/src/boot/i18n';
import { getLocalTime } from '@apps/control-hub/src/utils';
import { get, sortBy } from 'lodash-es';
import { getValue } from '@apps/control-hub/src/utils/yaml';
import DiffYaml from './DiffYaml.vue';
import { ObjectMapper } from '@apps/control-hub/src/utils/object.mapper';
import { MODULE_KIND_MAP } from '@apps/control-hub/src/utils/constants';
import QSectionStyle2 from '@apps/control-panel-common/src/components/QSectionStyle2.vue';

interface Props {
	detail: any;
	data: any;
}

const props = withDefaults(defineProps<Props>(), {
	detail: {},
	data: []
});

const model = ref({ revision: 0, createTime: '' });
const revisionsData = ref();
const loading = ref(false);
const route = useRoute();
const { namespace, kind: module, name }: { [key: string]: any } = route.params;

const getCurrentRevision = (
	workloadDetail: any,
	revisions: any,
	module = 'deployments'
) => {
	let revision: any = 0;

	switch (module) {
		default:
		case 'deployments':
			revision = get(
				workloadDetail,
				'annotations["deployment.kubernetes.io/revision"]'
			);
			break;
		case 'statefulsets':
		case 'daemonsets': {
			let maxRevision = get(revisions[0], 'revision', 0);
			for (let i = 1; i < revisions.length; i++) {
				if (revisions[i].revision > maxRevision) {
					maxRevision = revisions[i].revision;
				}
			}

			const cur = revisions.find(
				(item: any) =>
					item.name === get(workloadDetail, 'status.currentRevision')
			);
			revision = cur ? cur.revision : maxRevision;
			break;
		}
	}

	return parseInt(revision, 10);
};

const updateModel = (detail: any, data: any, module: any) => {
	const revision = getCurrentRevision(detail, data, module);
	console.log('revision', revision);
	const defaultRevision = revisionsData.value.find(
		(item: any) => item.revision === revision
	);
	console.log('revisionsData', revisionsData.value);
	model.value = defaultRevision;
	console.log('model', model);
};

const options = computed(() =>
	sortBy(revisionsData.value, (item) => parseInt(item.revision, 10)).reverse()
);

watch(
	() => props.detail,
	() => {
		model.value = props.detail.uid;
	}
);

const curRevision = computed(() => {
	const {
		namespace,
		kind: module,
		name
	}: { [key: string]: any } = route.params;

	return getCurrentRevision(props.detail, revisionsData.value, module);
});

const renderDiff = computed(() => {
	if (!model.value) {
		return;
	}
	const selectRevision = model.value.revision;
	const data = revisionsData.value || [];
	const lastRevision = Math.max(Number(selectRevision) - 1, 0);

	const newRevision = data.find(
		(item: any) => item.revision === selectRevision
	);
	const oldRevision = data.find((item: any) => item.revision === lastRevision);

	const newYaml = newRevision ? getValue(newRevision._originData) : '';
	const oldYaml = oldRevision ? getValue(oldRevision._originData) : '';

	return [oldYaml, newYaml];
});

const labelFormat = (item: any) => {
	if (!item || !item.revision) {
		return '';
	}
	return `#${item.revision} (${item.name.replace(`${item.ownerName}-`, '')})`;
};

const fetchRevisions = () => {
	const {
		cluster = 'default',
		namespace,
		kind
	}: Record<string, any> = route.params;
	console.log('props.detail', props.detail);
	loading.value = true;
	getRevisions(kind, props.detail)
		.then((res) => {
			revisionsData.value = res.data.items
				.map(ObjectMapper.revisions)
				.filter(
					(revision: any) =>
						revision.ownerName === props.detail.name &&
						revision.ownerKind === MODULE_KIND_MAP[kind]
				);
			console.log('revisionsData', revisionsData.value);
			updateModel(props.detail, revisionsData.value, module);
		})
		.finally(() => {
			loading.value = false;
		});
};
fetchRevisions();
</script>

<style></style>
