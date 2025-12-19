import {
	HardwareAccelerationTypeClass,
	HardwareDecodingCodeType,
	HardwareDecodingCodeTypeRecord,
	HardwareAccelerationType
} from '../../abstractions/video/service';

export class HardwareRKMPPClass extends HardwareAccelerationTypeClass {
	type = HardwareAccelerationType.Rockchip_MPP;

	initDecodingCodeList(hardwareDecodingCodecs: string[]) {
		return [
			HardwareDecodingCodeType.H_264,
			HardwareDecodingCodeType.HEVC,
			HardwareDecodingCodeType.MPEG1,
			HardwareDecodingCodeType.MPEG2,
			HardwareDecodingCodeType.MPEG4,
			HardwareDecodingCodeType.VP_8,
			HardwareDecodingCodeType.VP_9,
			HardwareDecodingCodeType.AV_1
		].map((item) => {
			return {
				name: HardwareDecodingCodeTypeRecord[item],
				type: item,
				value: hardwareDecodingCodecs.find((e) => e == item) != undefined
			};
		});
	}
}
