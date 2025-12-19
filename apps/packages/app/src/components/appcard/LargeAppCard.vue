<template>
	<div
		v-if="appAggregation"
		:style="{
			borderBottom: isLastLine ? 'none' : `1px solid ${separatorColor}`,
			'--iconSize': showIcon(type) ? '120px' : '224px',
			'--paddingLeft': showIcon(type) ? '20px' : '32px'
		}"
		class="row justify-between items-center application_single_larger"
		@click="goAppDetails"
		@mouseover="hoverRef = true"
		@mouseleave="hoverRef = false"
	>
		<app-icon
			v-if="showIcon(type)"
			:src="appIcon"
			size="120"
			cs-size="24"
			icon-radius="20"
			:cs-app="clusterScopedApp"
		/>

		<app-featured-image v-else width="224" :src="appFeaturedImage" />

		<div class="application_right column justify-center items-start">
			<div class="application_title_layout row justify-start items-baseline">
				<div class="application_title text-h5 text-ink-1">
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
					:larger="true"
					:is-update="isUpdate"
					:manager="manager"
					@on-error="onHandleErrorGroup"
				/>
			</span>

			<span
				v-if="globalConfig.isOfficial"
				class="text-body3 text-grey-5 q-mt-md"
				>{{ t('main.terminus') }}
			</span>
		</div>
	</div>
	<div
		v-else
		:style="{
			borderBottom: isLastLine ? 'none' : `1px solid ${separatorColor}`,
			'--iconSize': showIcon(type) ? '100px' : '224px',
			'--paddingLeft': showIcon(type) ? '20px' : '32px'
		}"
		class="application_single_larger row justify-between items-center"
	>
		<app-icon
			v-if="showIcon(type)"
			:skeleton="true"
			size="120"
			cs-size="24"
			icon-radius="20"
		/>

		<app-featured-image v-else :width="224" :skeleton="true" />

		<div class="application_right column justify-center items-start">
			<div class="application_title_layout row justify-start items-center">
				<q-skeleton class="text-h5" width="75px" height="26px" />
				<q-skeleton
					v-if="version"
					class="application_version text-body3"
					width="41px"
					height="20px"
				/>
			</div>

			<q-skeleton
				class="application_content q-mt-xs"
				width="500px"
				height="20px"
			/>
			<q-skeleton
				v-if="!globalConfig.isOfficial"
				class="application_btn"
				width="88px"
				height="30px"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import AppFeaturedImage from '../../components/appcard/AppFeaturedImage.vue';
import InstallButton from '../../components/appcard/InstallButton.vue';
import AppIcon from '../../components/appcard/AppIcon.vue';
import { CFG_TYPE, showIcon } from '../../constant/config';
import globalConfig from '../../api/market/config';
import useAppCard from './useAppCard';
import { useI18n } from 'vue-i18n';
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
const { t } = useI18n();

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
.application_single_larger {
	width: 100%;
	height: 140px;
	overflow: hidden;

	.application_right {
		width: calc(100% - var(--iconSize));
		height: 100%;
		padding-left: var(--paddingLeft);
		position: relative;

		.application_title_layout {
			width: 100%;
			text-overflow: ellipsis;
			white-space: nowrap;
			overflow: hidden;

			.application_title {
				max-width: 70%;
			}

			.application_version {
				max-width: 25%;
				margin-left: 8px;
			}
		}

		.application_content {
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 2;
			-webkit-box-orient: vertical;
		}

		.application_btn {
			margin-top: 12px;
		}
	}
}
</style>
