import { defineStore } from 'pinia';
import { useUserStore } from './user';
import { getAppPlatform } from '../application/platform';
import {
	ConfigVPNInterface,
	HostPeerInfo,
	instanceOfConfigVPNInterface,
	TermiPassVpnStatus
} from '../platform/terminusCommon/terminusCommonInterface';

import ipaddr from 'ipaddr.js';

export type DataState = {
	owner: string;
	vpnStatus: TermiPassVpnStatus;
	hostPeerInfo?: HostPeerInfo;
	vpnAbility: boolean;
};

export const useScaleStore = defineStore('scale', {
	state: () => {
		return {
			owner: '',
			vpnStatus: TermiPassVpnStatus.off,
			vpnAbility: true
		} as DataState;
	},

	getters: {
		vpnAbility(): boolean {
			const platform = getAppPlatform();
			return platform.vpnAbility;
		},

		isOn(): boolean {
			return this.vpnStatus == TermiPassVpnStatus.on;
		},
		isConnecting(): boolean {
			return this.vpnStatus == TermiPassVpnStatus.connecting;
		},
		isDisconnecting(): boolean {
			return this.vpnStatus == TermiPassVpnStatus.disconnecting;
		},
		isDirect(): boolean {
			if (!this.isOn || !this.hostPeerInfo) {
				return false;
			}

			if (!this.hostPeerInfo.CurAddr) {
				return false;
			}

			return true;
		},
		isLocal(): boolean {
			if (!this.isOn || !this.hostPeerInfo) {
				return false;
			}
			if (!this.hostPeerInfo.CurAddr) {
				return false;
			}
			let curaddr = this.hostPeerInfo.CurAddr;

			if (curaddr.includes(']')) {
				curaddr = curaddr.split(']')[0].substring(1);
			} else if (curaddr.includes(':')) {
				curaddr = curaddr.split(':')[0];
			}
			const parsedAddr1 = ipaddr.parse(curaddr);

			if (parsedAddr1.kind() == 'ipv4') {
				// return ipaddr.parse(curaddr).range() == 'unicast';
				return isPrivateIP(curaddr);
			}

			if (
				this.hostPeerInfo.Self &&
				this.hostPeerInfo.Self.Addrs.find((e) =>
					ipv6IsLocal(
						e.includes(']')
							? e.split(']')[0].substring(1)
							: e.includes(':')
							? e.split(':')[0]
							: e,
						curaddr
					)
				) != undefined
			) {
				return true;
			}
			return false;
		},
		localAddress(): string {
			if (!this.isOn || !this.hostPeerInfo) {
				return '';
			}
			if (!this.hostPeerInfo.CurAddr) {
				return '';
			}
			let curaddr = this.hostPeerInfo.CurAddr;

			if (curaddr.includes(']')) {
				curaddr = curaddr.split(']')[0].substring(1);
			} else if (curaddr.includes(':')) {
				curaddr = curaddr.split(':')[0];
			}
			return curaddr;
		}
	},

	actions: {
		async stop(force = false) {
			const platform = getAppPlatform();
			if (instanceOfConfigVPNInterface(platform)) {
				if (force || this.isOn || this.isConnecting) {
					const userStore = useUserStore();
					if (userStore.current_user) userStore.current_user!.isLocal = false;
					this.vpnStatus = TermiPassVpnStatus.disconnecting;
					try {
						await (platform as any as ConfigVPNInterface).vpnStop();
					} finally {
						setTimeout(() => {
							if (this.vpnStatus === TermiPassVpnStatus.disconnecting) {
								this.vpnStatus = TermiPassVpnStatus.off;
							}
						}, 5000);
					}
				} else {
					this.vpnStatus = TermiPassVpnStatus.off;
				}
			}
		},
		async waitForVpnStopSettled(timeoutMs = 6000) {
			if (this.vpnStatus !== TermiPassVpnStatus.disconnecting) {
				return;
			}
			const start = Date.now();
			while (this.vpnStatus === TermiPassVpnStatus.disconnecting) {
				if (Date.now() - start >= timeoutMs) {
					break;
				}
				await new Promise((resolve) => setTimeout(resolve, 100));
			}
		},
		reset() {
			if (this.owner) {
				this.stop();
			}
			this.owner = useUserStore().current_user?.name ?? '';
		},
		async configHostPeerInfo() {
			const platform = getAppPlatform();
			if (!instanceOfConfigVPNInterface(platform)) {
				return undefined;
			}
			this.hostPeerInfo = await (
				platform as any as ConfigVPNInterface
			).hostPeerInfo();

			return this.hostPeerInfo;
		},
		async resendCache() {
			const platform = getAppPlatform();
			const userStore = useUserStore();
			if (instanceOfConfigVPNInterface(platform)) {
				await (platform as any as ConfigVPNInterface).resendCache({
					server: userStore.getModuleSever(
						'headscale',
						undefined,
						undefined,
						false
					),
					authToken: userStore.current_user?.access_token || ''
				});
			}
		}
	}
});

const ipv6IsLocal = (addr1: string, addr2: string) => {
	const prefixLength = 64; // Subnet mask /64 from your input

	// Function to check if two IPv6 addresses are in the same subnet
	function areInSameSubnet(ip1: string, ip2: string, prefixLength: number) {
		try {
			// Parse the IPv6 addresses
			const parsedAddr1 = ipaddr.parse(ip1);
			const parsedAddr2 = ipaddr.parse(ip2);
			console.log('parsedAddr1 ===>', parsedAddr1);
			console.log('parsedAddr2 ===>', parsedAddr2);

			// Ensure both are IPv6
			if (parsedAddr1.kind() !== 'ipv6' || parsedAddr2.kind() !== 'ipv6') {
				console.error('One or both addresses are not valid IPv6 addresses.');
				return false;
			}

			// Check if addresses are in the same subnet
			const isSameSubnet = parsedAddr1.match(parsedAddr2, prefixLength);

			return isSameSubnet;
		} catch (e) {
			console.error('Error parsing IP addresses:', e.message);
			return false;
		}
	}

	// Execute the check
	const result = areInSameSubnet(addr1, addr2, prefixLength);
	console.log(
		`Are ${addr1} and ${addr2} in the same /${prefixLength} subnet? ${result}`
	);
	return result;
};

// More accurate approach: explicitly check private ranges.
function isPrivateIP(ip: string) {
	try {
		const addr = ipaddr.parse(ip);
		return addr.range() !== 'unicast'; // unicast = public; others (private, loopback, etc) = private
	} catch (e) {
		return false;
	}
}
