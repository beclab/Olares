<template>
	<router-view></router-view>
</template>

<script setup lang="ts">
// import PluginGoogleSearch from '../pages/Plugin/GoogleSearch.vue';
// import PluginYoutebuSearch from 'pages/Plugin/YoutebuSearch.vue';
import { useQuasar } from 'quasar';
import { ref, watchEffect, onMounted } from 'vue';
import { useUserStore } from '../stores/user';
import { setAssetsUrl } from '../stores/bex-url';
import urlUtils from '../utils/url';
import { updateUIToAddWeb } from '../platform/addItem';
import { useRouter } from 'vue-router';
import { browser } from 'src/platform/interface/bex/browser/target';
import { OverlayBackgroundExtensionMessage } from 'src/platform/interface/bex/autofill2/background/abstractions/overlay.background';
import { useBexStore } from '../stores/bex';
import { useApproval } from '../pages/Mobile/wallet/approval';
import { storage } from 'src/utils/bex/storage';
const userStore = useUserStore();
const searchParamsObj = urlUtils.getSearchParamsObj();

const $q = useQuasar();
const show = ref(false);
const extraShow = ref(false);
const router = useRouter();

// const isnotification = ref(searchParamsObj.notification);
const toggle = () => {
	show.value = !show.value;
};
let lasteventResponseKey = '';
$q?.bex?.on(
	'webos.app.status',
	(data: {
		data: { show?: boolean };
		respond: any;
		eventResponseKey: string;
	}) => {
		if (lasteventResponseKey == data.eventResponseKey) {
			return;
		}
		lasteventResponseKey = data.eventResponseKey;
		if (data.data.show != undefined) {
			show.value = data.data.show;
		} else {
			toggle();
		}
		data.respond();
	}
);

browser.runtime.onMessage.addListener(
	(message: OverlayBackgroundExtensionMessage) => {
		console.log('fjeowjgoejwg-2', message);
		autoFillHandler(message);
	}
);

async function autoFillHandler(message: OverlayBackgroundExtensionMessage) {
	switch (message?.type) {
		case 'frontUnlockVault':
			router.push({
				path: '/unlock'
			});
			storage.remove('autofillVaultMessage');
			break;
		case 'frontAddNewVaultItem':
			updateUIToAddWeb(
				message.url,
				router,
				message.username,
				message.password,
				false
			);
			storage.remove('autofillVaultMessage');
			break;
		case 'notification':
			{
				dealNotification();
			}
			storage.remove('autofillVaultMessage');
			break;
		default:
			storage.remove('autofillVaultMessage');
			break;
	}
}

const dealNotification = async () => {
	const store = useBexStore();
	const approval = await store.controller.getApproval();

	if (approval?.data.routerPath == '/unlock') {
		userStore.unlockFirst(() => {
			const { resolveApproval } = useApproval(router);
			setTimeout(() => {
				resolveApproval();
			}, 1000);
		}, true);
		return;
	}

	if (approval && approval.data.routerPath) {
		router.push({
			path: approval.data.routerPath,
			query: {
				type: 'notification'
			}
		});
	}
};

onMounted(async () => {
	const store = useBexStore();
	const approval = await store.controller.getApproval();
	if (approval) {
		dealNotification();
	} else {
		const params = await storage.get('autofillVaultMessage');
		console.log('fjeowjgoejwg', params);
		autoFillHandler(params);
	}
});

const initExtensionURL = async () => {
	const data = await $q?.bex?.send('webos.app.url', {
		init: true
	});
	setAssetsUrl(data.data);
};

initExtensionURL();

watchEffect(() => {
	extraShow.value =
		!!userStore.current_id &&
		!!userStore?.current_user?.access_token &&
		!!searchParamsObj.q;
});

// const asideWidth = '48px';
// const asideStyle: any = {
// 	position: 'absolute',
// 	top: '50%',
// 	left: `calc(-1 * ${asideWidth})`,
// 	width: asideWidth,
// 	height: '40px',
// 	cursor: 'pointer',
// 	display: 'flex',
// 	alignItems: 'center',
// 	borderRadius: '20px 0 0 20px',
// 	background: '#fff'
// };

// const asideIconStyle: any = {
// 	width: '32px',
// 	height: '32px',
// 	display: 'flex',
// 	alignItems: 'center',
// 	justifyContent: 'center',
// 	borderRadius: '50%',
// 	marginLeft: '4px'
// };
</script>

<style></style>
<style lang="scss" scoped>
.webos-app-container {
	background: $background-1;
	z-index: 70000;
	filter: drop-shadow(0px 0px 12px rgba(0, 0, 0, 0.12));
	text-align: left;
}
</style>
