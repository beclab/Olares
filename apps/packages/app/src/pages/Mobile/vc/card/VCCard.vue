<template>
	<div class="vc-card">
		<div :id="'preview' + id"></div>
	</div>
</template>

<script lang="ts">
import { defineComponent, nextTick, onMounted, watch, PropType } from 'vue';
import { VerifiableCredential } from './VerifiableCredential.mjs';
import { base64ToString } from '@didvault/sdk/src/core';
import { getRequireImage } from '../../../../utils/imageUtils';
import { VCCardInfo } from '../../../../utils/vc';

export default defineComponent({
	name: 'VCCard',
	props: {
		data: Object as PropType<VCCardInfo>,
		id: String
	},
	components: {},
	setup(props) {
		let hasData = false;

		async function updateCard(info: VCCardInfo | undefined) {
			if (!info || hasData) {
				return;
			}

			hasData = true;
			try {
				let previewElement = new VerifiableCredential();
				previewElement.showQR = false;
				previewElement.showAvatar = true;

				const vc = JSON.parse(
					base64ToString(info.verifiable_credential.split('.')[1])
				).vc;
				const manifest = JSON.parse(base64ToString(info.manifest));

				previewElement.data = vc;
				previewElement.manifest = manifest;
				if (vc.credentialSubject[`${props.data?.type.toLowerCase()}_name`]) {
					previewElement.name =
						vc.credentialSubject[`${props.data?.type.toLowerCase()}_name`];
				}

				if (vc.credentialSubject['profile_image']) {
					previewElement.avatar = vc.credentialSubject['profile_image'];
				}

				previewElement.defaultAvatar = getRequireImage(
					'account/default_avatar.svg'
				);

				const preview = document.getElementById('preview' + props.id);

				if (!preview) {
					throw Error('preview not found');
				}

				nextTick(() => {
					preview.appendChild(previewElement);
				});
			} catch (error) {
				hasData = false;
			}
		}

		watch(
			() => props.data,
			async (newVal: VCCardInfo | undefined) => {
				await updateCard(newVal);
			},
			{ immediate: true, deep: true }
		);

		onMounted(async () => {
			await updateCard(props.data);
		});

		return {};
	}
});
</script>

<style lang="scss" scoped>
.vc-card {
	width: 100%;
	height: auto;
	background: linear-gradient(
			180deg,
			rgba(222, 235, 255, 0.4) 0%,
			rgba(134, 183, 255, 0.4) 100%
		),
		linear-gradient(
			91.55deg,
			#ebf3ff 1.07%,
			#c5ddff 22.09%,
			#d9e7ff 53.16%,
			#bad6ff 83.15%,
			#dae8fd 99.07%
		);
	padding: 5px;
	border-radius: 8px;
}
</style>
