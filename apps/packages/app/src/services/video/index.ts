import {
	DownMixStereoAlgorithmType,
	EncoderPresetType,
	hardwareAccelerationOptions,
	HardwareAccelerationType,
	HardwareAccelerationTypeClass,
	VideoConfigData,
	VideoService as VideoServiceInterface
} from '../abstractions/video/service';

import { HardwareNoneClass } from './HadwareAcceleration/none';
import { HardwareAMFClass } from './HadwareAcceleration/amf';
import { HardwareNVENCClass } from './HadwareAcceleration/nvenc';
import { HardwareQSVClass } from './HadwareAcceleration/qsv';
import { HardwareRKMPPClass } from './HadwareAcceleration/rkmpp';
import { HardwareToolBoxClass } from './HadwareAcceleration/toobox';
import { HardwareV4l2Class } from './HadwareAcceleration/v4l2';
import { HardwareVAAPIClass } from './HadwareAcceleration/vaapi';
import { useVideoStore } from 'src/stores/settings/video';

class VideoService implements VideoServiceInterface {
	supportHardwareAccelerationTypeList = hardwareAccelerationOptions;
	hardwareAccelerationTypeClass: HardwareAccelerationTypeClass;
	initData: VideoConfigData;

	initMockData() {
		this.configInitData({
			HardwareDecodingCodecs: [],
			EnableDecodingColorDepth10Hevc: true,
			EnableDecodingColorDepth10Vp9: true,
			EnableHardwareEncoding: true,
			EnableIntelLowPowerH264HwEncoder: true,
			EnableIntelLowPowerHevcHwEncoder: true,
			EnableEnhancedNvdecDecoder: false,
			HardwareAccelerationType: 'amf',
			AllowHevcEncoding: true,
			AllowAv1Encoding: true,
			EncodingThreadCount: -1,
			TranscodingTempPath: '/cache/transcodes',
			EnableAudioVbr: true,
			DownMixAudioBoost: 2,
			DownMixStereoAlgorithm: DownMixStereoAlgorithmType.None,
			EncoderPreset: EncoderPresetType.Automatic,
			H265Crf: 33,
			H264Crf: 22,
			EnableThrottling: true,
			ThrottleDelaySeconds: 180,
			EnableSegmentDeletion: true,
			SegmentKeepSeconds: 720,
			MaxMuxingQueueSize: 2048,
			PreferSystemNativeHwDecoder: false
		});
	}

	configInitData(initData: VideoConfigData) {
		this.initData = initData;
		this.updateHardwareAccelerationType(
			(this.initData.HardwareAccelerationType as any) ||
				HardwareAccelerationType.None,
			true
		);
		this.updateEncodingScheme();
		this.updateTranscodingSettings();
		this.updateAudioTranscoding();
		this.updateConfigEncodingQuality();
		this.updateOthers();
	}

	updateHardwareAccelerationType(
		type: HardwareAccelerationType,
		isInit = false
	) {
		switch (type) {
			case HardwareAccelerationType.None:
				this.hardwareAccelerationTypeClass = new HardwareNoneClass();
				break;
			case HardwareAccelerationType.AMD_AMF:
				this.hardwareAccelerationTypeClass = new HardwareAMFClass(
					this.initData
				);
				break;
			case HardwareAccelerationType.Nvidia_NVENC:
				this.hardwareAccelerationTypeClass = new HardwareNVENCClass(
					this.initData
				);
				break;

			case HardwareAccelerationType.Intel_QuickSync:
				this.hardwareAccelerationTypeClass = new HardwareQSVClass(
					this.initData
				);
				break;
			case HardwareAccelerationType.Rockchip_MPP:
				this.hardwareAccelerationTypeClass = new HardwareRKMPPClass(
					this.initData
				);
				break;
			case HardwareAccelerationType.Apple_VideoToolBox:
				this.hardwareAccelerationTypeClass = new HardwareToolBoxClass(
					this.initData
				);
				break;
			case HardwareAccelerationType.Video4Linux2:
				this.hardwareAccelerationTypeClass = new HardwareV4l2Class(
					this.initData
				);
				break;
			case HardwareAccelerationType.Video_Acceleration_API:
				this.hardwareAccelerationTypeClass = new HardwareVAAPIClass(
					this.initData
				);
				break;
		}

		const videoStore = useVideoStore();
		videoStore.configHardwareAcceleration(this.hardwareAccelerationTypeClass);
		if (!isInit) {
			this.updateInitDataProps({
				HardwareAccelerationType: type
			});
		}
	}

	updateEncodingScheme() {
		const videoStore = useVideoStore();
		videoStore.configEncodingScheme(this.initData);
	}

	updateTranscodingSettings() {
		const videoStore = useVideoStore();
		videoStore.configTranscodingSettings(this.initData);
	}

	updateAudioTranscoding() {
		const videoStore = useVideoStore();
		videoStore.configAudioTranscoding(this.initData);
	}

	updateConfigEncodingQuality() {
		const videoStore = useVideoStore();
		videoStore.configEncodingQuality(this.initData);
	}

	updateOthers() {
		const videoStore = useVideoStore();
		videoStore.configOthers(this.initData);
	}

	updateInitDataProps(props: Partial<VideoConfigData>) {
		props && Object.assign(this.initData, props);
		const videoStore = useVideoStore();
		videoStore.updateVideoConfig(this.initData);
	}
}

export default new VideoService();
