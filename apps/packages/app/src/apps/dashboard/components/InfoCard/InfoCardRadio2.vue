<template>
	<div class="column q-col-gutter-lg">
		<div v-for="(item, index) in list" :key="item.name">
			<InfoCardItem
				:active="index === active"
				@click="itemClick(index, item)"
				:used="item.used"
				:total="item.total"
				:name="item.name"
				:unit-type="item.unitType"
				:img="item.img"
				:img_active="item.img_active"
				:loading="loading"
			></InfoCardItem>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, toRef, toRefs } from 'vue';
import InfoCardItem from './InfoCardItem2.vue';
import { InfoCardItemProps } from './InfoCardItem.vue';

export interface InfoCardRadioProps {
	list?: Array<InfoCardItemProps>;
	defaultActive?: number;
	loading?: boolean;
}

const emit = defineEmits<{
	(e: 'change', data: InfoCardItemProps, index: number): void;
}>();

const props = withDefaults(defineProps<InfoCardRadioProps>(), {
	defaultActive: 0
});

const active = ref(props.defaultActive);
const itemClick = (index: number, data: InfoCardItemProps) => {
	active.value = index;
	emit('change', data, index);
};
</script>

<style></style>
