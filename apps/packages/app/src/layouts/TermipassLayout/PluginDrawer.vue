<template>
	<q-footer class="drawer-footer-wrapper">
		<q-list class="pluginDrawer row items-center justify-around">
			<template v-for="(menu, index) in menus" :key="index">
				<div
					clickable
					class="column items-center justify-center menu-item"
					:class="menu.label === activeItem ? 'active-item' : ''"
					@click="handleItem(menu.label)"
				>
					<div class="icon-wrap q-mb-xs row items-center justify-center">
						<img :src="showMenuIcon(menu.label)" />
					</div>
					<div
						class="item-text text-body3 q-mb-md"
						:class="menu.label == activeItem ? 'text-ink-1' : 'text-ink-3'"
					>
						{{ menu.label }}
					</div>
				</div>
			</template>
		</q-list>
	</q-footer>
</template>

<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';
// import { useMenuStore } from '../../stores/files-menu';
// import { useUserStore } from '../../stores/user';
import { useRoute, useRouter } from 'vue-router';
import { getPluginImageUrl } from '../../stores/bex-url';

// import { useBexStore } from '../../stores/bex';

// const menuStore = useMenuStore();
// const userStore = useUserStore();
const router = useRouter();
const route = useRoute();

// const bexStore = useBexStore();

const menus = ref([
	// {
	// 	icon: 'sym_r_apps',
	// 	identify: 'home',
	// 	label: 'Chat'
	// },
	{
		icon: 'sym_r_apps',
		identify: 'collect',
		label: 'Collect'
	},
	// {
	// 	icon: 'sym_r_apps',
	// 	identify: 'read',
	// 	label: 'Read'
	// },
	{
		icon: 'sym_r_apps',
		identify: 'vault',
		label: 'Vault'
	},
	{
		icon: 'settings',
		identify: 'settings',
		label: 'Setting'
	}
]);
const activeItem = ref<string | null>('Chat');

onMounted(async () => {});

const showMenuIcon = (name: string) => {
	if (activeItem.value === name) {
		return getPluginImageUrl(`${name}-active.svg`);
	} else {
		return getPluginImageUrl(`${name}.svg`);
	}
};

const handleItem = (label: string) => {
	activeItem.value = label;
	switch (label) {
		case 'Chat':
			router.push({
				path: '/home'
			});
			break;

		case 'Vault':
			router.push({
				path: '/items'
			});
			break;
		case 'Collect':
			router.push({
				path: '/collect'
			});
			break;

		case 'Setting':
			router.push({
				path: '/setting'
			});
			break;

		default:
			break;
	}
};

watch(
	() => route.meta,
	() => {
		if (!route.meta || !(route.meta as any).tabIdentify) {
			activeItem.value = '';
			return;
		}
		activeItem.value =
			menus.value.find((e) => e.identify === (route.meta as any).tabIdentify)
				?.label || '';
	},
	{
		immediate: true
	}
);

// const showSettings = () => {
// 	router.push({
// 		path: `${route.meta.root}/setting`
// 	});
// };

// const showAcounts = () => {
// 	router.push({
// 		path: `${route.meta.root}/accounts`
// 	});
// };

// const toggleBex = () => {
// 	bexStore.controller.toggleBexDisplay();
// };
</script>

<style lang="scss">
.q-drawer {
	background-color: transparent;
}
</style>

<style lang="scss" scoped>
.plugin-header-container {
	background: linear-gradient(
			180deg,
			rgba(255, 249, 199, 0.3) 1.14%,
			rgba(255, 215, 73, 0.3) 99.98%
		),
		linear-gradient(
			220deg,
			rgba(255, 255, 243, 0.9) 0%,
			rgba(255, 233, 115, 0.9) 20.35%,
			rgba(251, 251, 233, 0.9) 51.21%,
			rgba(255, 255, 255, 0.9) 81.4%,
			rgba(254, 251, 228, 0.9) 101.75%
		);
	backdrop-filter: blur(50px);
	.avatar-container {
		cursor: pointer;
	}
}
.setting-icon-container {
	cursor: pointer;
	display: flex;
	justify-content: center;
	align-items: center;
	padding: 10px;
	background: rgba(255, 255, 255, 0.6);
	border-radius: 50%;
}
.menu-item {
	cursor: pointer;
	color: $ink-3;

	.icon-wrap {
		width: 20px;
		height: 20px;
		border-radius: 8px;

		img {
			width: 20px;
			height: 20px;
		}
	}
}

.active-item {
	.icon-wrap {
		background-color: $background-1;
	}

	.item-text {
		color: $ink-1;
	}
}

.pluginDrawer {
	padding: 6px 0;
}

.menu-footer {
	// position: absolute;
	// bottom: 32px;

	.settings {
		margin-bottom: 20px;
	}
}
.drawer-footer-wrapper {
	border-top: 1px solid $background-5;
	background-color: $background-1;
	// backdrop-filter: blur(27.182817459106445px);
}
.drawer-wrapper {
	& ::v-deep(.q-drawer) {
		background: transparent;
	}
}
</style>
