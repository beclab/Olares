<template>
	<q-layout
		class="main-layout row items-center justify-evenly"
		:class="$q.dark.isActive ? 'main-layout-dark-bg' : 'main-layout-light-bg'"
	>
		<div class="settings_box">
			<div class="settings_left">
				<bt-scroll-area class="full-height">
					<bt-menu
						active-class="my-active-link"
						:items="itemsRef"
						v-model="itemMenu"
						@select="changeItemMenu"
					>
						<template v-slot:header>
							<q-item
								:clickable="itemMenu !== '/'"
								:active="itemMenu === '/'"
								@click="changeItemMenu({ key: MENU_TYPE.Root })"
								active-class="my-active-link"
								class="person-item row justify-start items-center"
							>
								<setting-avatar :size="40" style="margin-left: 8px" />
								<div
									class="column justify-between"
									style="margin-left: 8px; max-width: calc(100% - 60px)"
								>
									<div
										class="text-subtitle1 person-text"
										:class="
											itemMenu === '/' ? 'text-blue-default' : 'text-ink-1'
										"
									>
										{{ adminStore.user.name }}
									</div>
									<div
										class="text-body3 person-text"
										:class="
											itemMenu === '/' ? 'text-blue-default' : 'text-ink-2'
										"
									>
										{{ '@' + adminStore.olaresId.split('@')[1] }}
									</div>
								</div>
							</q-item>
						</template>
					</bt-menu>
				</bt-scroll-area>
			</div>
			<div class="settings_content">
				<q-page-container class="settings_content_view">
					<router-view />
				</q-page-container>
			</div>
		</div>
	</q-layout>
</template>

<script lang="ts" setup>
import SettingAvatar from 'src/components/settings/base/SettingAvatar.vue';
import { useBackgroundStore } from 'src/stores/settings/background';
import { useAdminStore } from 'src/stores/settings/admin';
import { onMounted, ref, watch } from 'vue';
import { useRoute, useRouter, isNavigationFailure } from 'vue-router';
import { MENU_TYPE } from 'src/constant';
import { useQuasar } from 'quasar';

const router = useRouter();
const route = useRoute();
const $q = useQuasar();
const adminStore = useAdminStore();
const backgroundStore = useBackgroundStore();

const itemsRef = ref();
const itemMenu = ref('/');

function syncItemMenuFromRoute() {
	if (!itemsRef.value?.[0]?.children?.length) {
		return;
	}
	const menuKeys = new Set(Object.values(MENU_TYPE) as string[]);
	const menuRoutes = router
		.getRoutes()
		.filter((r) => typeof r.name === 'string' && menuKeys.has(r.name))
		.map((r) => ({ name: r.name as string, path: r.path.toLocaleLowerCase() }))
		.sort((a, b) => b.path.length - a.path.length);

	const currentPath = route.path.toLocaleLowerCase();
	const matched = menuRoutes.find(
		(r) => currentPath === r.path || currentPath.startsWith(`${r.path}/`)
	);

	if (matched?.name === MENU_TYPE.Root) {
		itemMenu.value = '/';
		return;
	}

	if (matched?.name) itemMenu.value = matched.name;
}

const changeItemMenu = (data: any): void => {
	const type = data.key;
	itemMenu.value = type === MENU_TYPE.Root ? '/' : type;
	router.push({ name: type }).then((result) => {
		if (isNavigationFailure(result)) {
			syncItemMenuFromRoute();
		}
	});
};

if (location.pathname === '/') {
	itemMenu.value = '/';
	router.push({ path: '/' });
}

onMounted(() => {
	configMenus();
});

const configMenus = () => {
	itemsRef.value = [
		{
			label: '',
			key: 'Settings',
			children: adminStore.menus.flat()
		}
	];

	if (itemsRef.value.length > 0) {
		syncItemMenuFromRoute();
	}
};

watch(() => route.fullPath, syncItemMenuFromRoute);

watch(
	() => backgroundStore.locale,
	() => {
		configMenus();
	}
);
</script>

<style lang="scss" scoped>
.settings_box {
	width: 848px;
	// max-width: 848px;
	// min-width: 848px;
	height: 100vh;
	display: flex;
	align-content: center;
	justify-content: center;
	border-radius: 8px;
	overflow: hidden;

	.settings_left {
		width: 240px;
		height: 100%;
		border-right: $separator;
		border-right-width: 1px;
		border-right-style: solid;

		.person-item {
			height: 48px;
			max-width: 100%;
			padding: 0;
			border-radius: 8px;
			.person-text {
				text-overflow: ellipsis;
				white-space: nowrap;
				overflow: hidden;
				max-width: 100%;
			}
		}
	}

	.settings_content {
		width: calc(100% - 240px);
		//padding-bottom: 20px;
		height: 100%;

		.settings_content_view {
			overflow: hidden;
			height: 100%;
			width: 100%;
			padding: 0;
		}
	}
}

.main-layout::v-deep .my-active-link {
	color: $theme-menu-color-hover;
	background-color: $theme-menu-bg-hover;
}

.main-layout-dark-bg {
	background-color: $theme-background-dark;
}

.main-layout-light-bg {
	background-color: $theme-background-light;
}
</style>
