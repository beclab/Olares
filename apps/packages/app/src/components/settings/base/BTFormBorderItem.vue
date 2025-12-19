<template>
	<div class="form-item-border-root column justify-start">
		<div class="text-body3 text-ink-3 q-mb-xs" v-if="title">
			{{ title }}
		</div>
		<q-item
			:clickable="chevronRight"
			class="form-item"
			:class="chevronRight ? 'item' : ''"
		>
			<q-item-section
				class="item-margin-left text-ink-2"
				:class="
					deviceStore.isMobile
						? 'text-subtitle3-m title-margin'
						: 'text-body2 title-margin-mobile'
				"
				v-if="itemTitle"
			>
				{{ itemTitle }}
			</q-item-section>
			<q-item-section
				v-else
				class="f item-margin-left"
				:class="deviceStore.isMobile ? 'text-subtitle3-m' : 'text-body1'"
			>
				<slot name="title" />
			</q-item-section>
			<q-item-section
				side
				:class="
					sideClasses && sideClasses.length ? sideClasses : 'item-margin-right'
				"
				:style="
					dataWidthP > 0
						? `width: ${dataWidthP}%;`
						: dataMaxWidthP > 0
						? `max-width:${dataMaxWidthP}%`
						: ''
				"
			>
				<div
					class="form-item-data"
					v-if="data"
					:class="deviceStore.isMobile ? 'text-body3-m' : 'text-body2'"
				>
					{{ data }}
				</div>
				<q-icon
					v-else-if="chevronRight"
					name="sym_r_chevron_right"
					size="20px"
				/>
				<slot v-else />
			</q-item-section>
		</q-item>
	</div>
</template>

<script lang="ts" setup>
import { useDeviceStore } from 'src/stores/settings/device';
// import BtSeparator from '../base/BtSeparator.vue';

defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	itemTitle: {
		type: String,
		required: false,
		default: ''
	},
	chevronRight: {
		type: Boolean,
		default: false,
		required: false
	},
	dataWidthP: {
		type: Number,
		default: 0
	},
	dataMaxWidthP: {
		type: Number,
		default: 60
	},
	data: {
		type: String,
		require: false
	},
	sideClasses: {
		type: String,
		require: false
	}
});
const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss">
.form-item-border-root {
	width: 100%;
	height: auto;
	// padding-right: 20px;

	.form-item {
		min-height: 40px;
		padding: 0;
		backdrop-filter: blur(6.07811px);
		border-radius: 8px;
		border: 1px solid $separator;
	}

	.item-margin-left {
		margin-left: 20px;
	}

	.item-margin-right {
		margin-right: 20px;
	}
}
</style>
