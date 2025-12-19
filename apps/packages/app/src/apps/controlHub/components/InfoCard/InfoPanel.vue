<template>
	<div class="info-card-container">
		<div
			class="row items-center info-card-wrapper"
			:class="[props.active ? 'info-card-wrapper-active' : '']"
			v-ripple
		>
			<img class="info-panel-icon" :src="icon" alt="" />
			<div class="info-wrapper">
				<div
					class="info-item"
					:class="[props.active ? 'info-item-active' : '']"
				>
					{{ _capitalize(name) }}
				</div>
				<div
					class="info-ratio"
					:class="[props.active ? 'info-ratio-active' : '']"
				>
					<span>{{ title }}</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { _capitalize } from '@apps/control-hub/src/utils/index';
import ringWhite from '@apps/control-hub/src/assets/ring-white.svg';
import serverWhite from '@apps/control-hub/src/assets/server-white.svg';
import ringBlue from '@apps/control-hub/src/assets/ring-blue.svg';
import serverBlue from '@apps/control-hub/src/assets/server-blue.svg';
import { computed } from 'vue';

export interface InfoCardItemProps {
	name?: string;
	title?: string;
	active?: boolean;
	icon?: string;
}

const props = withDefaults(defineProps<InfoCardItemProps>(), {
	active: false,
	title: ''
});

const icon = computed(() => {
	switch (props.icon) {
		case 'ring':
			return props.active ? ringWhite : ringBlue;
		case 'server':
			return props.active ? serverWhite : serverBlue;
		default:
			return props.active ? ringWhite : ringBlue;
	}
});
</script>

<style lang="scss" scoped>
.info-card-container {
	padding-bottom: 5px;
	cursor: default;
	width: 180px;
	.info-panel-icon {
		margin-right: 12px;
	}
	.info-card-wrapper {
		position: relative;
		padding-right: 44px;
		padding-left: 28px;
		padding: 12px;
		.info-wrapper {
			overflow: hidden;
		}

		&.info-card-wrapper-active {
			background: $primary;
			position: relative;
			border-radius: 4px;
			&::after {
				position: absolute;
				content: '';
				left: 50%;
				bottom: 0;
				border-left: 5px solid transparent;
				border-right: 5px solid transparent;
				border-top: 5px solid $primary;
				transform: translateY(100%);
			}
		}
		.info-item {
			font-size: 12px;
			line-height: 12px;
			color: #7e7e7e;
			&.info-item-active {
				color: rgba(255, 255, 255, 0.8);
			}
		}
		.info-ratio {
			font-weight: 500;
			font-size: 16px;
			line-height: 16px;
			color: #414141;
			margin-top: 8px;
			width: 120px;

			overflow: hidden;
			white-space: nowrap;
			text-overflow: ellipsis;
			&.info-ratio-active {
				color: #ffffff;
			}
			.ratio-foolter {
				font-weight: 400;
				font-size: 12px;
			}
		}
	}
}
</style>
