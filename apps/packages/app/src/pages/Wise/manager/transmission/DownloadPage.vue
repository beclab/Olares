<template>
	<common-page-container :menu-type="MenuType.Download">
		<template v-slot:tab-content-1>
			<download-tab-page />
		</template>
		<template v-slot:tab-content-2>
			<download-tab-page
				:status="[
					TransferStatus.Running,
					TransferStatus.Prepare,
					TransferStatus.Pending,
					TransferStatus.Resuming,
					TransferStatus.Paused
				]"
			/>
		</template>
		<template v-slot:tab-content-3>
			<download-tab-page :status="[TransferStatus.Completed]" />
		</template>
		<template v-slot:tab-content-4>
			<download-tab-page :status="[TransferStatus.Error]" />
		</template>
	</common-page-container>
</template>

<script lang="ts" setup>
import CommonPageContainer from '../../library/content/CommonPageContainer.vue';
import DownloadTabPage from 'src/pages/Wise/manager/transmission/DownloadTabPage.vue';
import { MenuType } from 'src/utils/rss-menu';
import { TransferStatus } from 'src/utils/interface/transfer';
</script>

<style scoped lang="scss"></style>
