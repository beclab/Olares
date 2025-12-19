<template>
	<page-container>
		<template v-slot:page>
			<div class="setting-scroll">
				<app-store-body title="Settings" :title-separator="true">
					<template v-slot:body>
						<div class="setting-label">NSFW settings</div>
						<q-checkbox
							:disable="loading"
							dense
							size="30px"
							label="Blocking NSFW applications"
							v-model="nsfw"
							@update:model-value="setNsfw"
							class="setting-checkbox"
						/>
					</template>
				</app-store-body>
			</div>
		</template>
	</page-container>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import PageContainer from '@apps/market/src/components/base/PageContainer.vue';
import AppStoreBody from '@apps/market/src/components/base/AppStoreBody.vue';
import { useSettingStore } from '@apps/market/src/stores/market/setting';

const nsfw = ref(false);
const settingStore = useSettingStore();
let loading = ref(false);

onMounted(async () => {
	loading.value = true;
	nsfw.value = settingStore.nsfw;
	loading.value = false;
});

const setNsfw = async () => {
	if (loading.value) {
		return;
	}
	loading.value = true;
	await settingStore.setNsfw(nsfw.value);
	loading.value = false;
};
</script>

<style scoped lang="scss">
.setting-scroll {
	width: 100%;
	height: 100%;
	padding: 56px 44px;

	.setting-label {
		margin-top: 32px;
		font-family: Roboto;
		font-size: 16px;
		font-weight: 700;
		line-height: 24px;
		letter-spacing: 0em;
		text-align: left;
		color: $title;
	}

	.setting-checkbox {
		margin-top: 12px;
		font-family: Roboto;
		font-size: 14px;
		font-weight: 400;
		line-height: 20px;
		letter-spacing: 0em;
		text-align: left;
		color: $sub-title;
	}
}
</style>
