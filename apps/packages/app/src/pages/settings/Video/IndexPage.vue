<template>
	<page-title-component :show-back="false" :title="t('video')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<template v-for="(page, index) in pages" :key="index">
				<bt-form-item
					:title="page.name"
					@click="enterPage(page.path)"
					:chevronRight="true"
					:width-separator="index + 1 < pages.length"
				/>
			</template>
		</bt-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { onMounted } from 'vue';
import { useVideoStore } from 'src/stores/settings/video';
import VideoService from 'src/services/video';
import BtList from 'src/components/settings/base/BtList.vue';

const { t } = useI18n();
const videoStore = useVideoStore();
const router = useRouter();

const pages = ref([
	{
		name: t('Hardware Acceleration'),
		path: '/video/hardwareAcceleration'
	},
	{
		name: t('Encoding Scheme'),
		path: '/video/encodingScheme'
	},
	{
		name: t('Transcoding Settings'),
		path: '/video/transcodingSettings'
	},
	{
		name: t('Audio Transcoding'),
		path: '/video/audioTranscoding'
	},
	{
		name: t('Encoding Quality'),
		path: '/video/encodingQuality'
	},
	{
		name: t('Others'),
		path: '/video/others'
	}
]);
const enterPage = (path: string) => {
	router.push(path);
};

onMounted(async () => {
	try {
		if (!VideoService.initData) {
			let config = await videoStore.getVideoConfig();
			if (typeof config == 'string') {
				config = JSON.parse(config);
			}
			console.log('config ===>', config);

			VideoService.configInitData(config);
			console.log(config);
		}
	} catch (error) {
		console.log(error);
	}
});
</script>

<style scoped lang="scss"></style>
