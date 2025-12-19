<template>
	<div class="png-download-root row justify-between items-center">
		<div class="column">
			<span class="png-download-title">{{ t('share.download_png') }}</span>
			<span class="png-download-label">{{
				t('share.high_quality_image')
			}}</span>
		</div>
		<div class="row justify-center">
			<span class="png-download-tip">.png</span>
			<q-img
				v-if="isLoading"
				class="download-image"
				:src="getRequireImage('loading-image.svg')"
			/>
			<q-icon
				v-else
				class="bio-clickable download-icon"
				size="20px"
				name="sym_r_download"
				color="grey-5"
				@click="download"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import html2canvas from 'html2canvas';
import { useQuasar } from 'quasar';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { getRequireImage } from 'src/utils/helper';
const { t } = useI18n();
const $q = useQuasar();
const props = defineProps({
	elementId: {
		type: String,
		required: true
	},
	pngName: {
		type: String,
		required: true
	}
});

const isLoading = ref(false);

const download = () => {
	const element = document.getElementById(props.elementId);
	if (element) {
		isLoading.value = true;
		html2canvas(element, {
			useCORS: true,
			logging: true,
			backgroundColor: null
		})
			.then((canvas) => {
				const imgUrl = canvas.toDataURL('image/png');
				let eleLink = document.createElement('a');
				eleLink.href = imgUrl;
				eleLink.download = props.pngName;
				document.body.appendChild(eleLink);
				eleLink.click();
				document.body.removeChild(eleLink);
			})
			.catch((e) => {
				$q.notify(`File download failed : ${e.message}`);
			})
			.finally(() => {
				setTimeout(() => {
					isLoading.value = false;
				}, 1000);
			});
	} else {
		$q.notify('File download failed');
	}
};
</script>

<style scoped lang="scss">
.png-download-root {
	width: 100%;
	height: 56px;
	padding: 12px;

	.png-download-title {
		overflow: hidden;
		color: $ink-1;
		font-family: Roboto;
		font-size: 14px;
		font-weight: 500;
		line-height: 20px;
		letter-spacing: 0em;
		text-align: left;
	}

	.png-download-label {
		font-family: Roboto;
		font-size: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: left;
		color: $ink-3;
	}

	.png-download-tip {
		font-family: Roboto;
		font-size: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: left;
		margin-right: 12px;
		color: $ink-3;
	}

	.download-icon {
		&:hover {
			border-radius: 4px;
			background: $btn-bg-pressed;
		}
	}

	.download-image {
		width: 20px;
		height: 20px;
		animation: animate 1.2s linear infinite;
		-webkit-animation: animate 1.2s linear infinite;
	}

	@keyframes animate {
		0% {
			transform: rotate(0deg);
			-webkit-transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
			-webkit-transform: rotate(360deg);
		}
	}
	@-webkit-keyframes animate {
		0% {
			transform: rotate(0deg);
			-webkit-transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
			-webkit-transform: rotate(360deg);
		}
	}
}
</style>
