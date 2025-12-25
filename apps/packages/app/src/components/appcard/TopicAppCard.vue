<template>
	<adaptive-layout>
		<template v-slot:pc>
			<div
				v-if="appAggregation"
				class="topic-app-card column justify-center items-center"
				:class="disabled ? '' : 'cursor-pointer'"
				@click="goAppDetails"
				:style="{ '--cardHeight': ' 268px' }"
			>
				<app-icon :src="appIcon" :size="100" :cs-app="clusterScopedApp" />
				<div class="topic-app-card-title text-h5 text-ink-1 q-mt-md">
					{{ appTitle }}
				</div>
				<div class="topic-app-card-content text-body2 text-ink-3">
					{{ appDesc }}
				</div>
				<install-button
					class="q-mt-md"
					v-if="appAggregation"
					:item="appAggregation.app_status_latest"
					:layout="layout"
					:version="appVersion"
					:app-name="appName"
					:source-id="sourceId"
				/>
			</div>
			<div v-else class="topic-app-card column justify-center items-center">
				<app-icon :skeleton="true" :size="100" />
				<q-skeleton width="60px" height="20px" class="q-mt-md" />
				<q-skeleton width="160px" height="16px" class="q-mt-xs" />
				<q-skeleton width="88px" height="24px" class="q-mt-md" />
			</div>
		</template>

		<template v-slot:mobile>
			<div
				v-if="appAggregation"
				class="topic-app-card column justify-center items-center"
				:class="disabled ? '' : 'cursor-pointer'"
				@click="goAppDetails"
				:style="{ '--cardHeight': '196px' }"
			>
				<app-icon :src="appIcon" :size="56" :cs-app="clusterScopedApp" />
				<div class="topic-app-card-title text-subtitle2 text-ink-1 q-mt-md">
					{{ appTitle }}
				</div>
				<div class="topic-app-card-content text-overline text-ink-3">
					{{ appDesc }}
				</div>
				<install-button
					class="q-mt-md"
					v-if="appAggregation"
					:item="appAggregation.app_status_latest"
					:layout="layout"
					:version="appVersion"
					:app-name="appName"
					:source-id="sourceId"
				/>
			</div>
			<div v-else class="topic-app-card column justify-center items-center">
				<app-icon :skeleton="true" :size="56" />
				<q-skeleton width="60px" height="20px" class="q-mt-md" />
				<q-skeleton width="160px" height="16px" class="q-mt-xs" />
				<q-skeleton width="88px" height="24px" class="q-mt-md" />
			</div>
		</template>
	</adaptive-layout>
</template>

<script lang="ts" setup>
import InstallButton from '../../components/appcard/InstallButton.vue';
import AppIcon from '../../components/appcard/AppIcon.vue';
import useAppCard from './useAppCard';
import AdaptiveLayout from '../settings/AdaptiveLayout.vue';

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
	layout: {
		type: String,
		default: 'row'
	}
});

const {
	appAggregation,
	clusterScopedApp,
	appIcon,
	appTitle,
	appDesc,
	appVersion,
	goAppDetails
} = useAppCard(props);

defineExpose({ goAppDetails });
</script>
<style lang="scss" scoped>
.topic-app-card {
	width: 100%;
	height: var(--cardHeight);
	overflow: hidden;

	.topic-app-card-right {
		width: calc(100% - 56px);
		height: 100%;
		position: relative;

		.topic-app-card-text {
			height: 56px;
			padding-right: 4px;
			margin-left: 8px;

			.topic-app-card-title {
				width: 100%;
				overflow: hidden;
				text-overflow: ellipsis;
				display: -webkit-box;
				-webkit-line-clamp: 1;
				-webkit-box-orient: vertical;
			}

			.topic-app-card-content {
				@extend .topic-app-card-title;
				-webkit-line-clamp: 2;
			}
		}
	}
}
</style>
