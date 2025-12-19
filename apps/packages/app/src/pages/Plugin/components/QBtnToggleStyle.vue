<template>
	<span class="q-button-style-contianer" :class="`${size}-style-wrapper`">
		<slot></slot>
	</span>
</template>
<script setup lang="ts">
interface Props {
	size: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
}

withDefaults(defineProps<Props>(), {
	size: 'xl'
});
</script>

<style lang="scss" scoped>
.q-button-style-contianer {
	&.xl-style-wrapper {
		::v-deep(.q-btn .q-btn__content .q-icon) {
			width: 20px;
			height: 20px;
			line-height: 20px;
			font-size: map-get($map: $h5, $key: size);
		}

		::v-deep(.q-btn--dense) {
			padding: 6px;
		}
	}
	::v-deep(.q-btn--outline:before) {
		border-color: $grey-2;
	}
	&.md-style-wrapper {
		::v-deep(.q-btn .q-btn__content .q-icon) {
			width: 16px;
			height: 16px;
			line-height: 16px;
			font-size: map-get($map: $body3, $key: size);
		}
		::v-deep(.q-btn) {
			min-height: 16px;
		}
		::v-deep(.q-btn .q-btn__content) {
			width: 16px;
			height: 16px;
			line-height: 16px;
		}

		::v-deep(.q-btn--dense) {
			padding: 0;
		}
	}
}
</style>
