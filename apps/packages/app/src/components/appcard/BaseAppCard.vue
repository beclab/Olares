<template>
	<div
		class="full-width"
		:style="{
			borderBottom:
				isLastLine || deviceStore.isMobile
					? 'none'
					: `1px solid ${separatorColor}`,
			'--iconSize': deviceStore.isMobile ? '56px' : '64px',
			'--paddingLeft': '12px',
			'--paddingTop': deviceStore.isMobile ? '0' : '20px',
			'--paddingBottom': deviceStore.isMobile ? '0' : '20px',
			'--cardHeight': deviceStore.isMobile ? '80px' : '113px'
		}"
	>
		<div
			v-if="appAggregation"
			class="application_single_normal row justify-between items-center cursor-pointer"
			@click="goAppDetails"
			@mouseover="hoverRef = true"
			@mouseleave="hoverRef = false"
		>
			<app-icon
				v-if="showIcon(type)"
				:src="appIcon"
				:size="deviceStore.isMobile ? 56 : 64"
				:cs-size="20"
				:cs-app="clusterScopedApp"
			/>

			<app-featured-image v-else :width="128" :src="appFeaturedImage" />

			<div
				v-if="!deviceStore.isMobile"
				class="application_right column justify-center items-start"
			>
				<div class="application_title_layout row justify-start items-baseline">
					<div class="application_title text-h6 text-ink-1">
						{{ appTitle }}
					</div>
				</div>

				<div
					class="text-ink-3 text-body3"
					:class="
						globalConfig.isOfficial
							? 'application_content_public'
							: 'application_content'
					"
				>
					{{ appDesc }}
				</div>
				<span
					v-if="!globalConfig.isOfficial"
					class="application_btn row justify-start items-center"
				>
					<install-button
						v-if="appAggregation"
						:item="appAggregation.app_status_latest"
						:app-name="appName"
						:version="appVersion"
						:source-id="sourceId"
						:larger="deviceStore.isMobile"
						:is-update="isUpdate"
						:manager="manager"
						:layout="deviceStore.isMobile ? 'column' : 'row'"
						@on-error="onHandleErrorGroup"
					/>
				</span>
			</div>

			<div v-else class="app-card-right row justify-between items-center">
				<div
					class="app-card-text column justify-center"
					:style="
						globalConfig.isOfficial
							? 'width: calc(100% - 12px);'
							: 'width: calc(100% - 108px - 12px);'
					"
				>
					<div class="app-card-title text-subtitle2 text-ink-1">
						{{ appTitle }}
					</div>
					<div class="app-card-content text-overline text-ink-3">
						{{ appDesc }}
					</div>
				</div>
				<install-button
					v-if="appAggregation"
					:item="appAggregation.app_status_latest"
					:layout="deviceStore.isMobile ? 'column' : 'row'"
					:larger="deviceStore.isMobile"
					:version="appVersion"
					:app-name="appName"
					:source-id="sourceId"
				/>
			</div>
		</div>
		<div
			v-else
			class="application_single_normal row justify-between items-center"
		>
			<app-icon
				v-if="showIcon(type)"
				:skeleton="true"
				:size="deviceStore.isMobile ? 56 : 64"
				:cs-size="20"
			/>

			<app-featured-image v-else :width="128" :skeleton="true" />

			<div class="application_right column justify-center items-start">
				<div class="application_title_layout row justify-start items-center">
					<q-skeleton width="60px" height="22px" />
				</div>

				<q-skeleton
					class="application_content q-mt-xs"
					width="140px"
					height="16px"
				/>
				<q-skeleton
					v-if="!globalConfig.isOfficial"
					class="application_btn"
					width="72px"
					height="22px"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import AppFeaturedImage from '../../components/appcard/AppFeaturedImage.vue';
import InstallButton from '../../components/appcard/InstallButton.vue';
import AppIcon from '../../components/appcard/AppIcon.vue';
import { useDeviceStore } from '../../stores/settings/device';
import { CFG_TYPE, showIcon } from '../../constant/config';
import globalConfig from '../../api/market/config';
import useAppCard from './useAppCard';
import { ref } from 'vue';

const props = defineProps({
	appName: {
		type: String,
		required: false
	},
	sourceId: {
		type: String,
		required: true
	},
	type: {
		type: String,
		default: CFG_TYPE.APPLICATION
	},
	version: {
		type: Boolean,
		default: false
	},
	isLastLine: {
		type: Boolean,
		default: false
	},
	disabled: {
		type: Boolean,
		default: false
	},
	isUpdate: {
		type: Boolean,
		required: false,
		default: false
	},
	manager: {
		type: Boolean,
		required: false,
		default: false
	}
});

const emit = defineEmits(['onError']);
const deviceStore = useDeviceStore();

const hoverRef = ref(false);

const onHandleErrorGroup = (value) => {
	emit('onError', value);
};

const {
	appAggregation,
	clusterScopedApp,
	appIcon,
	appTitle,
	appDesc,
	appVersion,
	appFeaturedImage,
	goAppDetails,
	separatorColor
} = useAppCard(props);
</script>
<style lang="scss" scoped>
.application_single_normal {
	width: 100%;
	height: var(--cardHeight);
	overflow: hidden;

	.application_right {
		width: calc(100% - var(--iconSize));
		height: 100%;
		padding-top: var(--paddingTop);
		padding-bottom: var(--paddingBottom);
		padding-left: var(--paddingLeft);

		.application_title_layout {
			width: 100%;
			max-height: 40px;
			text-overflow: ellipsis;
			white-space: nowrap;
			overflow: hidden;

			.application_title {
				max-width: 80%;
				overflow: hidden;
				text-overflow: ellipsis;
			}
		}

		.application_content {
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 1;
			-webkit-box-orient: vertical;
		}

		.application_content_public {
			@extend .application_content;
			-webkit-line-clamp: 1;
		}

		.application_btn {
			margin-top: 8px;
		}
	}

	.app-card-right {
		width: calc(100% - var(--iconSize));
		padding-left: var(--paddingLeft);
		height: 100%;
		position: relative;

		.app-card-text {
			height: 56px;

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
