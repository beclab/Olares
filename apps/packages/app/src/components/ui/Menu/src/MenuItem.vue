<template>
	<q-item
		:disable="data.disable"
		:clickable="data.disableClickable ? !data.disableClickable : true"
		v-ripple
		:active="modelValue === data.key"
		:active-class="activeClass"
		class="rounded-borders q-mt-xs menu-item-wrapper"
		lines="1"
		:class="[`bt-menu-item-size-${props.size}`]"
		@click="clickHanlder(data)"
		@mouseenter="mouseenter(data)"
		@mouseleave="mouseleave(data)"
	>
		<q-item-section
			avatar
			class="q-pr-sm"
			style="min-width: 0"
			v-if="data.icon || data.img"
		>
			<slot v-if="$slots[`icon-${data.key}`]" :name="`icon-${data.key}`"></slot>
			<AvatarIcon
				v-else
				:data="data"
				:active="modelValue === data.key"
				:size="props.size"
			>
			</AvatarIcon>
		</q-item-section>
		<ItemLabel
			:activeClass="activeClass"
			:data="data"
			:active="modelValue === data.key"
			:size="props.size"
		></ItemLabel>
		<q-item-section
			:style="{
				opacity: !data.defaultHide || data.key === mouseItemKey ? 1 : 0
			}"
			side
			@click.stop
			style="padding: 0"
			class="q-ml-xs"
		>
			<slot :name="`extra-${data.key}`" :item="data"></slot>
		</q-item-section>
	</q-item>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { defaultSize, Item, Size } from './Menu';
import ItemLabel from './ItemLabel.vue';
import AvatarIcon from './AvatarIcon.vue';

interface Props {
	data: Item;
	modelValue: string;
	activeClass: string;
	sameActiveable: boolean;
	size: Size;
}

const props = withDefaults(defineProps<Props>(), {
	sameActiveable: false,
	size: defaultSize
});

const emit = defineEmits(['select', 'update:modelValue']);
const clickHanlder = (item: Item) => {
	if (!props.sameActiveable && item.key === props.modelValue) {
		return;
	}
	if (!item.muted) {
		emit('update:modelValue', item.key);
	}
	emit('select', { item, key: item.key, selectedKeys: item.key });
};

const mouseItemKey = ref();

const mouseenter = (item: Item) => {
	mouseItemKey.value = item.key;
};

const mouseleave = (item) => {
	mouseItemKey.value = null;
};

const slots = useSlots();
</script>

<script lang="ts">
import { defineComponent, useSlots } from 'vue';

export default defineComponent({
	name: 'MenuItem'
});
</script>
<style lang="scss" scoped></style>
