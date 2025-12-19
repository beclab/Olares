<template>
	<div class="q-col-gutter-xs">
		<div v-for="(item, index) in pages" :key="index">
			<div class="row justify-between">
				<span>{{ item.x }}</span>
				<div class="row items-center" style="width: 200px">
					<span>{{ item.y }}</span>
					<q-linear-progress
						style="flex: 1"
						class="q-ml-md"
						size="20px"
						:value="progressLabel(item.z) / 100"
					>
						<div class="absolute-full flex">
							<q-badge :label="`${progressLabel(item.z)}%`" />
						</div>
					</q-linear-progress>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { percentFilter } from '@apps/dashboard/src/utils/filters';
import { computed } from 'vue';
import firstBy from 'thenby';
import { round } from 'lodash';

interface Props {
	data: any;
}

const props = withDefaults(defineProps<Props>(), {});

const pages = computed(() => {
	const data = percentFilter(
		props.data.pageviews
			.reduce((arr, { urlPath }) => {
				const row = arr.find(({ x }) => x === urlPath);

				if (!row) {
					arr.push({ x: urlPath, y: 1 });
				} else {
					row.y += 1;
				}
				return arr;
			}, [])
			.sort(firstBy('y', -1))
	);
	return data;
});

const progressLabel = (value: number) => {
	return round(value, 1);
};
</script>

<style></style>
