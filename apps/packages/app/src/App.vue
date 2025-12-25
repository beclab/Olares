<template>
	<bt-theme :follow-system="configStore.themeSetting === THEME_TYPE.AUTO" />
	<div
		@touchstart="startSwipe"
		@touchmove="handleSwipe"
		@touchend="endSwipe"
		v-if="$q.platform.is.ios && !$q.platform.is.ipad"
	>
		<router-view />
	</div>
	<router-view v-else />
</template>

<script lang="ts">
import { defineComponent, onMounted, onUnmounted, watchEffect } from 'vue';
import { onBeforeRouteUpdate, useRoute, useRouter } from 'vue-router';
import { useQuasar } from 'quasar';

import { useUserStore } from './stores/user';

import { ref } from 'vue';
import { watch } from 'vue';

import { useConfigStore } from './stores/rss-config';
import { THEME_TYPE } from './utils/rss-types';
import { getApplication } from './application/base';
import { login } from './utils/auth';
import { getNativeAppPlatform } from './application/platform';
//@ts-ignore
// split this style file

export default defineComponent({
	name: 'App',
	computed: {
		THEME_TYPE() {
			return THEME_TYPE;
		}
	},
	async preFetch({ redirect, currentRoute }) {
		await getApplication().appRedirectUrl(redirect, currentRoute);
		await login('', '');
	},
	setup() {
		const userStore = useUserStore();
		const $q = useQuasar();
		const route = useRoute();
		const router = useRouter();
		const configStore = useConfigStore();

		getApplication().appLoadPrepare({
			route,
			router,
			quasar: $q
		});

		onMounted(async () => {
			getApplication().appMounted();
		});

		onUnmounted(() => {
			getApplication().appUnMounted();
		});

		const transitionName = ref();

		onBeforeRouteUpdate((to, from, next) => {
			transitionName.value = '';
			next();
		});

		const position = ref(-1);

		watch(
			() => route.path,
			() => {
				if (router.options.history.state) {
					if ($q.platform.is.mobile && position.value != -1) {
						if (route.meta.tabIdentify) {
							transitionName.value =
								Number(router.options.history.state.position) >= position.value
									? ''
									: 'slide-right';
						} else {
							if (route.meta.noReturn || route.path.startsWith('/files')) {
								transitionName.value = '';
							} else {
								transitionName.value =
									Number(router.options.history.state.position) >=
									position.value
										? 'slide-left'
										: 'slide-right';
							}
						}
					}
					position.value = Number(router.options.history.state.position);
				}
			}
		);

		watchEffect(() => {
			$q?.bex?.send('webos.user.status', {
				login:
					!!userStore.current_id &&
					!!userStore.password &&
					userStore?.current_user?.access_token
			});
		});

		let startX = 0;
		let deltaX = 0;
		let isSwiping = false;

		const startSwipe = (event: any) => {
			if (!$q.platform.is.ios || !$q.platform.is.nativeMobile) {
				return;
			}
			startX = event.touches[0].clientX;
			deltaX = 0;
			if (startX > 50) {
				return;
			}
			isSwiping = true;
		};

		const handleSwipe = (event: any) => {
			if (!$q.platform.is.ios || !$q.platform.is.nativeMobile) {
				return;
			}
			if (!isSwiping) return;

			deltaX = event.touches[0].clientX - startX;

			// You can add additional logic to handle vertical swipes if needed
		};

		const endSwipe = () => {
			if (!$q.platform.is.ios || !$q.platform.is.nativeMobile) {
				return;
			}
			if (!isSwiping) return;

			if (deltaX > 30) {
				// Swipe to the right, trigger back navigation
				getNativeAppPlatform().hookBackAction();
			}

			startX = 0;
			deltaX = 0;
			isSwiping = false;
		};

		return {
			transitionName,
			startSwipe,
			handleSwipe,
			endSwipe,
			configStore
		};
	}
});
</script>

<style>
.slide-right-enter-active,
.slide-left-enter-active,
.slide-right-leave-active,
.slide-left-leave-active {
	box-shadow: -20px 0 20px 0px rgba(0, 0, 0, 0.1);
	will-change: transform;
	transition: all 0.3s ease-out;
	position: absolute;
}

.slide-right-enter-from {
	opacity: 0;
	transform: translateX(-50%);
}

.slide-right-leave-to {
	z-index: 100;
	transform: translateX(102%);
}

.slide-right-leave-from {
	box-shadow: -20px 0 20px 0px rgba(0, 0, 0, 0.1);
}

.slide-left-enter-from {
	z-index: 100;
	transform: translateX(100%);
	box-shadow: -20px 0 20px 0px rgba(0, 0, 0, 0.1);
}

.slide-left-leave-to {
	opacity: 0.4;
	transform: translateX(-50%);
}
</style>
