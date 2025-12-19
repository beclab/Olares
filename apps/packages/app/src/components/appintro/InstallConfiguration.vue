<template>
	<div
		class="slip_layer_level_item column justify-center items-center"
		:style="{ '--line': last ? 'none' : `1px solid ${separator}` }"
	>
		<div
			class="slip_layer_level_name text-ink-3"
			:class="deviceStore.isMobile ? 'text-overline' : 'text-body3'"
		>
			{{ name }}
		</div>

		<q-icon
			v-if="src"
			:size="deviceStore.isMobile ? '32px' : '24px'"
			class="slip_layer_level_icon"
			color="ink-2"
			:src="imageRef"
			:name="src"
		/>

		<!--		<div class="custom-icon-div slip_layer_level_icon" v-if="src">-->
		<!--			<img :src="showIconAddress('sym_r_group')" alt="group" />-->
		<!--		</div>-->

		<div v-else class="row justify-center items-center">
			<div
				class="slip_layer_level_number text-ink-2"
				:class="deviceStore.isMobile ? 'text-h4' : 'text-subtitle1'"
			>
				{{ data }}
			</div>
		</div>
		<div
			class="slip_layer_level_unit text-body3 text-ink-2"
			:class="deviceStore.isMobile ? 'text-caption' : 'text-text-body3'"
		>
			{{ unit ? unit : '-' }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useColor } from '@bytetrade/ui';
import { useDeviceStore } from '../../stores/settings/device';

defineProps({
	name: {
		type: String,
		default: '',
		require: true
	},
	src: String,
	data: {
		type: String,
		default: '-'
	},
	unit: {
		type: String,
		default: '-'
	},
	last: {
		type: Boolean,
		default: false
	}
});

const imageRef = ref();
const { color: separator } = useColor('separator');
const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss">
.slip_layer_level_item {
	height: 80px;
	min-width: 100px;
	width: 100%;
	align-items: center;
	border-right: var(--line);

	.slip_layer_level_name {
		text-align: center;
		width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		display: -webkit-box;
		-webkit-line-clamp: 1;
		-webkit-box-orient: vertical;
		line-height: 16px;
	}

	.slip_layer_level_icon {
		margin-top: 8px;
		margin-bottom: 8px;
	}

	.slip_layer_level_number {
		height: 24px;
		margin-top: 8px;
		margin-bottom: 8px;
		text-align: center;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.slip_layer_level_unit {
		text-align: center;
		width: 100%;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
}
</style>
