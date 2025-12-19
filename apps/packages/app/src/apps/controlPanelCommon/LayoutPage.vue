<template>
	<q-layout class="row items-center justify-evenly">
		<div class="settings_boxss">
			<div class="settings_box">
				<div class="settings_left">
					<bt-scroll-area
						class="nav-height-scroll-area-conf"
						:thumb-style="defaultScrollAreaThumbStyle"
						:bar-style="defaultScrollAreaThumbStyle"
					>
						<q-list
							class="listWidth column content-center"
							style="margin-top: 8px"
						>
							<q-item
								v-for="(itme, ci) in SettingsList"
								dense
								:key="`c` + ci"
								clickable
								:active="link.startsWith(itme.link)"
								@click="onSettingsItem(itme)"
								active-class="my-menu-link"
								class="my-menu setting-major-text"
							>
								<q-item-section>
									<div class="row items-center">
										<img
											:src="link === itme.link ? menuIPrimarycon : menuIcon"
											alt=""
										/>
										<div
											style="margin-left: 12px; font-size: 14px"
											:style="{ fontWeight: link === itme.link ? 500 : 400 }"
										>
											<div>{{ itme.name }}</div>
										</div>
									</div>
								</q-item-section>
							</q-item>
						</q-list>
					</bt-scroll-area>
				</div>
				<div class="settings_content">
					<bt-scroll-area style="height: calc(100vh)">
						<div class="title">
							{{ name }}
						</div>
						<q-page-container class="settings_content_view">
							<router-view />
						</q-page-container>
					</bt-scroll-area>
				</div>
			</div>
		</div>
	</q-layout>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import './css/common.scss';
import { defaultScrollAreaThumbStyle } from './constant';
import menuIcon from './imgs/menu.svg';
import menuIPrimarycon from './imgs/menu-primary.svg';

export interface Setting {
	name: string;
	icon: string;
	link: string;
}

export interface Props {
	data?: Setting[];
}
const props = withDefaults(defineProps<Props>(), {});

const SettingsList = props?.data ?? [];
const route = useRoute();
const router = useRouter();
const leftDrawerOpen = ref(false);
const link = ref(route.path);

const target = props.data?.find((item) => item.link === route.path);

const name = ref(target?.name);

if (location.pathname === '/') {
	link.value = '/';
	router.push({ path: '/' });
}

function onSettingsItem(item: Setting) {
	router.push({ path: `${item.link}` });
	link.value = item.link;
	name.value = item.name;
}

watch(
	() => route.path,
	() => {
		link.value = route.path;
	}
);
</script>

<style lang="scss" scoped>
$bg: rgba(34, 111, 255, 0.12);
$left-width: 208px;
.settings_boxss {
	width: 100%;
	height: 100vh;
	display: flex;
	align-content: center;
	justify-content: center;
	border-radius: 8px;
	overflow: hidden;
}

.settings_box {
	width: 100%;
	height: 100%;
	display: flex;
	overflow: hidden;
}

.settings_left {
	width: $left-width;
	padding: 16px 0;
	border-right: 1px solid #e0e0e0;

	.my-menu-link {
		background-color: $bg;
	}

	.listWidth {
		.my-menu {
			width: 192px;
			border-radius: 8px;
			overflow: hidden;
			font-weight: 400;
			font-size: 14px;
			height: 36px;
			margin-bottom: 4px;
		}
		> .q-item:hover {
			// background-color: $bg;
		}
	}
}

.settings_content {
	width: calc(100% - #{$left-width});
	height: 100%;
	.title {
		color: #1f1814;
		font-size: 24px;
		font-weight: 700;
		line-height: 32px;
		padding: 20px 32px;
	}
}

.settings_content_view {
	background: #ffffff;
	overflow: hidden;
	height: 100%;
	width: 100%;
	padding: 0 32px 20px 32px;
}
</style>
