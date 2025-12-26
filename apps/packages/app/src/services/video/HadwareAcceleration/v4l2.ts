import {
	HardwareAccelerationTypeClass,
	HardwareDecodingCodeType,
	HardwareDecodingCodeTypeRecord,
	HardwareAccelerationType
} from '../../abstractions/video/service';

export class HardwareV4l2Class extends HardwareAccelerationTypeClass {
	type = HardwareAccelerationType.Video4Linux2;

	support_enableDecodingColorDepth10Hevc = false;
	support_enableDecodingColorDepth10Vp9 = false;

	initDecodingCodeList(hardwareDecodingCodecs: string[]) {
		return [HardwareDecodingCodeType.H_264].map((item) => {
			return {
				name: HardwareDecodingCodeTypeRecord[item],
				type: item,
				value: hardwareDecodingCodecs.find((e) => e == item) != undefined
			};
		});
	}
}
