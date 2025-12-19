<template>
	<page-title-component :title="t('olares_space')" :show-back="true" />

	<div class="terminus-cloud-page column items-center">
		<QRCodeLogin @success="router.back()">
			<template v-slot:mode>
				<q-img
					class="terminus-cloud-icon"
					:src="
						$q.dark.isActive
							? 'settings/terminus_cloud_dark.png'
							: 'settings/terminus_cloud.png'
					"
				/>
			</template>
		</QRCodeLogin>
	</div>
</template>

<script lang="ts" setup>
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import QRCodeLogin from 'src/components/settings/QRCodeLogin.vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

export interface TokenData {
	userid: string;
	token: string;
	expired: number;
}

const { t } = useI18n();
const router = useRouter();
</script>

<style lang="scss" scoped>
.terminus-cloud-page {
	width: 100%;
	height: calc(100% - 56px);
	.terminus-cloud-icon {
		width: 170px;
	}
}
</style>
