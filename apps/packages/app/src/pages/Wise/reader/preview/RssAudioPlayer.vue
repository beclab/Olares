<template>
	<audio
		class="full-width"
		:src="path"
		ref="audioPlayer"
		controls
		preload="metadata"
		@ratechange="onRateChange"
		@volumechange="onVolumeChange"
		@timeupdate="onTimeUpdate"
		@loadedmetadata="onLoadedMetadata"
	/>
	<full-content-reader v-if="!src" :margin-top="true" />
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useTransferStore } from '../../../../stores/rss-transfer';
import { useReadingProgressStore } from '../../../../stores/rss-reading-progress';
import FullContentReader from './FullContentReader.vue';
import { useReaderStore } from '../../../../stores/rss-reader';
import HotkeyManager from '../../../../directives/hotkeyManager';
import { WISE_HOTKEY } from '../../../../directives/wiseHotkey';
const transferStore = useTransferStore();
const readerStore = useReaderStore();
const audioPlayer = ref();

const readingProgressStore = useReadingProgressStore();
const playbackRateList = [0.5, 1, 1.5, 2];
const playerRate = ref();
const playerVolume = ref();

const props = defineProps({
	src: {
		type: String,
		require: false
	}
});

const path = computed(() => {
	return transferStore.getDownloadUrl(props.src ? props.src : '');
});

const onLoadedMetadata = (event) => {
	const totalTime = event.target.duration;
	console.log('[radio] onLoadedMetadata totalTime' + totalTime);
	readingProgressStore.setTotalProgress(totalTime);

	if (!props.src && readerStore.readingEntry.progress) {
		const time = (readerStore.readingEntry.progress * totalTime) / 100;
		if (audioPlayer.value && time >= 0 && time <= audioPlayer.value.duration) {
			audioPlayer.value.currentTime = time;
			audioPlayer.value.play();
		}
	}
};

const onTimeUpdate = (event) => {
	const currentTime = event.target.currentTime;
	console.log('[radio] onTimeUpdate update' + currentTime);
	readingProgressStore.updateProgress(currentTime);
};

const onRateChange = (event) => {
	playerRate.value = event.target.playbackRate;
	console.log('[radio] onRateChange update' + playerRate.value);
};

const onVolumeChange = (event) => {
	playerVolume.value = event.target.volume;
	console.log('[radio] onVolumeChange update' + playerVolume.value);
};

function setPlaybackRate(increase: boolean) {
	if (!audioPlayer.value) {
		console.log('[radio] Player not initialized');
		return;
	}
	const index = playbackRateList.findIndex((item) => item == playerRate.value);
	if (increase && index > -1 && index < playbackRateList.length - 1) {
		audioPlayer.value.playbackRate = playbackRateList[index + 1];
		console.log('[radio] PlaybackRate increase');
	} else if (!increase && index > 0) {
		audioPlayer.value.playbackRate = playbackRateList[index - 1];
		console.log('[radio] PlaybackRate decrease');
	} else {
		console.log('[radio] Already at rate limit');
	}
}

function adjustVolume(increase: boolean) {
	if (!audioPlayer.value) {
		console.log('[radio] Player not initialized');
		return;
	}
	const newVolume = increase
		? playerVolume.value + 0.1
		: playerVolume.value - 0.1;
	const volume = Math.max(0, Math.min(1, newVolume));
	audioPlayer.value.volume = volume;
	audioPlayer.value.muted = volume === 0;
	console.log('[radio] adjustVolume has been set');
}

function seekPosition(forward: boolean) {
	if (!audioPlayer.value) {
		console.log('[radio] Player not initialized');
		return;
	}
	const newTime = forward
		? audioPlayer.value.currentTime + 10
		: audioPlayer.value.currentTime - 10;
	const targetTime = Math.max(0, Math.min(readingProgressStore.total, newTime));
	audioPlayer.value.currentTime = targetTime;
	console.log('[radio] seekPosition has been set');
}

function playOrPause() {
	if (!audioPlayer.value) {
		console.log('[radio] Player not initialized');
		return;
	}
	audioPlayer.value.paused || audioPlayer.value.ended
		? audioPlayer.value.play()
		: audioPlayer.value.pause();
	console.log('[radio] playOrPause has been set');
}

onMounted(() => {
	HotkeyManager.setScope('audio');
	HotkeyManager.registerHotkeys(
		{
			[WISE_HOTKEY.MEDIA.PLAY]: () => playOrPause(),
			[WISE_HOTKEY.MEDIA.FAST_SEEK]: () => seekPosition(true),
			[WISE_HOTKEY.MEDIA.BACK_SEEK]: () => seekPosition(false),
			[WISE_HOTKEY.MEDIA.VOLUME_INCREASE]: () => {
				adjustVolume(true);
			},
			[WISE_HOTKEY.MEDIA.VOLUME_DECREASE]: () => {
				adjustVolume(false);
			},
			[WISE_HOTKEY.MEDIA.RATE_INCREASE]: () => {
				setPlaybackRate(true);
			},
			[WISE_HOTKEY.MEDIA.RATE_DECREASE]: () => {
				setPlaybackRate(false);
			}
		},
		['audio']
	);
});

onBeforeUnmount(() => {
	HotkeyManager.deleteScope('audio');
});
</script>

<style lang="scss" scoped></style>
