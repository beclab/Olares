<template>
	<div class="empty-root">
		<adaptive-layout>
			<template v-slot:pc>
				<div class="empty-view-pc column justify-center items-center">
					<div style="padding: 9px">
						<q-img
							class="menu-icon-pc"
							:src="menuItem ? menuItem?.img : image"
						/>
					</div>
					<div class="text-h5 text-ink-1 q-mt-sm text-center">{{ title }}</div>
					<div v-if="message" class="text-body2 text-ink-3 q-mt-sm text-center">
						{{ message }}
					</div>
					<div v-if="hasMessageSlot">
						<slot name="message" />
					</div>
					<q-btn
						v-if="buttonLabel"
						dense
						class="q-px-md q-py-sm q-mt-lg text-body3 text-capitalize"
						text-color="white"
						color="blue-default"
						:label="buttonLabel"
						@click="emit('onButtonClick')"
					>
						<slot />
					</q-btn>
				</div>
			</template>
			<template v-slot:mobile>
				<div
					class="empty-view-mobile column justify-center items-center q-pa-lg"
				>
					<div style="padding: 9px">
						<q-img
							class="menu-icon-mobile"
							:src="menuItem ? menuItem?.img : image"
						/>
					</div>
					<div class="text-subtitle2 text-ink-1 q-mt-sm text-center">
						{{ title }}
					</div>
					<div
						v-if="message"
						class="text-overline text-ink-3 q-mt-sm text-center"
					>
						{{ message }}
					</div>
					<div v-if="hasMessageSlot">
						<slot name="message" />
					</div>
				</div>
				<div
					v-if="buttonLabel"
					class="empty-bottom-mobile justify-end items-center q-pa-lg"
				>
					<q-btn
						dense
						class="full-width q-pa-md text-subtitle2 text-capitalize"
						text-color="white"
						color="blue-default"
						:label="buttonLabel"
						@click="emit('onButtonClick')"
					>
						<slot />
					</q-btn>
				</div>
			</template>
		</adaptive-layout>
	</div>
</template>
<script setup lang="ts">
import { MENU_TYPE, useMenuItem } from 'src/constant';
import AdaptiveLayout from './AdaptiveLayout.vue';
import { computed, PropType, useSlots } from 'vue';

const props = defineProps({
	menuType: {
		type: Object as PropType<MENU_TYPE>,
		required: false
	},
	title: {
		type: String,
		required: true
	},
	image: {
		type: String,
		required: false
	},
	message: {
		type: String,
		default: ''
	},
	buttonLabel: {
		type: String,
		default: ''
	}
});

const menuItem = computed(() => {
	return useMenuItem(props.menuType);
});

const hasMessageSlot = !!useSlots().message;
const emit = defineEmits(['onButtonClick']);
</script>

<style scoped lang="scss">
.empty-root {
	width: 100%;
	height: calc(100vh - 56px);
	position: relative;

	.empty-view-pc {
		margin-top: 12px;
		margin-left: 20px;
		margin-right: 20px;
		padding: 80px;
		width: calc(100% - 40px);

		.menu-icon-pc {
			width: 83px;
			height: 83px;
		}
	}

	.empty-view-mobile {
		height: 236px;
		width: 100%;

		.menu-icon-mobile {
			width: 83px;
			height: 83px;
		}
	}

	.empty-bottom-mobile {
		width: 100%;
		height: 88px;
		position: absolute;
		bottom: 0;
	}
}
</style>
