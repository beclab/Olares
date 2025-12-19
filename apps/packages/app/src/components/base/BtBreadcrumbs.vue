<template>
	<div
		class="breadcrumbs-root row justify-start items-center"
		:style="{ '--margin': margin ? margin : '50px' }"
	>
		<div class="text-h6 row text-link-1 q-mr-lg">
			<div class="row items-center" v-if="!breadcrumb">
				<q-icon class="icon-background q-mr-sm" size="24px" :name="icon" />
				<q-breadcrumbs-el
					class="text-h6 text-link-1 menu-page-title"
					:label="title"
				/>
				<slot name="more" />
			</div>
			<q-breadcrumbs active-color="orange-default" v-else>
				<slot name="breadcrumb" />
			</q-breadcrumbs>
		</div>
		<q-scroll-area
			v-if="endSlot"
			class="breadcrumbs-end row"
			style="flex: 1; text-align: center"
		>
			<slot name="end" />
		</q-scroll-area>
	</div>
</template>

<script setup lang="ts">
import { useSlots } from 'vue';

defineProps({
	icon: {
		type: String,
		require: true
	},
	title: {
		type: String,
		require: true
	},
	margin: {
		type: String,
		require: false
	},
	breadcrumb: {
		type: Boolean,
		default: false
	}
});

const endSlot = useSlots().end;
</script>

<style scoped lang="scss">
.breadcrumbs-root {
	max-width: calc(100% - var(--margin));
	width: 100%;
	height: 56px;

	.menu-page-title {
		max-width: 240px;
		overflow: hidden;
		text-overflow: ellipsis;
		display: -webkit-box;
		-webkit-line-clamp: 1;
		-webkit-box-orient: vertical;
	}

	.icon-background {
		margin-left: 44px;
	}

	.breadcrumbs-end {
		height: 56px;
	}
}
</style>
