import {
	HardwareAccelerationTypeClass,
	HardwareDecodingCodeType,
	HardwareDecodingCodeTypeRecord,
	HardwareAccelerationType
} from '../../abstractions/video/service';

export class HardwareAMFClass extends HardwareAccelerationTypeClass {
	type = HardwareAccelerationType.AMD_AMF;

	initDecodingCodeList(hardwareDecodingCodecs: string[]) {
		return [
			HardwareDecodingCodeType.H_264,
			HardwareDecodingCodeType.HEVC,
			HardwareDecodingCodeType.MPEG2,
			HardwareDecodingCodeType.VC_1,
			HardwareDecodingCodeType.VP_9,
			HardwareDecodingCodeType.AV_1
		].map((item) => {
			return {
				name: HardwareDecodingCodeTypeRecord[item],
				type: item,
				value:
					hardwareDecodingCodecs != undefined &&
					hardwareDecodingCodecs.find((e) => e == item) != undefined
			};
		});
	}
}
