<template>
	<MyBottomSheetPage :title="$t(routeMeta.title)" @close="closeHandler">
		<router-view></router-view>
	</MyBottomSheetPage>
</template>

<script lang="ts" setup>
import MyBottomSheetPage from '../pages/Plugin/components/MyBottomSheetPage.vue';
import { useRoute, useRouter } from 'vue-router';
const route = useRoute();
const router = useRouter();
const routeMeta = route.meta as Record<string, string>;
const closeHandler = () => {
	router.replace({
		path: routeMeta.roots
	});
};
</script>
