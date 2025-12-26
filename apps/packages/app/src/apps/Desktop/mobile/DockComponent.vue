<template>
	<div class="desktop-box">
		<div
			class="desktop-app-box"
			:class="{ dockAni: appStore.dockerApps.length > 0 }"
		>
			<template v-for="element in appStore.dockerApps" :key="element.id">
				<div class="launch-app column items-center">
					<div
						class="app-img"
						@click="openWindow(element)"
						:id="element.id"
						:style="{
							width: element.id === 'bdock:launchpad' ? '56px' : '52px',
							height: element.id === 'bdock:launchpad' ? '56px' : '52px'
						}"
					>
						<img :src="element.icon" />
						<div class="overlay"></div>
					</div>

					<div class="app-name q-mt-xs" v-if="element.id !== 'bdock:launchpad'">
						{{ element.title }}
					</div>
				</div>
			</template>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { DockerAppInfo, AppClickInfo } from '../type/types';
import { useAppStore } from '../../../stores/desktop/app';

defineProps({
	drag_launch_app: {
		type: String,
		required: true
	}
});

const emits = defineEmits([]);

const appStore = useAppStore();

const openWindow = async (item: DockerAppInfo) => {
	emits('appClick', {
		appid: item.id,
		data: {}
	} as AppClickInfo);
};

const getAppIndexInDock = (id: string) => {
	let index = -1;
	for (var i = 0; i < appStore.dockerApps.length; i++) {
		const draggedEl: any = document.getElementById(appStore.dockerApps[i].id);
		if (draggedEl.id == id) {
			index = i;
		}
	}
	return index;
};

const handleOpenApp = (appid: string) => {
	let rid = appid;
	if (rid.startsWith('bdock:') == false) {
		rid = 'bdock:' + rid;
	}

	let cc = getAppIndexInDock(rid);
	if (cc < 0) {
		appStore.add_app_on_docker_bottom(rid);
	} else {
		appStore.dockerApps[cc].show_dot = true;
	}
};

defineExpose({
	handleOpenApp
});
</script>

<style lang="scss" scoped>
.desktop-box {
	.desktop-app-box {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-around;
		position: fixed;
		bottom: 76px;
		left: 0;
		right: 0;
		margin: auto;
		z-index: 9;

		.launch-app {
			.app-img {
				width: 52px;
				height: 52px;
				border-radius: 14px;
				-webkit-touch-callout: none;
				-webkit-user-select: none;
				-webkit-user-drag: none;
				overflow: hidden;
				position: relative;
				img {
					width: 100%;
					height: 100%;
				}
				.overlay {
					width: 100%;
					height: 100%;
					position: absolute;
					top: 0;
					left: 0;
					right: 0;
					bottom: 0;
					background: transparent;
				}
			}
			.app-name {
				width: 100%;
				height: 18px;
				border-radius: 14px;
				font-size: 14px;
				font-style: normal;
				font-weight: 400;
				line-height: 16px;
				color: #fff;
				text-align: center;
			}
		}
	}
}

// .dockAni {
// 	// animation: DockAni 0.5s ease-in-out;
// }

// @keyframes DockAni {
// 	from {
// 		bottom: 66px;
// 		opacity: 0;
// 	}
// 	to {
// 		bottom: 76px;
// 		opacity: 1;
// 	}
// }
</style>
