import {
	ComputeAvailabilityDevice,
	ComputeAvailabilityNode,
	ComputeNodeStatus,
	DeviceSupportType
} from 'src/constant/compute';
import { AcceleratorAppView } from 'src/stores/settings/accelerator';

/** A card (device) row in the compute selection dialog, built from the resume
 *  prompt's availability (operability / fit + mirrored bindings). `raw` keeps
 *  the source device so selection sub-components can call binding ops directly. */
export interface ComputeDeviceView {
	deviceId: string;
	name: string;
	supportType: DeviceSupportType | string;
	capacity: number;
	available: number;
	used: number;
	operable: boolean;
	fitLevel: string;
	boundApps: AcceleratorAppView[];
	raw: ComputeAvailabilityDevice;
}

/** A node and its devices. `raw` keeps the source node for selection ops. */
export interface ComputeNodeView {
	nodeName: string;
	gpuType: string;
	typeTags: string[];
	status: ComputeNodeStatus;
	devices: ComputeDeviceView[];
	raw: ComputeAvailabilityNode;
}
