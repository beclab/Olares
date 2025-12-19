<template>
	<div class="preview-layout" v-if="userStore.user">
		<preview-layout
			:style="{
				height:
					userStore.isMobile && route.path === '/preview'
						? 'calc(100% - 64px)'
						: '100vh'
			}"
			:font-size="16"
			:user="userStore.user"
			:pc="platform.is.desktop"
		>
			<div
				v-if="userStore.isMobile && route.path === '/preview'"
				class="close-icon column justify-center items-center"
			>
				<q-icon
					size="24px"
					class="text-ink-1 cursor-pointer"
					name="sym_r_close"
					@click="router.back()"
				/>
			</div>
		</preview-layout>
		<bio-share-header v-if="userStore.isMobile && route.path === '/preview'" />
	</div>
</template>

<script lang="ts" setup>
import PreviewLayout from '@apps/profile/src/components/profile/layout/PreviewLayout.vue';
import BioShareHeader from '@apps/profile/src/components/profile/share/BioShareHeader.vue';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { useQuasar } from 'quasar';
import { useRoute, useRouter } from 'vue-router';
const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
// const currentView = ref();

// if (userStore.user && userStore.user.background.style === 0) {
//   currentView.value = BioLayoutComponent1;
// }
const { platform } = useQuasar();

console.log(!!platform.is.desktop);
</script>

<style lang="scss" scoped>
.preview-layout {
	width: 100vw;
	height: 100vh;

	.close-icon {
		z-index: 1;
		position: absolute;
		height: 40px;
		width: 40px;
		right: 12px;
		top: 12px;
		background: #ffffff99;
		border-radius: 50%;
	}
}
</style>
