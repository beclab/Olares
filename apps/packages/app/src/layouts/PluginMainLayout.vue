<template>
	<q-layout view="lhr lpr lfr" class="lauout lauout-app">
		<div class="mainlayout" :style="{ width: isBex ? '100%' : '100vw' }">
			<q-page-container style="width: 100%; height: 100%">
				<TerminusUserHeaderReminder />
				<div class="row no-wrap bg-background-1">
					<div style="flex: 1; height: 100vh; overflow: hidden">
						<router-view />
					</div>
					<tabbar-component
						v-if="tabbarShow"
						class="tabbar"
						:class="$q.platform.is.ios ? 'tabbar-ios' : ''"
						:current="defaultIndex"
						@update-current="updateCurrent"
					>
						<template #footer>
							<q-img
								:src="settingsIcon"
								:ratio="1"
								spinner-size="0px"
								width="20px"
								class="q-mb-xs cursor-pointer"
								@click="linkToSetting(ROUTE_CONST.OPTIONS_ACCOUNT, router)"
							/>
							<div class="q-mt-md">
								<TerminusAccountAvatar class="q-mt-xs" />
							</div>
						</template>
					</tabbar-component>
				</div>
			</q-page-container>
		</div>
	</q-layout>
</template>

<script lang="ts" setup>
import TabbarComponent from 'src/components/common/TerminusTabbarPluginComponent.vue';
import { useMobileMainLayout } from 'src/composables/mobile/useMobileMainLayout';
import TerminusAccountAvatar from 'src/components/common/TerminusAccountPluginAvatar.vue';
import TerminusUserHeaderReminder from 'src/components/common/TerminusUserHeaderReminder.vue';
import settingsIcon from 'src/assets/plugin/settings.svg';
import { linkToSetting } from 'src/utils/bex/link';
import { ROUTE_CONST } from 'src/router/route-const';
import '../css/terminus.scss';
import { useRouter } from 'vue-router';
import { onMounted } from 'vue';
import { useLarepassWebsocketManagerStore } from 'src/stores/larepassWebsocketManager';
const socketStore = useLarepassWebsocketManagerStore();

const router = useRouter();

onMounted(() => {
	socketStore.restart();
});
const { menuStore, isBex, tabbarShow, defaultIndex, updateCurrent } =
	useMobileMainLayout();
</script>

<style lang="scss" scoped></style>
