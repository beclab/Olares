<template>
	<div class="menu-active">
		<q-scroll-area
			ref="scrollVaultMenuRef"
			style="height: 100%"
			:thumb-style="scrollBarStyle.thumbStyle"
			@scroll="getScroll"
		>
			<bt-menu
				:modelValue="store.currentItem"
				:items="store.menus"
				@select="selectHandler"
				style="width: 240px"
				active-class="text-subtitle2 bg-yellow-soft text-ink-1"
				class="text-ink-2"
				:key="locale"
			>
			</bt-menu>
		</q-scroll-area>

		<div class="row q-py-sm bottomBar">
			<q-icon
				v-if="store.syncInfo.syncing"
				class="q-ml-md q-mr-sm rotate"
				name="sym_r_progress_activity"
				size="24px"
				color="green"
			/>
			<q-icon
				v-else
				class="q-ml-md q-mr-sm cursor-pointer"
				name="sym_r_refresh"
				size="24px"
				@click="sync"
			>
				<q-tooltip class="bg-grey text-caption" :offset="[0, 0]">{{
					t('refresh')
				}}</q-tooltip>
			</q-icon>

			<span
				class="row items-center justify-center text-caption text-green"
				v-if="store.syncInfo.syncing"
			>
				{{ t('syncing') }}
			</span>
			<span class="row items-center justify-center text-caption" v-else>
				{{
					_t('last_sync_time', {
						time: store.syncInfo.lastSyncTime
					})
				}}
			</span>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { app, OrgMenu } from '../../globals';
import { useMenuStore } from '../../stores/menu';
import { scrollBarStyle, VaultMenuItem } from 'src/utils/contact';
import { useI18n } from 'vue-i18n';
import { getAppPlatform } from '../../application/platform';
import { computed } from 'vue';
import { useDeviceStore } from '../../stores/device';
import { _t } from '../../utils/i18n';
// const $q = useQuasar();
const Router = useRouter();
const deviceStore = useDeviceStore();

// const isMobile = ref(
// 	(process.env.PLATFORM == 'MOBILE' && ||
// 		process.env.IS_BEX
// );

const store = useMenuStore();
// const userStore = useUserStore();
// const current_user = ref(userStore.current_user);
const scrollVaultMenuRef = ref();
const { t, locale } = useI18n();

const selectHandler = (value) => {
	store.currentItem = value.item.key;
	if (isLeftDrawerOpen.value) {
		store.leftDrawerOpen = !store.leftDrawerOpen;
	}

	if (
		value.item.key === VaultMenuItem.SECURITYREPORT ||
		value.item.key === VaultMenuItem.PASSWORDGENERATOR ||
		value.item.key === VaultMenuItem.SETTINGS
	) {
		selectTools(value.key);
	} else if (value.item.key === VaultMenuItem.LOCKSCREEN) {
		lock();
	} else if (value.item && value.item.org_id) {
		selectOrgMenu(value.item.org_id, value.item.key);
	} else if (value.item && value.item.orgId) {
		gotoInvited(value.item, OrgMenu.INVITES);
	} else if (value.item && value.item.vaultId) {
		changeItemMenu(value.item.vaultId);
	} else {
		changeItemMenu();
	}
};

const isLeftDrawerOpen = computed(function () {
	if (process.env.PLATFORM == 'MOBILE') {
		if (getAppPlatform().isPad && deviceStore.isLandscape) {
			return false;
		}
		return true;
	}
	if (process.env.IS_BEX) {
		return true;
	}
	return false;
});

onMounted(async () => {
	scrollVaultMenuRef.value.setScrollPosition(
		'vertical',
		store.verticalPosition
	);
	store.updateMenuInfo();
});

function goto(path: string) {
	Router.push({
		path: path
	});
}

function lock() {
	store.clear();
	app.lock();
	Router.replace({
		path: '/unlock'
	});
}

async function sync() {
	store.handleSync();
}

const changeItemMenu = (vaultID = ''): void => {
	if (vaultID) {
		store.changeItemMenu(vaultID);
		store.currentItem = 'vault';
	} else {
		store.vaultId = '';
	}
	goto('/items/');
};

const selectOrgMenu = (org_id: string, menu: OrgMenu): void => {
	store.selectOrgMenu(org_id, menu);
	goto('/org/' + menu);
};

const gotoInvited = (invite: any, menu: OrgMenu) => {
	goto('/invite-recipient/' + invite.orgId + '/' + invite.id);
	store.selectOrgMenu(invite.orgId, menu);
};

const selectTools = (menu: string) => {
	store.clear();
	let path = menu;
	if (menu === VaultMenuItem.PASSWORDGENERATOR) {
		path = 'generator';
	} else if (menu === VaultMenuItem.SECURITYREPORT) {
		path = 'security';
	}
	goto(`/${path}`);
};

const getScroll = (info: any) => {
	store.verticalPosition = info.verticalPosition;
};
</script>

<style lang="scss" scoped>
.menu-active {
	width: 240px;
	height: 100%;
	padding-bottom: 42px;
	overflow: hidden;
}
.logo {
	width: 100%;
	height: 80px;
	display: flex;
	align-items: center;
	justify-center: justify-between;

	.name,
	.did {
		width: 150px;
		line-height: 24px;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
}

.bottomBar {
	border-top: 1px solid $separator;
}

.rotate {
	animation: aniRotate 0.8s linear infinite;
	&:hover {
		background: transparent !important;
	}
}

@keyframes aniRotate {
	0% {
		transform: rotate(0deg);
	}
	50% {
		transform: rotate(180deg);
	}
	100% {
		transform: rotate(360deg);
	}
}

.expArrow {
	transform: rotate(0deg);
	animation: rotate0 0.3s linear forwards;
}

.showExp {
	animation: rotate 0.3s linear forwards;
}

@keyframes rotate0 {
	0% {
		transform: rotate(0deg);
	}
	100% {
		transform: rotate(180deg);
	}
}

@keyframes rotate {
	0% {
		transform: rotate(180deg);
	}
	100% {
		transform: rotate(0deg);
	}
}
</style>
