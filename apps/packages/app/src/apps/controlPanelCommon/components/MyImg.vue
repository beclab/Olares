<template>
	<div>
		<q-skeleton
			v-if="loading"
			:type="imgType"
			:width="width || '32px'"
			:height="height || '32px'"
			:class="typeStyle"
		/>
		<q-img
			v-else
			class="rounded-borders-sm title-img my-avatar-img"
			:class="typeStyle"
			fit="contain"
			:width="width || '32px'"
			:height="height || '32px'"
			spinner-size="sm"
			spinner-color="grey-1"
			loading="lazy"
			:src="src"
			v-bind="$attrs"
		>
			<template #loading>
				<q-skeleton :type="imgType" class="full-width full-height" />
			</template>
		</q-img>
	</div>
</template>

<script setup lang="ts">
import { computed, toRefs } from 'vue';

interface Props {
	src: string;
	loading?: boolean;
	type?: 'rect' | 'circle';
	width?: string;
	height?: string;
}
const props = withDefaults(defineProps<Props>(), {
	type: 'rect'
});

const { type: imgType } = toRefs(props);

const typeStyle = computed(() =>
	imgType.value === 'circle' ? 'circle' : 'rect'
);
</script>

<style lang="scss" scoped>
.circle {
	border-radius: 50%;
}
.rect {
	border-radius: 8px;
}
.my-avatar-img {
	box-shadow: 0px 2px 4px 0px rgba(0, 0, 0, 0.1);
}
</style>
