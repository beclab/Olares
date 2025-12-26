<template>
	<q-list class="q-py-sm q-px-sm" dense style="height: 100%">
		<q-item
			clickable
			v-ripple
			:active="item.id || index === active"
			class="my-menu-link q-my-xs rounded-borders"
			active-class="my-menu-link-active text-primary"
			v-for="(item, index) in data"
			:key="index"
			@click="clickHandler(item, index)"
		>
			<q-item-section avatar style="min-width: 0" class="q-pr-sm">
				<q-avatar
					color="primary"
					size="sm"
					text-color="white"
					icon="bluetooth"
				/>
			</q-item-section>
			<q-item-section>
				<q-item-label lines="1">{{ item.title }}</q-item-label>
				<q-item-label caption lines="1">{{ item.desc }}</q-item-label>
			</q-item-section>
		</q-item>
	</q-list>
	<Empty v-if="isEmpty" size="mini" center></Empty>
	<q-inner-loading :showing="loading"></q-inner-loading>
</template>

<script setup lang="ts">
import { isArray, isNil } from 'lodash-es';
import { computed, ref, watch } from 'vue';

interface Props {
	data: any[];
	defaultActive?: string | number;
	loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	data: [] as any,
	loading: false
});

const emits = defineEmits(['selected']);
const active = ref(props.defaultActive);

const isEmpty = computed(() => {
	const empty = isNil(props.data)
		? true
		: isArray(props.data) && props.data.length === 0
		? true
		: false;
	return !props.loading && empty;
});

const clickHandler = (item: any, index: number) => {
	active.value = item.id || index;
	emits('selected', active.value);
};

watch(
	() => props.defaultActive,
	(newData) => {
		active.value = newData;
	}
);
</script>

<style lang="scss" scoped>
.my-menu-link {
	padding: 4px 6px;
}
.my-menu-link-active {
	background-color: rgba(34, 111, 255, 0.12);
}
.avatar-wrapper {
	border: 1px solid rgba(0, 0, 0, 0);
	&.avatar-active {
		border-radius: 50%;
		box-sizing: border-box;
		border: 1px solid $primary;
	}
}

.active {
	// color: $primary;
	// background-color: rgba(34, 111, 255, 0.12);
}
</style>
