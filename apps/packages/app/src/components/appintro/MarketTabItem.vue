<template>
	<adaptive-layout>
		<template v-slot:pc>
			<div
				class="row justify-center items-center"
				:class="selected ? 'tab-item-root-selected' : 'tab-item-root'"
			>
				<q-icon
					v-if="icon"
					:name="icon"
					class="font-icon q-mr-xs"
					size="16px"
				/>
				<div class="tab-item-title text-body2">{{ label }}</div>
			</div>
		</template>

		<template v-slot:mobile>
			<div class="column justify-start items-center relative">
				<div
					class="row justify-center items-center"
					:class="
						selected ? 'tab-item-root-selected-mobile' : 'tab-item-root-mobile'
					"
				>
					<q-icon
						v-if="icon"
						:name="icon"
						class="font-icon q-mr-xs"
						size="16px"
					/>
					<div class="tab-item-title-mobile text-body2">{{ label }}</div>
				</div>
				<div
					class="absolute"
					style="height: 3px; width: 20px; bottom: 0"
					:class="selected ? 'bg-blue-default' : 'bg-transparent'"
				/>
			</div>
		</template>
	</adaptive-layout>
</template>

<script setup lang="ts">
import AdaptiveLayout from '../settings/AdaptiveLayout.vue';

defineProps({
	selected: {
		type: Boolean,
		default: false
	},
	label: {
		type: String,
		default: '',
		require: false
	},
	icon: {
		type: String,
		default: '',
		require: false
	}
});
</script>

<style scoped lang="scss">
.tab-item-root {
	height: 32px;
	width: 100px;
	padding-left: 12px;
	padding-right: 12px;
	color: $ink-2;
	text-align: center;
	border-radius: 8px;
}

.tab-item-root-selected {
	@extend .tab-item-root;
	border-radius: 8px;
	color: $blue-default;
	background-color: $blue-soft;
}

.tab-item-title {
	text-align: center;
	text-transform: capitalize;
}

.tab-item-root-mobile {
	height: 32px;
	width: auto;
	padding-left: 12px;
	padding-right: 12px;
	color: $ink-2;
	text-align: center;
}

.tab-item-root-selected-mobile {
	@extend .tab-item-root-mobile;
	color: $blue-default;
}

.tab-item-title-mobile {
	text-align: center;
	text-transform: capitalize;
}
</style>
