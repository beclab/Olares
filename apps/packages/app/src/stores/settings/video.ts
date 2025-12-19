import { defineStore } from 'pinia';
import {
	DownMixStereoAlgorithmType,
	EncoderPresetType,
	HardwareAccelerationType,
	HardwareAccelerationTypeClass,
	HardwareDecodingCodeTypeInfo,
	VideoConfigData
} from 'src/services/abstractions/video/service';
import { i18n } from 'src/boot/i18n';
import { useTokenStore } from './token';
import axios from 'axios';

export const useVideoStore = defineStore('video', {
	state: () => ({
		hardwareAcceleration: {
			support_enableDecodingColorDepth10Hevc: false,
			support_enableDecodingColorDepth10Vp9: false,

			enableDecodingColorDepth10Hevc: {
				name: i18n.global.t('HEVC 10bit'),
				value: false
			},
			enableDecodingColorDepth10Vp9: {
				name: i18n.global.t('VP9 10bit'),
				value: false
			},

			support_enableHardwareEncoding: false,
			support_enableIntelLowPowerH264HwEncoder: false,
			support_enableIntelLowPowerHevcHwEncoder: false,
			support_enableEnhancedNvdecDecoder: false,
			support_preferSystemNativeHwDecoder: false,

			enableHardwareEncoding: {
				name: i18n.global.t('Enable hardware encoding'),
				value: false
			},
			enableIntelLowPowerH264HwEncoder: {
				name: i18n.global.t('Enable low-power Intel H.264 hardware encoder'),
				value: false
			},

			enableIntelLowPowerHevcHwEncoder: {
				name: i18n.global.t('Enable low-power Intel HEVC hardware encoder'),
				value: false
			},
			decodingCodeList: [] as HardwareDecodingCodeTypeInfo[],

			type: HardwareAccelerationType.None,

			enableEnhancedNvdecDecoder: {
				name: i18n.global.t('Enable Enhanced NVDEC Decoder'),
				value: false
			},
			preferSystemNativeHwDecoder: {
				name: i18n.global.t(
					"Prefer the system's native DXVA or VA-API hardware decoder"
				),
				value: false
			}
		},
		encodingeScheme: {
			allowHevcEncoding: {
				name: i18n.global.t('Allow encoding in HEVC format'),
				value: false
			},
			allowAv1Encoding: {
				name: i18n.global.t('Allow encoding in AV1 format'),
				value: false
			}
		},
		transcodingSettings: {
			encodingThreadCount: {
				name: i18n.global.t('Transcoding thread count'),
				value: -1
			},
			cranscodingTempPath: {
				name: i18n.global.t('Transcode path'),
				value: ''
			}
		},
		audioTranscoding: {
			enableAudioVbr: {
				name: i18n.global.t('Enable VBR audio encoding'),
				value: false
			},
			downMixAudioBoost: {
				name: i18n.global.t('Audio boost when downmixing'),
				value: 0
			},
			downMixStereoAlgorithm: {
				name: i18n.global.t('Stereo Downmix Algorithm'),
				value: DownMixStereoAlgorithmType.None
			}
		},
		encodingQuality: {
			encoderPreset: {
				name: i18n.global.t('Encoder Preset'),
				value: EncoderPresetType.Automatic
			},
			h265Crf: {
				name: i18n.global.t('H.265 encoding CRF'),
				value: 0
			},
			h264Crf: {
				name: i18n.global.t('H.264 encoding CRF'),
				value: 0
			}
		},
		others: {
			enableThrottling: {
				name: i18n.global.t('Throttle Transcodes'),
				value: true
			},
			enableSegmentDeletion: {
				name: i18n.global.t('Delete segments'),
				value: true
			},
			throttleDelaySeconds: {
				name: i18n.global.t('Throttle after'),
				value: 0
			},
			segmentKeepSeconds: {
				name: i18n.global.t('Time to keep segments'),
				value: 0
			},
			maxMuxingQueueSize: {
				name: i18n.global.t('Max muxing queue size'),
				value: 0
			}
		}
	}),

	getters: {},

	actions: {
		configHardwareAcceleration(hardwareClass: HardwareAccelerationTypeClass) {
			this.hardwareAcceleration.type = hardwareClass.type;
			this.hardwareAcceleration.decodingCodeList =
				hardwareClass.decodingCodeList;

			this.hardwareAcceleration.support_enableDecodingColorDepth10Hevc =
				hardwareClass.support_enableDecodingColorDepth10Hevc;
			this.hardwareAcceleration.support_enableDecodingColorDepth10Vp9 =
				hardwareClass.support_enableDecodingColorDepth10Vp9;

			this.hardwareAcceleration.enableDecodingColorDepth10Hevc.value =
				hardwareClass.enableDecodingColorDepth10Hevc;
			this.hardwareAcceleration.enableDecodingColorDepth10Vp9.value =
				hardwareClass.enableDecodingColorDepth10Vp9;

			this.hardwareAcceleration.support_enableHardwareEncoding =
				hardwareClass.support_enableHardwareEncoding;
			this.hardwareAcceleration.support_enableIntelLowPowerH264HwEncoder =
				hardwareClass.support_enableIntelLowPowerH264HwEncoder;
			this.hardwareAcceleration.support_enableIntelLowPowerHevcHwEncoder =
				hardwareClass.support_enableIntelLowPowerHevcHwEncoder;

			this.hardwareAcceleration.enableHardwareEncoding.value =
				hardwareClass.enableHardwareEncoding;
			this.hardwareAcceleration.enableIntelLowPowerH264HwEncoder.value =
				hardwareClass.enableIntelLowPowerH264HwEncoder;
			this.hardwareAcceleration.enableIntelLowPowerHevcHwEncoder.value =
				hardwareClass.enableIntelLowPowerHevcHwEncoder;

			this.hardwareAcceleration.support_enableEnhancedNvdecDecoder =
				hardwareClass.support_enableEnhancedNvdecDecoder;
			this.hardwareAcceleration.enableEnhancedNvdecDecoder.value =
				hardwareClass.enableEnhancedNvdecDecoder;

			this.hardwareAcceleration.support_preferSystemNativeHwDecoder =
				hardwareClass.support_preferSystemNativeHwDecoder;

			this.hardwareAcceleration.preferSystemNativeHwDecoder.value =
				hardwareClass.preferSystemNativeHwDecoder;
		},
		configEncodingScheme(encodingScheme: {
			AllowHevcEncoding: boolean;
			AllowAv1Encoding: boolean;
		}) {
			this.encodingeScheme.allowHevcEncoding.value =
				encodingScheme.AllowHevcEncoding;
			this.encodingeScheme.allowAv1Encoding.value =
				encodingScheme.AllowAv1Encoding;
		},

		configTranscodingSettings(transcodingSettings: {
			EncodingThreadCount: number;
			TranscodingTempPath: string;
		}) {
			this.transcodingSettings.encodingThreadCount.value =
				transcodingSettings.EncodingThreadCount;
			this.transcodingSettings.cranscodingTempPath.value =
				transcodingSettings.TranscodingTempPath;
		},

		configAudioTranscoding(audioTranscoding: {
			EnableAudioVbr: boolean;
			DownMixAudioBoost: number;
			DownMixStereoAlgorithm: DownMixStereoAlgorithmType;
		}) {
			this.audioTranscoding.enableAudioVbr.value =
				audioTranscoding.EnableAudioVbr;
			this.audioTranscoding.downMixAudioBoost.value =
				audioTranscoding.DownMixAudioBoost;
			this.audioTranscoding.downMixStereoAlgorithm.value =
				audioTranscoding.DownMixStereoAlgorithm;
		},
		configEncodingQuality(encodingQuality: {
			EncoderPreset: EncoderPresetType;
			H265Crf: number;
			H264Crf: number;
		}) {
			if (encodingQuality.EncoderPreset) {
				this.encodingQuality.encoderPreset.value =
					encodingQuality.EncoderPreset;
			}

			this.encodingQuality.h265Crf.value = encodingQuality.H265Crf;
			this.encodingQuality.h264Crf.value = encodingQuality.H264Crf;
		},

		configOthers(others: {
			EnableThrottling: boolean;
			EnableSegmentDeletion: boolean;
			ThrottleDelaySeconds: number;
			SegmentKeepSeconds: number;
			MaxMuxingQueueSize: number;
		}) {
			this.others.enableThrottling.value = others.EnableThrottling;
			this.others.enableSegmentDeletion.value = others.EnableSegmentDeletion;
			this.others.throttleDelaySeconds.value = others.ThrottleDelaySeconds;
			this.others.segmentKeepSeconds.value = others.SegmentKeepSeconds;
			this.others.maxMuxingQueueSize.value = others.MaxMuxingQueueSize;
		},
		async getVideoConfig() {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/files/video/config`
			);
			return data;
		},
		async updateVideoConfig(data: VideoConfigData) {
			const tokenStore = useTokenStore();
			await axios.post(`${tokenStore.url}/api/files/video/config`, {
				data
			});
		}
	}
});
