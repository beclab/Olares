<template>
	<path-location style="width: 100%; height: 56px">
		<template v-slot:start>
			<q-img
				class="directory-image"
				src="settings/folder_documents.svg"
				noSpinner
			/>
			<div class="column justify-between items-start" style="margin-left: 8px">
				<div class="text-subtitle2 directory-title">{{ path }}</div>
				<div class="text-body3 directory-label">
					{{
						t('count_files', {
							count: 0
						})
					}}
				</div>
			</div>
		</template>
		<template v-slot:midden>
			<div
				class="directory-circle"
				:style="{
					'--isRunning': status === 'running' ? '#29CC5F' : '#FF4D4D'
				}"
			/>
			<div class="directory-title">{{ status }}</div>
		</template>
		<template v-slot:end>
			<div class="directory-title">{{ formattedDate(time) }}</div>
		</template>
	</path-location>
</template>

<script lang="ts" setup>
import PathLocation from './PathLocation.vue';
import { date } from 'quasar';
import { useI18n } from 'vue-i18n';

defineProps({
	path: {
		type: String,
		required: true
	},
	status: {
		type: String,
		required: true
	},
	time: {
		type: String,
		required: true
	}
});

const { t } = useI18n();

const formattedDate = (datetime: string) => {
	const originalDate = new Date(datetime);
	return date.formatDate(originalDate, 'YYYY-MM-DD HH:mm:ss');
};
</script>

<style scoped lang="scss">
.directory-image {
	width: 24px;
	height: 24px;
}

.directory-title {
	color: $ink-1;
}

.directory-label {
	color: $ink-2;
}

.directory-circle {
	width: 8px;
	height: 8px;
	border-radius: 50%;
	margin-right: 4px;
	background: var(--isRunning);
}
</style>
