<template>
	<div
		v-if="appAggregation"
		class="recommend-app-card row justify-between items-center"
		:class="disabled ? '' : 'cursor-pointer'"
		@click="goAppDetails"
		:style="{
			borderBottom: isLastLine ? 'none' : `1px solid ${separatorColor}`
		}"
	>
		<app-icon :src="appIcon" :size="56" :cs-app="clusterScopedApp" />
		<div class="app-card-right row justify-between items-center">
			<div
				class="app-card-text column justify-center"
				:style="
					globalConfig.isOfficial
						? 'width: calc(100% - 22px);'
						: 'width: calc(100% - 108px - 22px);'
				"
			>
				<div class="app-card-title text-subtitle2 text-ink-1">
					{{ appTitle }}
				</div>
				<div class="app-card-content text-body3 text-ink-2">
					{{ appDesc }}
				</div>
			</div>
			<install-button
				v-if="appAggregation"
				:item="appAggregation.app_status_latest"
				:layout="layout"
				:larger="deviceStore.isMobile"
				:version="appVersion"
				:app-name="appName"
				:source-id="sourceId"
				:image-size="imageSize"
			/>
		</div>
	</div>
	<div
		v-else
		class="recommend-app-card row justify-between items-center"
		:style="{
			borderBottom: isLastLine ? 'none' : `1px solid ${separatorColor}`
		}"
	>
		<app-icon :skeleton="true" :size="56" />
		<div class="app-card-right row justify-between items-center">
			<div class="app-card-text column justify-center">
				<q-skeleton width="60px" height="20px" />
				<q-skeleton width="100px" height="16px" style="margin-top: 4px" />
			</div>
			<q-skeleton width="72px" height="24px" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import InstallButton from '../../components/appcard/InstallButton.vue';
import AppIcon from '../../components/appcard/AppIcon.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import globalConfig from '../../api/market/config';
import useAppCard from './useAppCard';

const props = defineProps({
	appName: {
		type: String,
		required: false
	},
	sourceId: {
		type: String,
		required: true
	},
	disabled: {
		type: Boolean,
		default: false
	},
	isLastLine: {
		type: Boolean,
		default: false
	},
	layout: {
		type: String,
		default: 'row'
	},
	imageSize: {
		type: Boolean,
		default: true
	}
});
const deviceStore = useDeviceStore();

const {
	appAggregation,
	clusterScopedApp,
	appIcon,
	appTitle,
	appDesc,
	appVersion,
	goAppDetails,
	separatorColor
} = useAppCard(props);

defineExpose({ goAppDetails });
</script>
<style lang="scss" scoped>
.recommend-app-card {
	width: 100%;
	height: 80px;
	overflow: hidden;

	.app-card-right {
		width: calc(100% - 56px);
		height: 100%;
		position: relative;

		.app-card-text {
			height: 56px;
			padding-right: 4px;
			margin-left: 8px;

			.app-card-title {
				width: 100%;
				overflow: hidden;
				text-overflow: ellipsis;
				display: -webkit-box;
				-webkit-line-clamp: 1;
				-webkit-box-orient: vertical;
			}

			.app-card-content {
				@extend .app-card-title;
				-webkit-line-clamp: 2;
			}
		}
	}
}
</style>
