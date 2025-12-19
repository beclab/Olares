import { defineStore } from 'pinia';
import { useUserStore } from './user';
import { axiosInstanceProxy } from '../platform/httpProxy';
import { AxiosInstance } from 'axios';
import { utcToDate } from '../utils/utils';
import { getAppPlatform } from '../application/platform';
import {
	ConfigVPNInterface,
	HostPeerInfo,
	instanceOfConfigVPNInterface,
	TermiPassVpnStatus
} from '../platform/terminusCommon/terminusCommonInterface';
import { app } from '../globals';
import { PreAuthKey } from '@didvault/sdk/src/core/api';
import { useTermipassStore } from './termipass';
import { UserStatusActive } from 'src/utils/checkTerminusState';
import { TermiPassStatus } from 'src/utils/termipassState';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';
import { i18n } from 'src/boot/i18n';

import ipaddr from 'ipaddr.js';

export type DataState = {
	owner: string;
	authKey: PreAuthKey | undefined;
	scaleServer: string;
	instance?: AxiosInstance;
	vpnStatus: TermiPassVpnStatus;
	hostPeerInfo?: HostPeerInfo;
};

export const useScaleStore = defineStore('scale', {
	state: () => {
		return {
			authKey: undefined,
			scaleServer: '',
			vpnStatus: TermiPassVpnStatus.off
		} as DataState;
	},

	getters: {
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
			if (this.hostPeerInfo.Relay && !this.hostPeerInfo.CurAddr) {
				return false;
			} else if (this.hostPeerInfo.CurAddr) {
				return true;
			}
			return false;
		},
		isLocal(): boolean {
			if (!this.isOn || !this.hostPeerInfo) {
				return false;
			}
			if (!this.hostPeerInfo.Relay || !this.hostPeerInfo.CurAddr) {
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
		}
	},

	actions: {
		async init() {
			const userStore = useUserStore();
			if (!userStore.current_user) {
				this.instance = undefined;
				return;
			}
			this.owner = userStore.current_user.name;
			this.scaleServer = userStore.getModuleSever('headscale');
			const data = await this.getDecryptData(this.owner);
			this.instance = axiosInstanceProxy({
				baseURL: this.scaleServer,
				timeout: 10000,
				headers: {
					'Content-Type': 'application/json'
				}
			});
			if (data) {
				if (typeof data == 'string') {
					this.authKey = new PreAuthKey().fromJSON(data);
				} else {
					this.authKey = data;
				}
			}
		},
		async setEncryptData(key: string, value: any) {
			await getAppPlatform().userStorage.setItem(key, value);
		},

		async getDecryptData(key: string) {
			return await getAppPlatform().userStorage.getItem(key);
		},

		async reLogin(): Promise<boolean> {
			const userStore = useUserStore();
			if (!(await userStore.unlockFirst())) {
				return false;
			}

			if (userStore.current_user?.name !== this.owner) {
				await this.init();
			}
			if (
				!this.authKey ||
				utcToDate(this.authKey.expiration).getTime() <= new Date().getTime() ||
				this.authKey.olares_device_id !=
					userStore.current_user?.olares_device_id
			) {
				try {
					getAppPlatform().getQuasar()?.loading.show();
					const data = await app.getPreAuthKey(
						userStore.current_user!.access_token
					);
					getAppPlatform().getQuasar()?.loading.hide();
					if (!data) {
						throw Error('get preauthKey error');
					}
					if (userStore.current_user?.olares_device_id) {
						data.olares_device_id = userStore.current_user.olares_device_id;
					}
					this.authKey = data;
					await this.setEncryptData(this.owner, data.toJSON());
				} catch (e) {
					notifyFailed(e.message);
					getAppPlatform().getQuasar()?.loading.hide();
					console.error('get preauthKey error', e);
				}
			}

			return (
				!!this.authKey &&
				utcToDate(this.authKey.expiration).getTime() > new Date().getTime() &&
				!!this.scaleServer
			);
		},
		async start() {
			if (!this.notifyUserCannotCorrespondMethod()) {
				this.vpnStatus = TermiPassVpnStatus.off;
				return;
			}
			if (await this.reLogin()) {
				if (!this.authKey) {
					this.vpnStatus = TermiPassVpnStatus.off;
					return;
				}
				const platform = getAppPlatform();
				const userStore = useUserStore();
				const info = {
					authKey: this.authKey.key,
					server: userStore.getModuleSever(
						'headscale',
						undefined,
						undefined,
						false
					),
					acceptDns: userStore.current_user?.isLargeVersion12 || false
				};

				if (instanceOfConfigVPNInterface(platform)) {
					(platform as any as ConfigVPNInterface).vpnOpen(info);
					this.vpnStatus = TermiPassVpnStatus.connecting;

					setTimeout(() => {
						if (this.vpnStatus == TermiPassVpnStatus.connecting) {
							this.reset();
							this.vpnStatus = TermiPassVpnStatus.Invalid;
						}
					}, 60000);
				} else {
					this.vpnStatus = TermiPassVpnStatus.off;
				}
			} else {
				this.vpnStatus = TermiPassVpnStatus.off;
			}
		},
		async stop() {
			const platform = getAppPlatform();
			if (instanceOfConfigVPNInterface(platform)) {
				if (this.isOn || this.isConnecting) {
					const userStore = useUserStore();
					if (userStore.current_user) userStore.current_user!.isLocal = false;
					await (platform as any as ConfigVPNInterface).vpnStop();
					this.vpnStatus = TermiPassVpnStatus.disconnecting;
					setTimeout(() => {
						this.vpnStatus = TermiPassVpnStatus.off;
					}, 5000);
				} else {
					this.vpnStatus = TermiPassVpnStatus.off;
				}
			}
		},
		reset() {
			if (this.owner) {
				this.stop();
				this.owner = '';
				this.scaleServer = '';
				this.authKey = undefined;
			}
			this.init();
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
		notifyUserCannotCorrespondMethod() {
			const termiPassStore = useTermipassStore();
			if (
				termiPassStore.totalStatus?.isError != UserStatusActive.active &&
				termiPassStore.totalStatus?.status != TermiPassStatus.RequiresVpn
			) {
				notifyFailed(
					i18n.global.t('the_current_status_this_module_cannot_be_accessed', {
						status: termiPassStore.totalStatus?.title
					})
				);
				return false;
			}
			return true;
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
					)
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

// 更准确的方式：显式检查私有范围
function isPrivateIP(ip: string) {
	try {
		const addr = ipaddr.parse(ip);
		return addr.range() !== 'unicast'; // unicast = 公网，其他（private, loopback, etc）= 私有
	} catch (e) {
		return false;
	}
}
