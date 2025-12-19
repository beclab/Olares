<template>
	<adaptive-layout>
		<template v-slot:pc>
			<layout-pc />
		</template>
		<template v-slot:mobile>
			<layout-mobile />
		</template>
	</adaptive-layout>
</template>

<script lang="ts" setup>
import LayoutPc from './LayoutPc.vue';
import LayoutMobile from './LayoutMobile.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import 'src/css/common.scss';
</script>
