import { defineStore } from 'pinia';
import { useVaultSocketStore } from './larepassVaultWebsocket';
import { useUserStore } from './user';
import { useDeviceStore } from './device';
import { useWebsocketManager2Store } from './websocketManager2';
import { useAppAbilitiesStore } from './appAbilities';
import { useLarePassWiseSocket } from './larepassWiseSocket';

export const useLarepassWebsocketManagerStore = defineStore(
	'larepassWebsocketManager',
	{
		state: () => ({
			vaultConnectedUserId: '',
			wiseConnectedUserId: ''
		}),

		actions: {
			restart() {
				const userStore = useUserStore();

				const managerStore = useWebsocketManager2Store();
				if (userStore.current_user?.isLargeVersion12) {
					managerStore.restart();
					this.startWiseSocket();
					return;
				}

				const vaultStore = useVaultSocketStore();
				if (!this.checkVaultConnectedUser()) {
					this.vaultConnectedUserId = userStore.current_id ?? '';
					vaultStore.restart();
				}

				if (!this.checkWiseConnectedUser()) {
					this.wiseConnectedUserId = userStore.current_id ?? '';
					managerStore.restart();
				}
			},
			dispose() {
				const vaultStore = useVaultSocketStore();
				const managerStore = useWebsocketManager2Store();

				vaultStore.dispose();
				managerStore.dispose();
				this.disposeWiseSocket();
			},
			checkVaultConnectedUser() {
				const vaultStore = useWebsocketManager2Store();
				const userStore = useUserStore();
				const deviceStore = useDeviceStore();

				if (!deviceStore.networkOnLine) {
					return true;
				}
				if (!userStore.connected) {
					return true;
				}
				if (
					(vaultStore.isConnected() || vaultStore.isConnecting()) &&
					this.vaultConnectedUserId == userStore.current_id
				) {
					return true;
				}
				return false;
			},
			checkWiseConnectedUser() {
				const managerStore = useWebsocketManager2Store();
				const userStore = useUserStore();
				const deviceStore = useDeviceStore();

				if (!deviceStore.networkOnLine) {
					return false;
				}
				if (!userStore.connected) {
					return false;
				}
				if (
					(managerStore.isConnected() || managerStore.isConnecting()) &&
					this.wiseConnectedUserId == userStore.current_id
				) {
					return true;
				}
				return false;
			},

			startWiseSocket() {
				const wiseWSStore = useLarePassWiseSocket();
				wiseWSStore.start();
			},
			disposeWiseSocket() {
				const wiseWSStore = useLarePassWiseSocket();
				wiseWSStore.dispose();
			}
		}
	}
);
