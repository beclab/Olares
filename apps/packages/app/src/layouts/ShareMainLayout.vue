<template>
	<q-layout view="hHh lpR fFf" style="height: 100vh">
		<div class="header row items-center q-px-md">
			<img src="share/files.svg" width="24" height="24" />
			<div class="q-ml-sm text-ink-1 text-h6">Files</div>
		</div>
		<q-page-container class="content">
			<div class="full-height full-width row items-center justify-center">
				<!-- {{ shareStore.token }} -->
				<link-expired
					v-if="shareStore.expiredInfo.status"
					:time="`${date.formatDate(
						(shareStore.expiredInfo.time || 0) * 1000,
						'YYYY-MM-DD HH:mm:ss'
					)}`"
				/>
				<login-view
					v-else-if="!shareStore.token"
					@login="doLogin"
					:message="requestError"
					@reset-error="resetError"
				/>
				<router-view v-else />
			</div>
		</q-page-container>
	</q-layout>
</template>

<script lang="ts" setup>
import { useRoute } from 'vue-router';
import LoginView from '../pages/Share/LoginView.vue';
import LinkExpired from '../pages/Share/LinkExpired.vue';
import share from '../api/files/v2/common/share';
import { onMounted, ref } from 'vue';
import { useShareStore } from 'src/stores/share/share';
import { date } from 'quasar';

const route = useRoute();

const shareStore = useShareStore();

const share_id = route.params.share_id;

const requestError = ref('');

const query = async () => {
	if (!share_id) {
		return;
	}

	shareStore.path_id = share_id as string;

	const token = await shareStore.getToken();
	if (token) {
		shareStore.token = token;
	}
	await shareStore.requestShareInfo();
};

const doLogin = async (password: string) => {
	if (!share_id) {
		return;
	}
	const result = await share.getShareToken(share_id as string, password);
	if (result && result.code == 0) {
		shareStore.setToken(result.token);
		await shareStore.requestShareInfo();
	} else if (result && !!result.message) {
		requestError.value = result.message;
	}
};

const resetError = () => {
	requestError.value = '';
};

onMounted(() => {
	query();
});
</script>

<style lang="scss" scoped>
.header {
	height: 48px;
	border-bottom: 1px solid $separator;
}

.content {
	background: linear-gradient(217.53deg, #f4fbff 12.25%, #fcfbf6 87.36%);
	height: calc(100% - 48px);
}
</style>
