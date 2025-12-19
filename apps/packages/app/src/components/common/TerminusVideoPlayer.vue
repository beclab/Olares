<template>
	<div v-bind="$attrs" class="videoPlayer">
		<video
			class="video-js vjs-theme-city"
			playsinline
			webkit-playsinline
			:id="id"
			@mousemove="handleMouseMove"
			style="width: 100%; height: 100%"
		/>
	</div>
</template>

<script setup lang="ts">
import { useReadingProgressStore } from 'src/stores/rss-reading-progress';
import { onMounted, onBeforeUnmount, ref } from 'vue';
import videojs from 'video.js';
import Player from 'video.js/dist/types/player';
import 'video.js/dist/video-js.css';
import './../../css/video/city.video.css';
// import SettingsPlugin from '../video_plugins/SettingsPlugin';
// import QualityPlugin from '../video_plugins/QualityPlugin';
import { useUserStore } from 'src/stores/user';
import {
	VideoTimeType,
	setVideoCurTime,
	getVideoCurTime,
	removeVideoCurTime
} from 'src/utils/playTime';
import throttle from 'lodash.throttle';
import HotkeyManager from 'src/directives/hotkeyManager';
import { WISE_HOTKEY } from 'src/directives/wiseHotkey';

const readingProgressStore = useReadingProgressStore();
const playbackRateList = [0.5, 1, 1.5, 2];
const playerRate = ref();
const playerVolume = ref();

const props = defineProps({
	id: { type: String, default: 'vd' },
	src: { type: String, default: '' },
	path: { type: String, default: '' },
	platform: { type: String, default: 'web' },
	//rss reading progress
	progress: { type: Boolean, default: true },
	//local history
	history: { type: Boolean, default: true },
	playedTime: { string: String, required: false }
});

const emit = defineEmits(['videoPlay', 'holdShowTitle']);

let player: Player;

const Component = videojs.getComponent('Component');
class FastReplayButton extends Component {
	constructor(player, options = {}) {
		super(player, options);
		this.on('click', this.handleClick);
	}
	createEl() {
		return videojs.dom.createEl('button', {
			title: 'Rewind 10s',
			className: 'vjs-fast-replay-button vjs-control vjs-button',
			innerHTML:
				'<span aria-hidden="true" class="vjs-icon-placeholder"></span><span class="vjs-control-text" aria-live="polite">Fast Replay</span>'
		});
	}
	handleClick() {
		seekPosition(false);
	}
}
videojs.registerComponent('FastReplayButton', FastReplayButton);

class FastForwardButton extends Component {
	constructor(player, options = {}) {
		super(player, options);
		this.on('click', this.handleClick);
	}
	createEl() {
		return videojs.dom.createEl('button', {
			title: 'Forward 10S',
			className: 'vjs-fast-forward-button vjs-control vjs-button',
			innerHTML:
				'<span aria-hidden="true" class="vjs-icon-placeholder"></span><span class="vjs-control-text" aria-live="polite">Fast Forword</span>'
		});
	}
	handleClick() {
		seekPosition(true);
	}
}
videojs.registerComponent('FastForwardButton', FastForwardButton);

function options() {
	let controllBarPlug = {};
	if (props.platform === 'mobile') {
		controllBarPlug = {
			children: []
		};
	} else {
		controllBarPlug = {
			volumePanel: true,
			children: ['playToggle', 'FastReplayButton', 'FastForwardButton']
		};
	}

	return {
		autoplay: true,
		muted: false,
		loop: false,
		controls: true,
		hotkeys: false,
		playsinline: true,
		controlBar: {
			playToggle: true,
			fullscreenToggle: props.platform !== 'mobile',
			progressControl: true,
			currentTimeDisplay: true,
			durationDisplay: true,
			autoHideTime: 2500,
			...controllBarPlug
		},
		notSupportedMessage:
			'This video cannot be played temporarily, please try again later.',
		playbackRates: playbackRateList,
		sources: [
			{
				src: props.src,
				type: 'application/vnd.apple.mpegurl'
			}
		]
	};
}

const handleMouseMove = () => {
	if (!player.paused()) {
		emit('holdShowTitle', false);
	}
};

function setPlaybackRate(fast: boolean) {
	if (!player) {
		console.log('[video] Player not initialized');
		return;
	}
	const index = playbackRateList.findIndex((item) => item == playerRate.value);
	if (fast && index > -1 && index < playbackRateList.length - 1) {
		player.playbackRate(playbackRateList[index + 1]);
		console.log('[video] PlaybackRate increase');
	} else if (!fast && index > 0) {
		player.playbackRate(playbackRateList[index - 1]);
		console.log('[video] PlaybackRate decrease');
	} else {
		console.log('[video] Already at rate limit');
	}
}

function adjustVolume(large: boolean) {
	if (!player) {
		console.log('[video] Player not initialized');
		return;
	}
	const newVolume = large ? playerVolume.value + 0.1 : playerVolume.value - 0.1;
	const volume = Math.max(0, Math.min(1, newVolume));
	player.volume(volume);
	player.muted(volume === 0);
	console.log('[video] adjustVolume has been set');
}

function seekPosition(fast: boolean) {
	if (!player) {
		console.log('[video] Player not initialized');
		return;
	}
	const newTime = fast ? player.currentTime() + 10 : player.currentTime() - 10;
	const targetTime = Math.max(0, Math.min(readingProgressStore.total, newTime));
	player.currentTime(targetTime);
	console.log('[video] seekPosition has been set');
}

function playOrPause() {
	if (!player) {
		console.log('[video] Player not initialized');
		return;
	}
	player.paused() ? player.play() : player.pause();
	console.log('[video] playOrPause has been set');
}

function setFullScreen() {
	if (!player) {
		return;
	}
	player.isFullscreen() ? player.exitFullscreen() : player.requestFullscreen();
	console.log('[video] setFullScreen has been set');
}

onMounted(() => {
	console.log('[video] onMounted');
	HotkeyManager.setScope('video');
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
			},
			[WISE_HOTKEY.MEDIA.FULL_SCREEN]: () => setFullScreen()
		},
		['video']
	);
	const userStore = useUserStore();
	try {
		(videojs as any).Vhs.xhr.onRequest((req) => {
			if (userStore.current_user?.access_token) {
				req.headers = req.headers || {};
				req.headers['X-Authorization'] = userStore.current_user.access_token;
			}
			return req;
		});
		// console.log(videojs.Vhs);

		const cur_path = props.path + '_' + (userStore.current_id || '');

		player = videojs(props.id, options(), () => {
			videojs.log('player ready!');
			if (props.history) {
				const playedTime = props.playedTime || getVideoCurTime(cur_path);

				console.log('playtime', playedTime);

				if (playedTime > 0) {
					player.currentTime(playedTime);
					player.play();
					if (props.progress) {
						readingProgressStore.updateProgress(playedTime);
					}
				}
			}

			player.on('loadstart', function () {
				// videojs.log('playing');
				emit('holdShowTitle', false);
			});

			player.on('loadedmetadata', () => {
				videojs.log('player loadedmetadata!');
				const totalDuration = player.duration();
				if (props.progress) {
					readingProgressStore.setTotalProgress(totalDuration);
				}
			});

			player.on('play', function () {
				videojs.log('player start');
				// emit('holdShowTitle', false);
			});

			player.on('volumechange', function () {
				playerVolume.value = player.volume();
				console.log('[radio] volumechange update' + playerVolume.value);
			});

			player.on('ratechange', function () {
				playerRate.value = player.playbackRate();
				console.log('[radio] ratechange update' + playerRate.value);
			});

			player.on('playing', function () {
				// videojs.log('playing');
				// emit('holdShowTitle', false);
			});

			player.on('pause', function () {
				emit('holdShowTitle', true);
			});

			player.on('touchstart', function () {
				emit('holdShowTitle', false);
			});

			player.on('ended', function () {
				videojs.log('ended');
				if (props.history) {
					removeVideoCurTime(cur_path);
				}
			});

			player.on(
				'timeupdate',
				throttle(function () {
					if (player.ended()) {
						return false;
					}
					// console.log('player', player);
					const currentTime = player.currentTime();

					const parmas: VideoTimeType = {
						path: cur_path,
						time: currentTime || 0
					};

					if (props.history) {
						setVideoCurTime(parmas);
					}
					if (props.progress) {
						readingProgressStore.updateProgress(currentTime);
					}
				}, 1000)
			);

			player.on('error', (error: string) => {
				emit('holdShowTitle', true);
			});
		});
		playerRate.value = player.playbackRate();
		playerVolume.value = player.volume();
		// videojs.registerPlugin('settingsPlugin', SettingsPlugin);
		// videojs.registerPlugin('qualityPlugin', QualityPlugin);

		// (player as any).settingsPlugin();
		// (player as any).qualityPlugin();
	} catch (error) {
		console.log('catch', error);
	}
});

onBeforeUnmount(() => {
	if (player) {
		player.dispose();
	}
	HotkeyManager.deleteScope('video');
	console.log('[video] onBeforeUnmount');
	console.log('===========>', 'onBeforeUnmount' + props.src);
});
</script>

<style lang="scss">
.videoPlayer {
	width: 100%;
	height: 100%;
	position: relative;
	overflow: hidden;
}
</style>
