<template>
	<adaptive-layout>
		<template v-slot:pc>
			<q-layout class="main-layout" view="lHh Lpr lFf">
				<q-drawer
					v-if="showMenu"
					v-model="menuStore.leftDrawerOpen"
					show-if-above
					:bordered="!globalConfig.isOfficial"
					:style="globalConfig.isOfficial ? 'padding-left: 20px' : ''"
					height="100%"
					:breakpoint="0"
					:width="globalConfig.isOfficial ? 260 : 240"
				>
					<bt-scroll-area
						:style="
							globalConfig.isOfficial
								? 'height : 100dvh'
								: 'height: calc(100dvh - 65px)'
						"
						@scroll="onScroll"
					>
						<bt-menu
							active-class="my-active-link"
							:items="itemsRef"
							v-model="menuStore.currentItem"
							@select="changeItemMenu"
						>
							<template v-if="globalConfig.isOfficial" v-slot:header>
								<div class="header-bar column justify-start q-px-md q-mb-xs">
									<q-img
										class="header-icon"
										src="market/icons/favicon-128x128.png"
									>
										<template v-slot:loading>
											<q-skeleton
												class="header-icon"
												style="border-radius: 20px"
											/>
										</template>
									</q-img>
									<span class="text-h5 text-grey-10 q-mt-md">{{
										t('main.terminus_market')
									}}</span>
								</div>
							</template>

							<!--					<template-->
							<!--						v-slot:[`icon-${menu.key}`]-->
							<!--						v-for="menu in itemsRef[0].children"-->
							<!--						:key="menu.key"-->
							<!--					>-->
							<!--						<div class="custom-icon-div">-->
							<!--							<img-->
							<!--								:src="showIconAddress(menu.icon)"-->
							<!--								:alt="menu.label"-->
							<!--								:class="menuStore.currentItem === menu.key ? 'active-icon' : ''"-->
							<!--							/>-->
							<!--						</div>-->
							<!--					</template>-->
						</bt-menu>
					</bt-scroll-area>
					<div
						v-if="!globalConfig.isOfficial"
						class="bottom-menu-root items-center"
						:style="{
							'--showShadow': showBarShadow ? '1px solid #EBEBEB' : 'none'
						}"
					>
						<q-item
							dense
							clickable
							:active="menuStore.currentItem === TRANSACTION_PAGE.MyTerminus"
							active-class="my-active-link"
							class="text-ink-2 bottom-menu row justify-start items-center cursor-pointer"
							@click="changeItemMenu({ key: TRANSACTION_PAGE.MyTerminus })"
						>
							<q-icon name="sym_r_home" size="20px" />
							<div class="text-body1 q-ml-sm">
								{{ t('main.my_terminus') }}
							</div>
						</q-item>
					</div>
				</q-drawer>

				<q-page-container>
					<router-view v-slot="{ Component }">
						<transition
							:name="transitionName"
							:style="{
								width: showMenu
									? globalConfig.isOfficial
										? 'calc(100% - 260px)'
										: 'calc(100% - 240px)'
									: '100%',
								position: 'absolute'
							}"
						>
							<keep-alive :exclude="keepAliveExclude" max="10">
								<component
									:is="Component"
									style="overflow-y: hidden"
									:key="route.fullPath"
								/>
							</keep-alive>
						</transition>
					</router-view>
				</q-page-container>

				<q-btn
					v-if="globalConfig.isOfficial"
					class="btn-size-md float-btn"
					@click="installOS"
					>{{ t('main.install_terminus_os') }}
				</q-btn>
			</q-layout>
		</template>

		<template v-slot:mobile>
			<q-layout class="main-layout">
				<q-drawer
					v-if="showMenu"
					v-model="menuStore.leftDrawerOpen"
					show-if-above
					bordered
					height="100%"
					:width="240"
				>
					<bt-scroll-area
						style="height: calc(100dvh - 65px)"
						@scroll="onScroll"
					>
						<bt-menu
							active-class="my-active-link"
							:items="itemsRef"
							v-model="menuStore.currentItem"
							@select="changeItemMenu"
						>
							<template v-if="globalConfig.isOfficial" v-slot:header>
								<div class="header-bar-mobile row justify-center items-center">
									<q-img
										class="header-icon"
										src="market/icons/favicon-128x128.png"
									>
										<template v-slot:loading>
											<q-skeleton
												class="header-icon"
												style="border-radius: 20px"
											/>
										</template>
									</q-img>
									<span class="text-h5 text-grey-10 q-ml-md">{{
										globalConfig.isOfficial
											? t('main.terminus_market')
											: t('market')
									}}</span>
								</div>
							</template>
						</bt-menu>
					</bt-scroll-area>
					<div
						v-if="!globalConfig.isOfficial"
						class="bottom-menu-root items-center"
						:style="{
							'--showShadow': showBarShadow ? '1px solid #EBEBEB' : 'none'
						}"
					>
						<q-item
							dense
							clickable
							:active="menuStore.currentItem === TRANSACTION_PAGE.MyTerminus"
							active-class="my-active-link"
							class="text-ink-2 bottom-menu row justify-start items-center cursor-pointer"
							@click="changeItemMenu({ key: TRANSACTION_PAGE.MyTerminus })"
						>
							<q-icon name="sym_r_home" size="20px" />
							<div class="text-body1 q-ml-sm">
								{{ t('main.my_terminus') }}
							</div>
						</q-item>
					</div>
					<div v-else class="row justify-between items-center q-mx-lg">
						<div class="text-subtitle2 text-ink-2">{{ t('language') }}</div>
						<div class="row bg-background-3 q-pa-xs" style="border-radius: 4px">
							<template v-for="item in supportLanguages" :key="item.value">
								<div
									:class="
										settingStore.currentLanguage === item.value
											? 'language-select'
											: 'language-default'
									"
									class="text-subtitle2 column justify-center items-center"
									@click="onLanguagechange(item.value)"
								>
									<span>{{ item.label }}</span>
								</div>
							</template>
						</div>
					</div>
				</q-drawer>

				<q-page-container>
					<router-view v-slot="{ Component }">
						<transition
							:name="transitionName"
							:style="{
								width: '100%',
								position: 'absolute'
							}"
						>
							<keep-alive :exclude="keepAliveExclude" max="10">
								<component
									:is="Component"
									style="overflow-y: hidden"
									:key="route.fullPath"
								/>
							</keep-alive>
						</transition>
					</router-view>
				</q-page-container>
			</q-layout>
		</template>
	</adaptive-layout>
</template>

<script lang="ts" setup>
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import { onBeforeRouteUpdate, useRoute, useRouter } from 'vue-router';
import { useSettingStore } from '@apps/market/src/stores/market/setting';
import { useMenuStore } from '@apps/market/src/stores/market/menu';
import { useDeviceStore } from 'src/stores/settings/device';
import { useCenterStore } from 'src/stores/market/center';
import { TRANSACTION_PAGE } from '../constant/constants';
import { computed, onMounted, ref, watch } from 'vue';
import globalConfig from 'src/api/market/config';
import { SupportLanguageType } from 'src/i18n';
import { SelectorProps } from 'src/constant';
import { useI18n } from 'vue-i18n';

const keepAliveExclude = ref('LogPage');
const showBarShadow = ref(false);
const settingStore = useSettingStore();
const deviceStore = useDeviceStore();
const centerStore = useCenterStore();
const menuStore = useMenuStore();
const position = ref(-1);
const transitionName = ref();
const Router = useRouter();
const route = useRoute();
const { t } = useI18n();
const supportLanguages: SelectorProps[] = [
	{ value: 'zh-CN', label: 'ZH' },
	{ value: 'en-US', label: 'EN' }
];

const onLanguagechange = (language: SupportLanguageType) => {
	settingStore.languageUpdate(language);
	settingStore.currentLanguage = language;
};

const showMenu = computed(() => {
	if (deviceStore.isMobile) {
		return menuStore.leftDrawerOpen;
	}
	const menuParam = route.query.showMenu;
	return menuParam === undefined ? true : menuParam === 'true';
});

const itemsRef = computed(() => {
	if (globalConfig.isOfficial) {
		return [
			{
				label: t('base.extensions'),
				key: 'Application',
				children: menuStore.categoryMenu
			}
		];
	} else {
		return [
			{
				label: t('base.extensions'),
				key: 'Application',
				children: menuStore.categoryMenu
			},
			{
				label: t('manage'),
				key: 'Manage',
				children: [
					{
						label: t('search'),
						key: TRANSACTION_PAGE.Search,
						icon: 'sym_r_search'
					},
					{
						label: t('updates'),
						key: TRANSACTION_PAGE.Update,
						icon: 'sym_r_upgrade',
						count:
							centerStore.updateList.length !== 0
								? centerStore.updateList.length
								: null
					}
				]
			}
		];
	}
});

const changeItemMenu = (data: any): void => {
	const type = data.key;
	menuStore.changeItemMenu(type);
	switch (type) {
		case TRANSACTION_PAGE.All:
		case TRANSACTION_PAGE.MyTerminus:
		case TRANSACTION_PAGE.Search:
		case TRANSACTION_PAGE.Update:
			Router.push({
				name: type
			});
			break;
		default:
			Router.push({
				name: TRANSACTION_PAGE.CATEGORIES,
				params: {
					categories: type
				}
			});
			break;
	}
};

onBeforeRouteUpdate((to, from, next) => {
	transitionName.value = '';
	next();
});

watch(
	() => route.path,
	(to, from) => {
		// console.log(Router.options.history)
		if (Router.options.history.state) {
			if (
				route.name === TRANSACTION_PAGE.App ||
				route.name === TRANSACTION_PAGE.TOPIC ||
				route.name === TRANSACTION_PAGE.List ||
				route.name === TRANSACTION_PAGE.Preview ||
				route.name === TRANSACTION_PAGE.Version ||
				route.name === TRANSACTION_PAGE.Log ||
				from.includes('/app/') ||
				from.includes('/middleware/') ||
				from.includes('/recommend/') ||
				from.includes('/model/') ||
				from.includes('/discover') ||
				from.includes('/list') ||
				from.includes('/preview') ||
				from.includes('/log') ||
				from.includes('/versionHistory')
			) {
				transitionName.value =
					Number(Router.options.history.state.position) >= position.value
						? 'slide-left'
						: 'slide-right';
			} else {
				transitionName.value = '';
			}
			// console.log(`router position ${Router.options.history.state.position}`)
			// console.log(`current position ${position.value}`)
			// console.log(transitionName.value)
			position.value = Number(Router.options.history.state.position);
		}
		updateMenu();
	}
);

// nsfw restore
// watch(() => {
//   return settingStore.restore
// },(newValue) => {
//   console.log('reload')
//   if (newValue){
//     keepAliveExclude.value = 'LogPage,InstalledPage,HomePage,DiscoverPage,CategoryPage,AppListPage,AppDetailPage'
//   }else {
//     keepAliveExclude.value = 'SearchPage,LogPage';
//   }
// })

watch(
	() => {
		return menuStore.currentItem;
	},
	() => {
		if (settingStore.restore) {
			settingStore.restore = false;
		}
	}
);

onMounted(async () => {
	updateMenu();
});

const updateMenu = () => {
	switch (route.name) {
		case TRANSACTION_PAGE.CATEGORIES:
			if (route.params && route.params.categories) {
				menuStore.changeItemMenu(route.params.categories as string);
			}
			break;
		case TRANSACTION_PAGE.All:
		case TRANSACTION_PAGE.MyTerminus:
		case TRANSACTION_PAGE.Search:
		case TRANSACTION_PAGE.Update:
			menuStore.changeItemMenu(route.name);
			break;
		case TRANSACTION_PAGE.Preference:
			menuStore.changeItemMenu(TRANSACTION_PAGE.MyTerminus);
			break;
	}
};

const onScroll = async (info: any) => {
	showBarShadow.value =
		info.verticalSize > info.verticalContainerSize &&
		info.verticalPercentage !== 1;
};

const installOS = async () => {
	window.open(
		settingStore.currentLanguage === 'en-US'
			? globalConfig.install_en_docs
			: globalConfig.install_zh_docs,
		'_blank'
	);
};
</script>

<style lang="scss" scoped>
.header-bar {
	margin-top: 44px;

	.header-icon {
		width: 56px;
		height: 56px;
	}
}

.header-bar-mobile {
	.header-icon {
		width: 24px;
		height: 24px;
	}
}

.language-select {
	color: $blue-default;
	background: $background-1;
	border-radius: 4px;
	width: 56px;
	height: 32px;
}

.language-default {
	color: $ink-3;
	border-radius: 4px;
	width: 56px;
	height: 32px;
}

.main-layout {
	position: relative;
	overflow: hidden;
	width: 100dvw;
	height: 100dvh;

	.bottom-menu-root {
		border-top: var(--showShadow);
		padding: 12px 16px;

		.bottom-menu {
			height: 40px;
			border-radius: 8px;
			padding-left: 8px;
			padding-right: 8px;

			.bottom-menu-update-size {
				width: 32px;
				height: 16px;
				text-align: center;
				border-radius: 8px;
				margin-left: 8px;
			}
		}
	}

	.float-btn {
		background: $background-2;
		color: $ink-1;
		position: absolute;
		right: 24px;
		bottom: 64px;
		border: 1px solid $info;
		border-radius: 46px !important;
		text-transform: unset;
		box-shadow: 0 8px 40px 0 #00000033;
	}
}

.main-layout ::v-deep(.my-active-link) {
	color: $blue-default !important;
	background-color: $blue-soft !important;
}
</style>
