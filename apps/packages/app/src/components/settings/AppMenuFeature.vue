<template>
	<bt-list first>
		<div
			class="row justify-between items-center q-pa-lg"
			:class="clickable ? 'cursor-pointer' : ''"
		>
			<div class="row justify-start items-center">
				<q-img class="menu-icon" :src="menuItem ? menuItem.img : image" />
				<div
					style="width: calc(100% - 52px)"
					class="column justify-start q-ml-md"
				>
					<div class="row justify-start items-center">
						<div class="text-h6 text-ink-1">
							{{
								menuItem
									? menuItem.title
										? t(menuItem.title)
										: t(menuItem.label)
									: label
							}}
						</div>
						<slot name="status" />
					</div>
					<div class="text-body3 text-ink-3">
						{{ menuItem ? t(menuItem.description) : description }}
					</div>
				</div>
			</div>
			<div style="flex: 0 0 24" v-if="clickable && !buttonSlot">
				<q-icon
					class="text-ink-2"
					name="sym_r_keyboard_arrow_right"
					size="24px"
				/>
			</div>
			<slot name="button" />
		</div>
	</bt-list>
</template>
<script setup lang="ts">
import BtList from './base/BtList.vue';
import { computed, PropType, useSlots } from 'vue';
import { MENU_TYPE, useMenuItem } from 'src/constant';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	menuType: {
		type: Object as PropType<MENU_TYPE>,
		required: false
	},
	label: {
		type: String,
		required: false
	},
	image: {
		type: String,
		required: false
	},
	description: {
		type: Boolean,
		required: false
	},
	clickable: {
		type: Boolean,
		default: false
	}
});

const menuItem = computed(() => {
	return useMenuItem(props.menuType);
});

const buttonSlot = !!useSlots().button;
const { t } = useI18n();
</script>

<style scoped lang="scss">
.menu-icon {
	width: 40px;
	height: 40px;
}
</style>
