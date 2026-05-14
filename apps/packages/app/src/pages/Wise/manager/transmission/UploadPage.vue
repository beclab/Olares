<template>
	<common-page-container :menu-type="MenuType.Upload">
		<template v-slot:tab-content-1>
			<upload-tab-page />
		</template>
		<template v-slot:tab-content-2>
			<upload-tab-page
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
			<upload-tab-page :status="[TransferStatus.Completed]" />
		</template>
		<template v-slot:tab-content-4>
			<upload-tab-page :status="[TransferStatus.Error]" />
		</template>
	</common-page-container>
</template>

<script lang="ts" setup>
import CommonPageContainer from '../../library/content/CommonPageContainer.vue';
import UploadTabPage from 'src/pages/Wise/manager/transmission/UploadTabPage.vue';
import { MenuType } from 'src/utils/rss-menu';
import { TransferStatus } from 'src/utils/interface/transfer';
</script>

<style scoped lang="scss"></style>
