export enum LocalVPNSDKStatus {
	NoState = 0,
	InUseOtherUser = 1,
	NeedsLogin = 2,
	NeedsMachineAuth = 3,
	Stopped = 4,
	Starting = 5,
	Running = 6
}

export enum TermiPassVpnStatus {
	off = 0,
	on = 1,
	Invalid = 2,
	connecting = 3,
	disconnecting = 4
}

const LocalVPNSDKStatusMap: Record<string, LocalVPNSDKStatus> = {
	NoState: LocalVPNSDKStatus.NoState,
	InUseOtherUser: LocalVPNSDKStatus.InUseOtherUser,
	NeedsLogin: LocalVPNSDKStatus.NeedsLogin,
	NeedsMachineAuth: LocalVPNSDKStatus.NeedsMachineAuth,
	Stopped: LocalVPNSDKStatus.Stopped,
	Starting: LocalVPNSDKStatus.Starting,
	Running: LocalVPNSDKStatus.Running
};

export const localVPNSDKStatusStringToStatus = (statusString: string) => {
	return LocalVPNSDKStatusMap[statusString] || LocalVPNSDKStatus.NoState;
};

export interface ConfigVPNInterface {
	vpnDiscriminator: 'ConfigVPNInterface' | string;
	vpnOpen(options: {
		authKey: string;
		server: string;
		acceptDns: boolean;
	}): Promise<void>;
	vpnStop(): Promise<void>;
	currentNodeId(): Promise<string>;
	hostPeerInfo(): Promise<HostPeerInfo | undefined>;
	resendCache(options: { server: string }): Promise<void>;
}

export const instanceOfConfigVPNInterface = (obj: any) => {
	if (!obj.vpnDiscriminator) {
		return false;
	}
	return obj.vpnDiscriminator === 'ConfigVPNInterface';
};

export interface TermiPassVpnStatusInterface {
	name: string;
	icon: string;
	color: string;
	description: string;
	status: boolean;
}

export interface HostPeerInfo {
	PrimaryRoutes: string[];
	CurAddr: string;
	Relay: string;
	Self: {
		Addrs: string[];
	};
}

export const TermiPassVpnStatusInfo: Record<
	TermiPassVpnStatus,
	TermiPassVpnStatusInterface
> = {
	[TermiPassVpnStatus.off]: {
		icon: 'sym_r_gpp_maybe',
		name: 'Off',
		color: '#B2B0AF',
		description: 'vpn.off.description',
		status: false
	},
	[TermiPassVpnStatus.on]: {
		icon: 'sym_r_verified_user',
		name: 'On',
		color: '#29CC5F',
		description: 'vpn.on.description',
		status: true
	},
	[TermiPassVpnStatus.Invalid]: {
		icon: 'sym_r_gpp_bad',
		name: 'Invalid',
		color: '#FF4D4D',
		description: 'vpn.invalid.description',
		status: false
	},
	[TermiPassVpnStatus.connecting]: {
		icon: 'sym_r_arming_countdown',
		name: 'Connecting',
		color: '$blue-4',
		description: 'vpn.connecting.description',
		status: false
	},
	[TermiPassVpnStatus.disconnecting]: {
		icon: 'sym_r_arming_countdown',
		name: 'Disconnecting',
		color: '$red-4',
		description: 'vpn.disconnecting.description',
		status: false
	}
};
