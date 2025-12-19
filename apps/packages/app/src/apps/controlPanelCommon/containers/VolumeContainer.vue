<template>
	<div class="q-gutter-y-md">
		<VolumeItem
			v-for="(item, index) in newVolumes"
			:key="index"
			:volume="item"
			:isMultiProject="isMultiProject"
			style="with: 100%"
		>
		</VolumeItem>
	</div>
	<Empty v-if="newVolumes.length === 0"></Empty>
</template>

<script setup lang="ts">
import { isEmpty } from 'lodash';
import { computed } from 'vue';
import VolumeItem from './VolumeItem.vue';
import Empty from '../components/Empty.vue';

interface Props {
	volumes: any[];
	containers: any[];
	isMultiProject: boolean;
}

const props = withDefaults(defineProps<Props>(), {});

const newVolumes = computed(() => {
	return props.volumes.map((item) => ({
		...item,
		volumeMounts: getVolumeMounts(item)
	}));
});

const getVolumeMounts = (volume: any) => {
	const containers = props.containers;
	const mounts: any[] = [];
	if (!containers) return undefined;
	console.log('containers', containers);
	containers.forEach((container: any) => {
		if (!isEmpty(container.volumeMounts)) {
			container.volumeMounts.forEach((mount: any) => {
				if (mount.name === volume.mountName) {
					mounts.push({
						...mount,
						containerName: container.name,
						accessMode: mount.readOnly ? 'read-only' : 'read-write'
					});
				}
			});
		}
	});
	return mounts;
};
</script>

<style></style>
