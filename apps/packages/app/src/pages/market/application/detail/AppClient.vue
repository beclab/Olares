<template>
	<app-intro-card :title="t('detail.get_a_client')" class="q-mt-lg">
		<template
			v-slot:content
			v-if="
				appEntry.supportClient &&
				(appEntry.supportClient.ios ||
					appEntry.supportClient.android ||
					appEntry.supportClient.edge ||
					appEntry.supportClient.windows ||
					appEntry.supportClient.mac ||
					appEntry.supportClient.chrome ||
					appEntry.supportClient.linux)
			"
		>
			<div class="row">
				<support-client
					:value="appEntry.supportClient.android"
					:type="CLIENT_TYPE.android"
				/>
				<support-client
					:value="appEntry.supportClient.ios"
					:type="CLIENT_TYPE.ios"
					class="q-mt-lg"
				/>
				<support-client
					:value="appEntry.supportClient.mac"
					:type="CLIENT_TYPE.mac"
					class="q-mt-lg"
				/>
				<support-client
					:value="appEntry.supportClient.windows"
					:type="CLIENT_TYPE.windows"
					class="q-mt-lg"
				/>
				<support-client
					:value="appEntry.supportClient.chrome"
					:type="CLIENT_TYPE.chrome"
					class="q-mt-lg"
				/>
				<support-client
					:value="appEntry.supportClient.edge"
					:type="CLIENT_TYPE.edge"
					class="q-mt-lg"
				/>
				<support-client
					:value="appEntry.supportClient.linux"
					:type="CLIENT_TYPE.linux"
					class="q-mt-lg"
				/>
			</div>
		</template>
	</app-intro-card>
</template>

<script setup lang="ts">
import SupportClient from '../../../../components/appintro/SupportClient.vue';
import AppIntroCard from '../../../../components/appintro/AppIntroCard.vue';
import { CLIENT_TYPE } from '../../../../constant/constants';
import { PropType } from 'vue';
import { useI18n } from 'vue-i18n';

defineProps({
	appEntry: {
		type: Object as PropType<any>,
		require: true
	},
	appName: {
		type: String,
		require: true
	},
	sourceId: {
		type: String,
		require: true
	}
});

const { t } = useI18n();
</script>

<style scoped lang="scss"></style>
