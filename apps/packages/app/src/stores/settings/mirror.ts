import { defineStore } from 'pinia';
import axios from 'axios';
import { useAdminStore } from './admin';
import { useTokenStore } from './token';

export interface RegistryMirror {
	name: string;
	image_count: number;
	image_size: number;
	endpoints: string[] | null;
}

export type MirrorState = {
	registries: RegistryMirror[];
};

export interface RegistryImage {
	id: string;
	size: number;
	repo_tags: string[];
}

export const useMirrorStore = defineStore('mirror', {
	state: () => {
		return {
			registries: []
		} as MirrorState;
	},

	actions: {
		async getRegistryMirrors() {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/containerd/registries`,
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			console.log(data);
			// this.registries = data;
			this.registries = data.sort((a, b) => {
				return b.image_size - a.image_size;
			});
			return data;
		},
		async getRegistryEndpoint(registry: string) {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/containerd/registry/mirrors/${registry}`,
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			console.log(data);
			return data.endpoint;
		},
		async putRegistryEndpoint(registry: string, endpoint: string[]) {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			if (endpoint.length == 0) {
				return await this.deleteRegistryEndpoint(registry);
			}
			const data: any = await axios.put(
				`${tokenStore.url}/api/containerd/registry/mirrors/${registry}`,
				{
					endpoint: endpoint.length > 0 ? endpoint : null
				},
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			return data.endpoint;
		},
		async deleteRegistryEndpoint(registry: string) {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			await axios.delete(
				`${tokenStore.url}/api/containerd/registry/mirrors/${registry}`,
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			return [];
		},
		async getRegistryImages(registry?: string) {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/containerd/images`,
				{
					headers: {
						'X-Signature': admin.olares_device_id
					},
					params: {
						registry
					}
				}
			);
			if (data && data.length > 0) {
				(data as RegistryImage[]).sort((a: RegistryImage, b: RegistryImage) => {
					if (a.repo_tags.length > 0 && b.repo_tags.length > 0) {
						const aTag = a.repo_tags[0];
						const bTag = b.repo_tags[0];
						if (aTag && aTag.includes('/') && bTag && bTag.includes('/')) {
							return aTag
								.substring(aTag.indexOf('/') + 1)
								.localeCompare(bTag.substring(bTag.indexOf('/') + 1));
						}
						return a.repo_tags[0].localeCompare(b.repo_tags[0]);
					}
					return 1;
				});
			}
			return data;
		},
		async deleteRegistryImage(id: string) {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.delete(
				`${tokenStore.url}/api/containerd/images/${id}`,
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			return data;
		},
		async deleteImagesPrune() {
			const admin = useAdminStore();
			const tokenStore = useTokenStore();
			const data: any = await axios.post(
				`${tokenStore.url}/api/containerd/images/prune`,
				{
					headers: {
						'X-Signature': admin.olares_device_id
					}
				}
			);
			return data;
		}
	}
});
