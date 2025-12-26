<template>
	<q-card flat class="volume-item-container q-pa-lg">
		<div bordered separator class="row no-wrap q-col-gutter-x-md">
			<div class="col-4">
				<ListItem
					size="sm"
					icon-size="lg"
					:title="titleName"
					:sub-title="description"
				>
					<template #avatar>
						<q-img
							:src="volumesDataIcon"
							fit="contain"
							style="flex: 0 0 32px"
						/>
					</template>
				</ListItem>
			</div>
			<div class="col-8 row">
				<div
					class="col q-col-gutter-x-md"
					v-for="(item, index) in details"
					:key="index"
				>
					<ListItem
						size="sm"
						icon-size="lg"
						:title="item.title"
						:sub-title="item.description"
					>
					</ListItem>
				</div>
			</div>
		</div>
		<q-separator color="separator" spaced style="margin: 20px 0" />
		<div v-if="!isEmpty(volume.volumeMounts)" class="q-col-gutter-y-sm">
			<div
				v-for="(item, index) in volume.volumeMounts"
				:key="index"
				class="row items-center q-col-gutter-x-md"
			>
				<div class="col-4">
					<ListItem size="sm" iconSize="sm" :bolder="false">
						<template #avatar>
							<q-img :src="dockerIcon" />
						</template>
						<template #title>
							<span class="text-ink-1">{{ item.containerName }}</span>
						</template>
					</ListItem>
				</div>
				<div class="col-8">
					<ListItem size="sm" iconSize="sm" :bolder="false">
						<template #avatar>
							<q-img :src="padlocIcon" />
						</template>
						<template #title>
							<div class="row q-gutter-x-md text-ink-1">
								<div>
									<span>{{ item.mountPath }}</span>
									<span class="q-ml-xs text-ink-3"
										>({{
											item.readOnly
												? t('READ_ONLY_LOW')
												: t('READ_AND_WRITE_LOW')
										}})</span
									>
								</div>
								<span v-if="item.logPath">{{ item.logPath }}</span>
							</div>
						</template>
					</ListItem>
				</div>
			</div>
		</div>
	</q-card>
</template>

<script lang="ts" setup>
import { get, isEmpty } from 'lodash';
import { computed, h, ref, toRef, toRefs, watch } from 'vue';
import { useRoute } from 'vue-router';
import MyCard from '../components/MyCard.vue';
import dockerIcon from '../assets/docker.svg';
import volumesDataIcon from '../assets/volumes-data.svg';
import padlocIcon from '../assets/padloc.svg';
import ListItem from '../components/MyListItem/ListItem.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface Props {
	volume: any;
	isMultiProject?: boolean;
}
const props = withDefaults(defineProps<Props>(), {});

const route = useRoute();
const workspace = ref(route.params.workspace);
const cluster = ref(route.params.cluster);
const namespace = ref(route.params.namespace);

const { volume, isMultiProject }: any = toRefs(props);

const prefix = computed(
	() =>
		`${workspace.value ? `/${workspace.value}` : ''}/clusters/${
			cluster.value
		}/projects/${namespace.value}`
);

const titleName = computed(() => {
	return get(volume, 'mountName', volume.value.name);
});

const icon = computed(() => {
	if (volume.value.hostPath || volume.value.emptyDir) {
		return 'storage';
	} else if (volume.value.configMap) {
		return 'hammer';
	} else if (volume.value.secret) {
		return 'key';
	} else {
		return 'storage';
	}
});

const description = ref('');
const details: any = ref([]);
watch(
	() => props.volume,
	(volume) => {
		updateData(volume);
	},
	{
		immediate: true
	}
);

function updateData(volume: any) {
	const prefix = `${workspace.value ? `/${workspace.value}` : ''}/clusters/${
		cluster.value
	}/projects/${namespace.value}`;

	let icon = 'storage';
	if (volume.hostPath) {
		description.value = t('TYPE_HOSTPATH');
		details.value = [
			{ title: get(volume, 'hostPath.path', '-'), description: 'HostPath' }
		];
	} else if (volume.emptyDir) {
		description.value = t('TYPE_EMPTYDIR');
		details.value = [];
	} else if (volume.configMap) {
		icon = 'hammer';
		description.value = t('TYPE_CONFIGMAP');
		const configmap = get(volume, 'configMap.name', '-');
		details.value = [
			{
				title: configmap,
				link: `${prefix}/configmaps/${configmap}`,
				description: t('CONFIGMAP')
			}
		];
	} else if (volume.secret) {
		icon = 'key';
		description.value = t('TYPE_SECRET');
		const secret = get(volume, 'secret.secretName', '-');
		details.value = [
			{
				title: secret,
				link: `${prefix}/secrets/${secret}`,
				description: t('SECRET')
			}
		];
	} else {
		description.value = t('STORAGE_CLASS_VALUE', {
			value: get(volume, 'storageClassName', '-')
		});
		const volumeName = get(volume, 'persistentVolumeClaim.claimName', '-');
		details.value = [
			{
				title: volumeName,
				link: `${prefix}/volumes/${volumeName}`,
				description: t('PERSISTENT_VOLUME_CLAIM')
			},
			{
				title: get(volume, 'capacity', '-'),
				description: t('CAPACITY')
			},
			{
				title: get(volume, 'accessMode', '-'),
				description: t('ACCESS_MODE_SCAP')
			}
		];
	}
}
</script>

<style lang="scss" scoped>
.volume-item-container {
	border: 1px solid $separator;
	border-radius: 8px;
}
</style>
