<template>
	<div>
		<MyGridLayout col-width="168px" gap="lg" class="info-card-radio-grid">
			<div
				v-for="(item, index) in list"
				:key="item.name"
				@click="routeTo(item)"
			>
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
					:info="item.info"
					:percent="item.percent"
				>
					<template #network v-if="item.id === 'network'">
						<slot name="network"></slot>
					</template>
					<template #fan v-if="item.id === 'fan'">
						<slot name="fan"></slot>
					</template>
				</InfoCardItem>
			</div>
			<slot></slot>
		</MyGridLayout>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, toRef, toRefs } from 'vue';
import InfoCardItem from './InfoCardItem.vue';
import { InfoCardItemProps } from './InfoCardItem.vue';
import { useRouter } from 'vue-router';
import MyGridLayout from '@apps/control-panel-common/src/components/MyGridLayout.vue';
const router = useRouter();
export interface InfoCardRadioProps {
	list: Array<InfoCardItemProps>;
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

const routeTo = (item: InfoCardItemProps) => {
	router.push(item.route);
};

const gridRepeatNum = computed(() => props.list.length || 6);
const gridRepeatNumHalf = computed(() => Math.ceil(gridRepeatNum.value / 2));
</script>

<style lang="scss" scoped>
.info-card-radio-grid {
	grid-template-columns: repeat(
		v-bind(gridRepeatNum),
		minmax(168px, 1fr)
	) !important;
}
@media (max-width: 1636px) {
	.info-card-radio-grid {
		grid-template-columns: repeat(
			v-bind(gridRepeatNumHalf),
			minmax(168px, 1fr)
		) !important;
	}
}
</style>
