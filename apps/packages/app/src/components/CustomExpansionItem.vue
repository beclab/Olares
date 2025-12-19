<template>
	<div class="custom-expansion-item q-pa-lg" :class="{ 'is-open': isOpen }">
		<div class="expansion-header" @click.stop="toggle">
			<slot name="header" :isOpen="isOpen">
				<div class="full-width row justify-between items-center">
					<div class="text-ink-1 text-subtitle2">{{ label }}</div>
					<div
						class="arrow"
						:style="{ transform: isOpen ? 'rotate(180deg)' : 'rotate(0)' }"
					>
						<q-icon
							class="text-ink-2"
							size="24px"
							:name="icon ? icon : 'sym_r_keyboard_arrow_up'"
						/>
					</div>
				</div>
			</slot>
		</div>

		<transition name="slide-down">
			<div v-if="isOpen">
				<slot />
			</div>
		</transition>
	</div>
</template>

<script setup lang="ts">
import { ref, defineProps, defineEmits } from 'vue';

const props = defineProps<{
	initialOpen?: boolean;
	disabled?: boolean;
	label?: string;
	icon?: string;
}>();

const emit = defineEmits<{
	(e: 'toggle', value: boolean): void;
}>();

const isOpen = ref(props.initialOpen);

const toggle = () => {
	if (props.disabled) return;
	isOpen.value = !isOpen.value;
	emit('toggle', isOpen.value);
};
</script>

<style scoped>
.custom-expansion-item {
	border-radius: 4px;
	overflow: hidden;
}

.expansion-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	cursor: pointer;
	user-select: none;
}

.arrow {
	transition: transform 0.3s ease;
	font-size: 12px;
}

.slide-down-enter-from,
.slide-down-leave-to {
	max-height: 0;
	padding-top: 0;
	padding-bottom: 0;
	opacity: 0;
}

.slide-down-enter-active,
.slide-down-leave-active {
	transition: all 0.3s ease;
	max-height: 500px;
}
</style>
