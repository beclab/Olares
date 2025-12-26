<template>
	<div
		class="my-empty-contianer"
		:class="{ 'my-empty-container-center': center }"
	>
		<div class="empty-content" :class="{ 'empty-content-center': center }">
			<img :class="emptyLogoZize" :src="emptyIcon" alt="" />
			<div style="font-size: 12px" class="text-ink-2">
				{{ text ? text : $t('NO_DATA_AVAILABLE') }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import emptyIcon from '../assets/empty.svg';

interface Props {
	text?: string;
	size?: 'mini' | 'small' | 'medium' | 'large';
	center?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	size: 'small'
});

const emptyLogoZize = `${props.size}`;
</script>

<style lang="scss" scoped>
.my-empty-contianer {
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
