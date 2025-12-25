<template>
	<div
		class="my-empty-container"
		:class="{ 'my-empty-container-center': center }"
	>
		<div class="empty-content" :class="{ 'empty-content-center': center }">
			<img
				:class="emptyLogoSize"
				:src="getRequireImage('log_empty.svg')"
				alt=""
			/>
			<div style="font-size: 12px">{{ text ? text : 'No Data' }}</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { getRequireImage } from 'src/utils/rss-utils';

interface Props {
	text?: string;
	size?: 'mini' | 'small' | 'medium' | 'large';
	center?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	size: 'small'
});

const emptyLogoSize = `${props.size}`;
</script>

<style lang="scss" scoped>
.my-empty-container {
	text-align: center;
	color: $grey-6;
	&.my-empty-container-center {
		position: absolute;
		top: 0;
		right: 0;
		bottom: 0;
		left: 0;
		z-index: 1;
	}
	.empty-content {
		&.empty-content-center {
			position: absolute;
			top: 50%;
			left: 50%;
			transform: translate(-50%, -50%);
		}
		.empty-icon {
			font-size: 22px;
		}
		.large {
			width: 160px;
		}
		.medium {
			width: 120px;
		}
		.small {
			width: 80px;
		}
		.mini {
			width: 56px;
			font-size: 12px;
		}
	}
}
</style>
