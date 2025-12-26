<template>
	<page-container :title-height="56">
		<template v-slot:title>
			<title-bar :show="true" @onReturn="router.back()" />
		</template>
		<template v-slot:page>
			<div class="preferences-scroll">
				<app-store-body :title="t('Settings')" :title-separator="true">
					<template v-slot:body>
						<div class="text-h6 text-ink-1 q-mt-xl">
							{{ t('Market Source') }}
						</div>
						<div class="text-body2 text-ink-1 q-mt-md">
							{{
								t(
									'Choose a remote market source to retrieve application information.'
								)
							}}
						</div>

						<div class="q-my-md">
							<market-source-item
								v-for="item in centerStore.remoteSource"
								:key="item.id"
								:source="item"
								:model-value="settingStore.marketSourceId == item.id"
								class="q-mb-md not:first-child"
							/>
						</div>

						<request-btn
							class="q-mt-xs"
							color="blue-default"
							:label="t('Add Source')"
							:loading="clearLoading"
							@request="addSource"
						/>

						<div class="text-ink-1 text-h6 q-mt-xl">
							{{ t('about') }}
						</div>
						<div class="text-ink-2 text-body2 q-mt-md">
							{{ t('preferences.current_version', { version: versionRef }) }}
						</div>
					</template>
				</app-store-body>
			</div>
		</template>
	</page-container>
</template>

<script setup lang="ts">
import MarketSourceItem from '../../../components/appcard/MarketSourceItem.vue';
import PageContainer from '../../../components/base/PageContainer.vue';
import AppStoreBody from '../../../components/base/AppStoreBody.vue';
import RequestBtn from '../../../components/rss/RequestBtn.vue';
import TitleBar from '../../../components/base/TitleBar.vue';
import AddSourceDialog from './AddSourceDialog.vue';
import { useSettingStore } from '../../../stores/market/setting';
import { useCenterStore } from '../../../stores/market/center';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { ref } from 'vue';

const $q = useQuasar();
const { t } = useI18n();
const router = useRouter();
const clearLoading = ref(false);
const centerStore = useCenterStore();
const settingStore = useSettingStore();
const versionRef = ref(process.env.APP_VERSION);

const addSource = () => {
	$q.dialog({
		component: AddSourceDialog
	});
};
</script>

<style scoped lang="scss">
.preferences-scroll {
	height: calc(100vh - 56px);
	padding: 0 44px;
	width: 528px;
}
</style>
