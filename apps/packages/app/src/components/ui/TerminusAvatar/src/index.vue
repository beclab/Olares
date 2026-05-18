<template>
	<div :style="{ width: size + 'px', height: size + 'px' }">
		<img
			v-if="src"
			:src="src"
			style="width: 100%; height: 100%"
			@load="onLoad"
			@error="onError"
		/>
	</div>
</template>
<script lang="ts">
import { defineComponent, ref, PropType, watch, inject, computed } from 'vue';
import { account_es6 } from 'terminus-sdk-es6';

import getAvatarAddress = account_es6.getAvatarAddress;
import { getCacheImg, saveCacheImg, delay } from './cache';

interface TerminusInfo {
	terminusName: string;
	wizardStatus?: string;
	selfhosted?: boolean;
	tailScaleEnable?: boolean;
	osVersion?: string;
	avatar?: string;
	loginBackground?: string;
	terminusId: string;
}

export default defineComponent({
	name: 'TerminusAvatar',
	props: {
		info: {
			type: Object as PropType<TerminusInfo>,
			require: false
		},
		size: {
			type: Number,
			default: 48
		},
		isMe: {
			type: Boolean,
			default: true
		},
		useGlobalCDN: {
			type: Boolean,
			default: true
		},
		useCache: {
			type: Boolean,
			default: false
		}
	},
	components: {},
	setup(props: any) {
		const src = ref();

		const displaySrc = ref('');

		const defaultCacheAvatar = inject<boolean>('defaultCacheAvatar');

		const cacheAvatar = computed(() => {
			let classes = (props.useCache as boolean) || defaultCacheAvatar || false;
			return classes;
		});

		watch(
			() => props.info,
			(newValue) => {
				if (newValue) {
					const url = getAvatarAddress(
						newValue,
						props.useGlobalCDN,
						props.isMe
					);
					if (!cacheAvatar.value) {
						src.value = url;
					} else {
						displaySrc.value = url;
						onLoad();
					}
				}
			},
			{
				immediate: true,
				deep: true
			}
		);

		async function onLoad() {
			if (!cacheAvatar.value) {
				return;
			}

			if (!displaySrc.value || displaySrc.value.startsWith('data:image')) {
				return;
			}

			const cached = await getCacheImg(displaySrc.value);

			if (cached && cached.expire > Date.now()) {
				src.value = cached.value;
				return;
			}

			urlToBase64(displaySrc.value)
				.then(async (base64) => {
					await saveCacheImg(base64, displaySrc.value);
					src.value = base64;
				})
				.catch((e) => {
					onError(e);
				});
		}

		async function urlToBase64(url: string, retriesLeft = 3): Promise<string> {
			try {
				await delay((3 - retriesLeft) * 5000);
				const response = await fetch(url);
				if (!response.ok) {
					throw new Error(`HTTP ${response.status}`);
				}
				return blobToDataUrl(await response.blob());
			} catch (error) {
				if (retriesLeft <= 1) {
					throw error;
				}
				return urlToBase64(url, retriesLeft - 1);
			}
		}

		async function blobToDataUrl(blob: Blob): Promise<string> {
			const result = await new Promise<string | ArrayBuffer | null>(
				(resolve, reject) => {
					const reader = new FileReader();
					reader.onloadend = () => resolve(reader.result);
					reader.onerror = () =>
						reject(reader.error ?? new Error('FileReader failed'));
					reader.readAsDataURL(blob);
				}
			);
			if (typeof result !== 'string') {
				throw new Error('Invalid data URL');
			}
			return result;
		}

		async function onError(e: any) {
			const cached = await getCacheImg(displaySrc.value);
			if (cached) {
				src.value = cached.value;
			}
		}

		return {
			src,
			onLoad,
			onError
		};
	}
});
</script>
