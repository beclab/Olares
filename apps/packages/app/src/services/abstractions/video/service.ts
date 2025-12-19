import { i18n } from '../../../boot/i18n';

export enum HardwareAccelerationType {
	None = 'none',
	AMD_AMF = 'amf',
	Nvidia_NVENC = 'nvenc',
	Intel_QuickSync = 'qsv',
	Video_Acceleration_API = 'vaapi',
	Rockchip_MPP = 'rkmpp',
	Apple_VideoToolBox = 'videotoolbox',
	Video4Linux2 = 'v4l2m2m'
}

export const hardwareAccelerationOptions = [
	{
		label: i18n.global.t('None'),
		value: HardwareAccelerationType.None,
		enable: true
	},
	{
		label: i18n.global.t('AMD AMF'),
		value: HardwareAccelerationType.AMD_AMF,
		enable: true
	},
	{
		label: i18n.global.t('Nvidia NVENC'),
		value: HardwareAccelerationType.Nvidia_NVENC,
		enable: true
	},
	{
		label: i18n.global.t('Intel QuickSync (QSV)'),
		value: HardwareAccelerationType.Intel_QuickSync,
		enable: true
	},
	{
		label: i18n.global.t('Video Acceleration API (VAAPI)'),
		value: HardwareAccelerationType.Video_Acceleration_API,
		enable: true
	},
	{
		label: i18n.global.t('Rockchip MPP (RKMPP)'),
		value: HardwareAccelerationType.Rockchip_MPP,
		enable: true
	},
	{
		label: i18n.global.t('Apple VideoToolBox'),
		value: HardwareAccelerationType.Apple_VideoToolBox,
		enable: true
	},
	{
		label: i18n.global.t('Video4Linux2 (V4L2)'),
		value: HardwareAccelerationType.Video4Linux2,
		enable: true
	}
];

export enum HardwareDecodingCodeType {
	H_264 = 'h264',
	HEVC = 'hevc',
	MPEG1 = 'mpeg1video',
	MPEG2 = 'mpeg2video',
	MPEG4 = 'mpeg4',
	VC_1 = 'vc1',
	VP_9 = 'vp9',
	VP_8 = 'vp8',
	AV_1 = 'av1'
}

export const HardwareDecodingCodeTypeRecord: Record<
	HardwareDecodingCodeType,
	string
> = {
	[HardwareDecodingCodeType.H_264]: 'H264',
	[HardwareDecodingCodeType.HEVC]: 'HEVC',
	[HardwareDecodingCodeType.MPEG1]: 'MPEG1',
	[HardwareDecodingCodeType.MPEG2]: 'MPEG2',
	[HardwareDecodingCodeType.MPEG4]: 'MPEG4',
	[HardwareDecodingCodeType.VC_1]: 'VC1',
	[HardwareDecodingCodeType.VP_9]: 'VP9',
	[HardwareDecodingCodeType.VP_8]: 'VP8',
	[HardwareDecodingCodeType.AV_1]: 'AV1'
};

export enum DownMixStereoAlgorithmType {
	None = 'None',
	Dave750 = 'Dave750',
	NightmodeDialogue = 'NightmodeDialogue'
}

export const downMixStereoAlgorithmTypeOptions = [
	{
		label: DownMixStereoAlgorithmType.None,
		value: DownMixStereoAlgorithmType.None,
		enable: true
	},
	{
		label: DownMixStereoAlgorithmType.Dave750,
		value: DownMixStereoAlgorithmType.Dave750,
		enable: true
	},
	{
		label: DownMixStereoAlgorithmType.NightmodeDialogue,
		value: DownMixStereoAlgorithmType.NightmodeDialogue,
		enable: true
	}
];

export enum EncoderPresetType {
	Automatic = 'auto',
	veryslow = 'veryslow',
	slower = 'slower',
	slow = 'slow',
	medium = 'medium',
	fast = 'fast',
	faster = 'faster',
	veryfast = 'veryfast',
	superfast = 'superfast',
	ultrafast = 'ultrafast'
}

export const encoderPresetTypeOptions = [
	{
		label: i18n.global.t('Automatic'),
		value: EncoderPresetType.Automatic,
		enable: true
	},
	{
		label: EncoderPresetType.veryslow,
		value: EncoderPresetType.veryslow,
		enable: true
	},
	{
		label: EncoderPresetType.slower,
		value: EncoderPresetType.slower,
		enable: true
	},
	{
		label: EncoderPresetType.slow,
		value: EncoderPresetType.slow,
		enable: true
	},
	{
		label: EncoderPresetType.medium,
		value: EncoderPresetType.medium,
		enable: true
	},
	{
		label: EncoderPresetType.fast,
		value: EncoderPresetType.fast,
		enable: true
	},
	{
		label: EncoderPresetType.faster,
		value: EncoderPresetType.faster,
		enable: true
	},
	{
		label: EncoderPresetType.veryfast,
		value: EncoderPresetType.veryfast,
		enable: true
	},
	{
		label: EncoderPresetType.superfast,
		value: EncoderPresetType.superfast,
		enable: true
	},
	{
		label: EncoderPresetType.ultrafast,
		value: EncoderPresetType.ultrafast,
		enable: true
	}
];

export interface HardwareDecodingCodeTypeInfo {
	name: string;
	type: HardwareDecodingCodeType;
	value: boolean;
}

export interface HardwareAccelerationTypeInfo {
	value: HardwareAccelerationType;
	label: string;
	enable: boolean;
}

export const encodingThreadCountSelectCommonOption = [
	{
		label: i18n.global.t('Automatic'),
		value: -1,
		enable: true
	},
	{
		label: i18n.global.t('Max'),
		value: 0,
		enable: true
	}
];

export const encodingThreadCountSelectMax = 16;

export abstract class HardwareAccelerationTypeClass {
	type: HardwareAccelerationType;
	decodingCodeList: HardwareDecodingCodeTypeInfo[];

	support_enableDecodingColorDepth10Hevc = true;
	support_enableDecodingColorDepth10Vp9 = true;
	support_enableHardwareEncoding = true;
	support_enableIntelLowPowerH264HwEncoder = false;
	support_enableIntelLowPowerHevcHwEncoder = false;
	support_enableEnhancedNvdecDecoder = false;
	support_preferSystemNativeHwDecoder = false;

	enableDecodingColorDepth10Hevc = false;

	enableDecodingColorDepth10Vp9 = false;

	enableHardwareEncoding = false;

	enableIntelLowPowerH264HwEncoder = false;

	enableIntelLowPowerHevcHwEncoder = false;

	enableEnhancedNvdecDecoder = false;

	preferSystemNativeHwDecoder = false;

	abstract initDecodingCodeList(
		hardwareDecodingCodecs: string[]
	): HardwareDecodingCodeTypeInfo[];

	constructor(props?: {
		HardwareDecodingCodecs: string[];
		EnableDecodingColorDepth10Hevc: boolean;
		EnableDecodingColorDepth10Vp9: boolean;
		EnableHardwareEncoding: boolean;
		EnableIntelLowPowerH264HwEncoder: boolean;
		EnableIntelLowPowerHevcHwEncoder: boolean;
		EnableEnhancedNvdecDecoder: boolean;
		PreferSystemNativeHwDecoder: boolean;
	}) {
		if (props) {
			this.decodingCodeList = this.initDecodingCodeList(
				props.HardwareDecodingCodecs
			);
			this.enableDecodingColorDepth10Hevc =
				props.EnableDecodingColorDepth10Hevc;
			this.enableDecodingColorDepth10Vp9 = props.EnableDecodingColorDepth10Vp9;
			this.enableHardwareEncoding = props.EnableHardwareEncoding;
			this.enableIntelLowPowerH264HwEncoder =
				props.EnableIntelLowPowerH264HwEncoder;
			this.enableIntelLowPowerHevcHwEncoder =
				props.EnableIntelLowPowerHevcHwEncoder;

			this.enableEnhancedNvdecDecoder = props.EnableEnhancedNvdecDecoder;
			this.preferSystemNativeHwDecoder = props.PreferSystemNativeHwDecoder;
		}
	}
}

export interface VideoConfigData {
	HardwareDecodingCodecs: string[];
	EnableDecodingColorDepth10Hevc: boolean;
	EnableDecodingColorDepth10Vp9: boolean;
	EnableHardwareEncoding: boolean;
	EnableIntelLowPowerH264HwEncoder: boolean;
	EnableIntelLowPowerHevcHwEncoder: boolean;
	EnableEnhancedNvdecDecoder: boolean;
	HardwareAccelerationType: string;
	AllowHevcEncoding: boolean;
	AllowAv1Encoding: boolean;
	EncodingThreadCount: number;
	TranscodingTempPath: string;
	EnableAudioVbr: boolean;
	DownMixAudioBoost: number;
	DownMixStereoAlgorithm: DownMixStereoAlgorithmType;
	EncoderPreset: EncoderPresetType;
	H265Crf: number;
	H264Crf: number;
	EnableThrottling: boolean;
	ThrottleDelaySeconds: number;
	EnableSegmentDeletion: boolean;
	SegmentKeepSeconds: number;
	MaxMuxingQueueSize: number;
	PreferSystemNativeHwDecoder: boolean;
}

export interface VideoService {
	supportHardwareAccelerationTypeList: HardwareAccelerationTypeInfo[];
	hardwareAccelerationTypeClass: HardwareAccelerationTypeClass;
	initData: VideoConfigData;
}
