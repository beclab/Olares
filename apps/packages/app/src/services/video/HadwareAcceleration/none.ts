import {
	HardwareAccelerationType,
	HardwareAccelerationTypeClass
} from '../../abstractions/video/service';

export class HardwareNoneClass extends HardwareAccelerationTypeClass {
	type = HardwareAccelerationType.None;
	support_enableDecodingColorDepth10Hevc = false;
	support_enableDecodingColorDepth10Vp9 = false;
	support_enableHardwareEncoding = false;
	decodingCodeList = [];
	initDecodingCodeList() {
		return [];
	}
}
