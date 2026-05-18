<template>
	<div v-for="item in data" :key="item" style="position: relative">
		<q-expansion-item
			v-if="item.children"
			:disable="item.disable"
			default-opened
			expand-icon-class="expand-icon-wrapper"
			:class="`bt-menu-item-size-${props.size}`"
			class="q-mt-xs"
		>
			<template #header>
				<q-item-section
					avatar
					:style="{ paddingLeft: `calc(${unit})` }"
					style="min-width: 0"
					class="q-pr-sm"
					v-if="item.icon || item.img"
				>
					<slot :name="`icon-${item.key}`"></slot>
					<AvatarIcon
						:data="item"
						:active="active === item.key"
						:size="props.size"
					>
					</AvatarIcon>
				</q-item-section>
				<ItemLabel
					:activeClass="activeClass"
					:data="item"
					:active="active === item.key"
					:size="props.size"
				></ItemLabel>
				<q-item-section side @click.stop style="padding: 0" class="q-ml-sm">
					<slot :name="`extra-${item.key}`" :item="item"></slot>
				</q-item-section>
			</template>
			<template v-for="child in item.children" :key="child">
				<q-expansion-item
					v-if="child.children"
					:disable="child.disable"
					default-opened
					expand-icon-class="expand-icon-wrapper"
					:header-class="`bt-menu-item-size-${props.size}`"
					class="q-mt-xs"
				>
					<template #header>
						<q-item-section
							avatar
							:style="{ paddingLeft: `calc(${unit} * 2)` }"
							style="min-width: 0"
							class="q-pr-sm"
							v-if="child.icon || child.img"
						>
							<slot :name="`icon-${child.key}`"></slot>
							<AvatarIcon
								:data="child"
								:active="active === child.key"
								:size="props.size"
							>
							</AvatarIcon>
						</q-item-section>
						<ItemLabel
							:activeClass="activeClass"
							:data="child"
							:active="active === child.key"
							:size="props.size"
						></ItemLabel>
						<q-item-section side @click.stop style="padding: 0" class="q-ml-sm">
							<slot :name="`extra-${child.key}`" :item="child"></slot>
						</q-item-section>
					</template>
					<template v-for="childDeep in child.children" :key="childDeep">
						<MenuItem
							:data="childDeep"
							:style="{ paddingLeft: `calc(${nomal} + ${unit} * 3)` }"
							:activeClass="activeClass"
							v-model="active"
							:size="props.size"
							@select="selectHandler"
						>
							<template v-slot:[`extra-${childDeep.key}`]="scope">
								<slot
									:name="`extra-${childDeep.key}`"
									:item="scope.item"
								></slot>
							</template>
							<template v-slot:[`icon-${childDeep.key}`]="scope">
								<slot :name="`icon-${childDeep.key}`" :item="scope.item"></slot>
							</template>
						</MenuItem>
					</template>
				</q-expansion-item>

				<MenuItem
					v-else
					:data="child"
					:style="{ paddingLeft: `calc(${nomal} + ${unit} * 2)` }"
					:activeClass="activeClass"
					v-model="active"
					:size="props.size"
					@select="selectHandler"
				>
					<template v-slot:[`extra-${child.key}`]="scope">
						<slot :name="`extra-${child.key}`" :item="scope.item"></slot>
					</template>
					<template
						v-if="$slots[`icon-${child.key}`]"
						v-slot:[`icon-${child.key}`]="scope"
					>
						<slot :name="`icon-${child.key}`" :item="scope.item"></slot>
					</template>
				</MenuItem>
			</template>
		</q-expansion-item>
		<MenuItem
			v-else
			:data="item"
			:style="{ paddingLeft: `calc(${nomal} + ${unit})` }"
			:activeClass="activeClass"
			v-model="active"
			:size="props.size"
			@select="selectHandler"
		>
			<template v-slot:[`extra-${item.key}`]="scope">
				<slot :name="`extra-${item.key}`" :item="scope.item"></slot>
			</template>
			<template
				v-if="$slots[`icon-${item.key}`]"
				v-slot:[`icon-${item.key}`]="scope"
			>
				<slot :name="`icon-${item.key}`" :item="scope.item"></slot>
			</template>
		</MenuItem>
	</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import MenuItem from './MenuItem.vue';
import { defaultSize, Item, Size } from './Menu';
import ItemLabel from './ItemLabel.vue';
import AvatarIcon from './AvatarIcon.vue';

interface Props {
	data: Item[];
	modelValue: string;
	activeClass: string;
	size: Size;
}

const props = withDefaults(defineProps<Props>(), {
	size: defaultSize
});

const active = ref(props.modelValue);

watch(
	() => props.modelValue,
	() => {
		active.value = props.modelValue;
	}
);

const nomal = '8px';
const unit = '12px';

const emit = defineEmits(['select', 'update:modelValue']);

const selectHandler = (data: { item: { muted: any }; key: any }) => {
	if (!data.item.muted) {
		emit('update:modelValue', data.key);
	}
	emit('select', data);
};
</script>

<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent({
	name: 'SubMenu'
});
</script>
