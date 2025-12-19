<template>
	<common-page-container :menu-type="MenuType.Transmission">
		<template v-slot:tab-content-1>
			<download-page />
		</template>
		<template v-slot:tab-content-2>
			<upload-page />
		</template>
	</common-page-container>
</template>

<script lang="ts" setup>
import CommonPageContainer from '../../library/content/CommonPageContainer.vue';
import { MenuType } from '../../../../utils/rss-menu';
import DownloadPage from './DownloadPage.vue';
import UploadPage from './UploadPage.vue';
</script>

<style scoped lang="scss"></style>
