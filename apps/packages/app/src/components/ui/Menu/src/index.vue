<template>
	<div class="bt-menu-container q-pa-md">
		<Theme v-if="props.showThemeToggle" :follow-system="props.followSystem" />
		<div class="q-pb-lg" v-if="$slots.header">
			<slot name="header"></slot>
		</div>
		<template v-for="(item, index) in items" :key="item">
			<q-item-label
				header
				class="q-py-none q-px-sm"
				lines="1"
				:class="[index !== 0 ? topLabelClass : 0]"
				:style="{
					marginTop: item.label ? '0px' : '-4px'
				}"
			>
				<div class="row items-center">
					<div style="flex: 1" class="text-ink-3" :class="topLabelTextClass">
						{{ item.label }}
					</div>
					<div>
						<slot :name="`extra-${item.key}`"></slot>
					</div>
				</div>
			</q-item-label>
			<template v-for="child in item.children" :key="child">
				<q-expansion-item
					:hide-expand-icon="hideExpandIcon"
					v-if="child.children"
					:disable="child.disable"
					:default-opened="
						child.defaultClose
							? child.defaultClose
							: defaultClose
							? false
							: true
					"
					expand-icon-class="expand-icon-wrapper"
					:header-class="['q-pa-sm', 'expand-header-wrapper']"
					:header-style="{
						borderRadius: '8px'
					}"
					class="q-mt-xs"
					@update:model-value="
						(value) => {
							return toggleVaule(value, child);
						}
					"
				>
					<template #header>
						<q-item-section
							avatar
							:style="{ paddingLeft: `calc(${unit})` }"
							style="min-width: 0"
							class="q-pr-sm"
							v-if="child.icon || child.img"
						>
							<slot
								v-if="$slots[`icon-${child.key}`]"
								:name="`icon-${child.key}`"
							></slot>
							<AvatarIcon
								v-else
								:data="child"
								:active="link === child.key"
								:size="props.size"
							>
							</AvatarIcon>
						</q-item-section>
						<ItemLabel
							:activeClass="activeClassConputed"
							:data="child"
							:active="link === child.key"
							:size="props.size"
						></ItemLabel>
						<q-item-section side @click.stop style="padding: 0" class="q-ml-sm">
							<slot :name="`extra-${child.key}`" :item="child"></slot>
						</q-item-section>
					</template>
					<SubMenu
						:activeClass="activeClassConputed"
						:data="child.children"
						:size="props.size"
						@select="selectHandler"
						v-model="link"
					>
						<template v-for="(value, key) in $slots" v-slot:[key]="scope">
							<slot :name="key" :item="scope.item"></slot>
						</template>
					</SubMenu>
				</q-expansion-item>
				<MenuItem
					v-else
					:activeClass="activeClassConputed"
					:same-activeable="sameActiveable"
					:data="child"
					:size="props.size"
					@select="selectHandler"
					v-model="link"
				>
					<template v-slot:[`extra-${child.key}`]>
						<slot :name="`extra-${child.key}`" :item="child"></slot>
					</template>
					<template
						v-if="$slots[`icon-${child.key}`]"
						v-slot:[`icon-${child.key}`]
					>
						<slot :name="`icon-${child.key}`"></slot>
					</template>
				</MenuItem>
			</template>
		</template>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import MenuItem from './MenuItem.vue';
import SubMenu from './SubMenu.vue';
import { defaultSize, Item, Size } from './Menu';
import ItemLabel from './ItemLabel.vue';
import AvatarIcon from './AvatarIcon.vue';
import Theme from '../../theme/src/IndexPage.vue';

interface Props {
	/**
	 * The Menu data
	 */
	items: Array<Item>;
	/**
	 * defaultSelected: string
	 */
	defaultActive: string;
	/**
	 * hightlight class style
	 */
	activeClass: string;
	/**
	 *  v-model
	 */
	modelValue: string;
	/**
	 * defaultClose: boolean
	 */
	defaultClose: boolean;

	/**
	 * hideExpandIcon: boolean
	 */
	hideExpandIcon: boolean;
	sameActiveable: boolean;
	showThemeToggle: boolean;
	followSystem?: boolean;
	size: Size;
}
const props = withDefaults(defineProps<Props>(), {
	sameActiveable: false,
	showThemeToggle: true,
	followSystem: true,
	size: defaultSize
});

const link = ref(props.defaultActive || props.modelValue);
const activeClassConputed = computed(
	() => props.activeClass || 'bt-menu-active-link'
);

const topLabelClass = computed(() => {
	switch (props.size) {
		case 'md':
			return 'q-pt-lg';
		case 'sm':
			return 'q-pt-sm';
		default:
			return 'q-pt-lg';
	}
});

const topLabelTextClass = computed(() => {
	switch (props.size) {
		case 'md':
			return 'text-subtitle2';
		case 'sm':
			return 'text-subtitle3';
		default:
			return 'text-subtitle2';
	}
});

watch(
	() => props.modelValue,
	() => {
		link.value = props.modelValue;
	}
);

const emit = defineEmits<
	{
		/**
		 * Menu item selected event
		 */
		(
			e: 'select',
			data: { item: Item[]; key: string; selectedKeys: string }
		): void;
	},
	{
		/**
		 * Menu item toogle event
		 */
		(e: 'toggleVaule', data: { value: boolean; item: Item[] }): void;
	}
>();
const selectHandler = (data: {
	item: Item[];
	key: string;
	selectedKeys: string;
}) => {
	if (props.modelValue && !data.item.muted) {
		emit('update:modelValue', data.key);
	}
	emit('select', data);
};

const toggleVaule = (value, item) => {
	emit('toggleVaule', { value, item });
};

const menuWidthDefault = computed(() => {
	switch (props.size) {
		case 'md':
			return '240px';
		case 'sm':
			return '180px';
		default:
			return '240px';
	}
});
</script>

<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent({
	name: 'BtMenu'
});
</script>

<style lang="scss" scoped>
.bt-menu-item-size-md-var {
	border-radius: 8px;
	padding: 8px;
	min-height: 36px;
}
.bt-menu-item-size-sm-var {
	border-radius: 4px;
	padding: 4px 8px;
	min-height: 32px;
}
.bt-menu-container {
	width: v-bind(menuWidthDefault);
	color: $grey-8;
	.bt-menu-item-size-md {
		@extend .bt-menu-item-size-md-var;
	}
	.bt-menu-item-size-sm {
		@extend .bt-menu-item-size-sm-var;
	}

	::v-deep(.bt-menu-active-link) {
		color: $blue-6;
		background-color: $blue-1;
	}
	::v-deep(.expand-icon-wrapper) {
		padding-left: 0;
	}
	::v-deep(.expand-header-wrapper) {
		min-height: auto;
	}
	::v-deep(.q-hoverable:hover > .q-focus-helper) {
		opacity: 0.08;
	}
	::v-deep(.q-item__section--side > .q-icon) {
		font-size: 20px;
	}
	::v-deep(.q-ripple) {
		opacity: 0.25;
	}
}
::v-deep(.q-item) {
	color: $ink-2;
}
::v-deep(.bt-menu-item-size-md .q-expansion-item__container > .q-item) {
	@extend .bt-menu-item-size-md-var;
}
::v-deep(.bt-menu-item-size-md .q-expansion-item__content > .q-item) {
	@extend .bt-menu-item-size-md-var;
}
::v-deep(.bt-menu-item-size-sm .q-expansion-item__container > .q-item) {
	@extend .bt-menu-item-size-sm-var;
}
::v-deep(.bt-menu-item-size-sm .q-expansion-item__content > .q-item) {
	@extend .bt-menu-item-size-sm-var;
}

::v-deep(.bt-menu-item-size-sm .q-item__section--side > .q-icon) {
	font-size: 18px;
}
</style>
