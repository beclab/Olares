import {
	HardwareAccelerationTypeClass,
	HardwareDecodingCodeType,
	HardwareDecodingCodeTypeRecord,
	HardwareAccelerationType
} from '../../abstractions/video/service';

export class HardwareToolBoxClass extends HardwareAccelerationTypeClass {
	type = HardwareAccelerationType.Apple_VideoToolBox;

	initDecodingCodeList(hardwareDecodingCodecs: string[]) {
		return [
			HardwareDecodingCodeType.H_264,
			HardwareDecodingCodeType.HEVC,
			HardwareDecodingCodeType.VP_8,
			HardwareDecodingCodeType.VP_9
		].map((item) => {
			return {
				name: HardwareDecodingCodeTypeRecord[item],
				type: item,
				value: hardwareDecodingCodecs.find((e) => e == item) != undefined
			};
		});
	}
}
