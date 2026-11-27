<template>
	<adaptive-layout>
		<template v-slot:pc>
			<layout-pc />
		</template>
		<template v-slot:mobile>
			<layout-mobile />
			<landscape-screen-index v-if="showLandscapeBlock" />
		</template>
	</adaptive-layout>
</template>

<script lang="ts" setup>
import LayoutPc from './LayoutPc.vue';
import LayoutMobile from './LayoutMobile.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import LandscapeScreenIndex from 'src/apps/Desktop/LandscapeScreenIndex.vue';
import { useLandscapeScreen } from 'src/composables/useLandscapeScreen';

const { showLandscapeBlock } = useLandscapeScreen();
</script>
