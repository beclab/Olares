<template>
	<div class="row items-center justify-evenly">
		<q-icon
			name="sym_r_chevron_left"
			size="32px"
			@click="goBack"
			color="white"
			class="back-icon"
		/>
		<span class="text-white text-h5 scan-title">{{ t('scan_qr_code') }}</span>
		<div class="img-box" />
		<canvas id="myCanvas" />
		<q-btn
			flat
			style="position: absolute; bottom: 77px; right: 50px"
			@click="toPhotoAlbum"
		>
			<div class="column justify-between items-center content-center">
				<div
					class="row justify-center items-center content-center photo-album-img-bg"
				>
					<q-icon name="sym_r_imagesmode" size="20px" />
				</div>
				<div class="text-white text-body3 album_title">
					{{ t('photo_album') }}
				</div>
			</div>
		</q-btn>
	</div>
</template>

<script lang="ts" setup>
import { onMounted, onUnmounted, watch } from 'vue';
import { useQuasar } from 'quasar';
import { useMenuStore } from '../../stores/menu';
import { useI18n } from 'vue-i18n';
import { notifyFailed } from '../../utils/notifyRedefinedUtil';
import { ScreenOrientation } from '@capacitor/screen-orientation';
import { onBeforeMount } from 'vue';
import { StatusBar } from '@capacitor/status-bar';
import AndroidPlugins from 'src/platform/interface/capacitor/android/androidPlugins';
import { getNativeAppPlatform } from 'src/application/platform';

const props = defineProps({
	start: {
		type: Boolean,
		required: false,
		default: true
	}
});

const $q = useQuasar();
const menuStore = useMenuStore();
const { t } = useI18n();

const emits = defineEmits(['cancel', 'scanResult']);

let canvas: any;
let ctx: any;

const startScan = async () => {
	try {
		const result = await getNativeAppPlatform().scanQR();
		if (result && result.length) {
			resoleScanResult(result);
		}
	} catch (error) {
		console.error(error);
	}
};

watch(
	() => props.start,
	async () => {
		if (props.start) {
			checkScanPermissionAndStart();
		} else {
			await getNativeAppPlatform().stopScanQR();
		}
	}
);

onBeforeMount(async () => {
	if ($q.platform.is.android) {
		// StatusBar.setOverlaysWebView({ overlay: true });
		await AndroidPlugins.EdgeToEdge.disable();
	}
	if (getNativeAppPlatform().isPad) {
		await ScreenOrientation.lock({
			orientation: 'portrait-primary'
		});
	}
	document.body.classList.add('qrscanner');
	menuStore.changeSafeArea(false);
	menuStore.updateHideBackground(true);
	setTimeout(() => {
		initCanvas();
		window.addEventListener('onresize', initCanvas);
	}, 300);
});

onMounted(async () => {
	setTimeout(() => {
		checkScanPermissionAndStart();
	}, 500);
});

const checkScanPermissionAndStart = () => {
	getNativeAppPlatform()
		.scanQRDidUserGrantPermission()
		.then((permission) => {
			if (permission) {
				startScan();
			} else {
				getNativeAppPlatform().scanQrCheckPermission();
			}
		});
};

onUnmounted(async () => {
	getNativeAppPlatform().stopScanQR();
	document.body.classList.remove('qrscanner');
	menuStore.changeSafeArea(true);
	menuStore.updateHideBackground(false);
	window.removeEventListener('onresize', initCanvas);
	if (getNativeAppPlatform().isPad) {
		ScreenOrientation.unlock();
	}
	if ($q.platform.is.android) {
		// StatusBar.setOverlaysWebView({ overlay: false });
		await AndroidPlugins.EdgeToEdge.enable();
	}
});

const toPhotoAlbum = async () => {
	const qrcodeContent =
		await getNativeAppPlatform().getQRCodeImageFromPhotoAlbum();
	if (!qrcodeContent || qrcodeContent.length === 0) {
		notifyFailed(t('read_fail_try_again'));
		return;
	}
	resoleScanResult(qrcodeContent);
	$q.loading.hide();
};

const goBack = () => {
	emits('cancel');
};

const resoleScanResult = async (result: string) => {
	emits('scanResult', result);
};

function initCanvas() {
	canvas = document.getElementById('myCanvas');
	ctx = canvas.getContext('2d');
	canvas.width = window.innerWidth;
	canvas.height = window.innerHeight;
	drawCanvas();
}
function drawCanvas() {
	// outside
	ctx.moveTo(0, 0);
	ctx.lineTo(window.innerWidth, 0);
	ctx.lineTo(window.innerWidth, window.innerHeight);
	ctx.lineTo(0, window.innerHeight);
	ctx.closePath();

	// inner
	const height = 290;
	const width = 290;
	const top = 200.5;
	const left = (window.innerWidth - 290) / 2;
	const cornerRadius = 16;

	// left-top
	ctx.moveTo(left + cornerRadius, top);
	ctx.arcTo(left, top, left, top + cornerRadius, cornerRadius);

	// left-bottom
	ctx.lineTo(left, top + height - cornerRadius);
	ctx.arcTo(
		left,
		top + height,
		left + cornerRadius,
		top + height,
		cornerRadius
	);

	// right-bottom
	ctx.lineTo(left + width - cornerRadius, top + height);
	ctx.arcTo(
		left + width,
		top + height,
		left + width,
		top + height - cornerRadius,
		cornerRadius
	);

	// right-top
	ctx.lineTo(left + width, top + cornerRadius);
	ctx.arcTo(left + width, top, left + width - cornerRadius, top, cornerRadius);

	ctx.closePath();

	ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
	ctx.fill();
}

defineExpose({ checkScanPermissionAndStart });
</script>

<style lang="scss" scoped>
.back-icon {
	position: absolute;
	left: 20px;
	top: calc(env(safe-area-inset-top) + 39px);
}

.scan-title {
	position: absolute;
	width: 200px;
	height: 30px;
	left: calc(50% - 60px);
	right: calc(50% - 60px);
	top: calc(env(safe-area-inset-top) + 39px);
	text-align: left;
}

.img-box {
	position: absolute;
	background: linear-gradient($yellow, $yellow) left top,
		linear-gradient($yellow, $yellow) left top,
		linear-gradient($yellow, $yellow) right top,
		linear-gradient($yellow, $yellow) right top,
		linear-gradient($yellow, $yellow) right bottom,
		linear-gradient($yellow, $yellow) right bottom,
		linear-gradient($yellow, $yellow) left bottom,
		linear-gradient($yellow, $yellow) left bottom;
	background-repeat: no-repeat;
	background-size: 2px 40px, 40px 2px;
	border-radius: 16px;
	top: 197px;
	height: 296px;
	width: 296px;
}

.scan-label {
	position: absolute;
	width: auto;
	height: 12px;
	top: 509px;
}

.photo-album-img-bg {
	width: 32px;
	height: 32px;
	background: $white;
	border-radius: 50%;
}

.album_title {
	margin-top: 10px;
	text-align: center;
	text-transform: capitalize;
}
</style>
