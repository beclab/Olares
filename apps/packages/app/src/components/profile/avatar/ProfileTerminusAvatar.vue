<template>
	<div
		class="terminus-avatar-root row justify-center items-center"
		:style="{
			width: size,
			'--avatarRadius': radius,
			'--avatarBorder': `${showBorder ? 2 : 0}px solid ${borderColor}`
		}"
	>
		<div
			class="terminus-avatar-image-background"
			:style="{
				width: showBorder && isNft ? '90%' : '100%',
				height: showBorder && isNft ? '90%' : '100%',
				maskImage: mask
					? 'linear-gradient(to bottom,rgba(0, 0, 0, 1) 70%,rgba(0, 0, 0, 0) 100%)'
					: ''
			}"
		>
			<q-img
				v-if="srcResource"
				:src="srcResource"
				class="terminus-avatar-image"
			/>
			<slot />
		</div>
		<q-img
			class="terminus-avatar-board"
			v-if="showBorder && isNft"
			src="https://app.cdn.olares.com/avatar3/nft.png"
		/>
	</div>
</template>
<script lang="ts" setup>
import { ref, PropType, computed } from 'vue';
import { getNftUrl } from './encoder';
import { OlaresInfo } from '@bytetrade/core';

const props = defineProps({
	info: {
		type: Object as PropType<OlaresInfo>,
		require: false
	},
	size: {
		type: String,
		default: ''
	},
	radius: {
		type: String,
		default: '50%'
	},
	mask: {
		type: Boolean,
		default: false
	},
	isMe: {
		type: Boolean,
		default: true
	},
	useGlobalCDN: {
		type: Boolean,
		default: true
	},
	showBorder: {
		type: Boolean,
		default: false
	},
	borderColor: {
		type: String,
		default: '#1F1F1F'
	},
	showImageBorder: {
		type: Boolean,
		default: true
	}
});

const isNft = computed(() => {
	if (props.info && props.info?.avatar) {
		return getNftUrl(props.info?.avatar).isNft;
	} else {
		return false;
	}
});

function stringToIntHash(str: string, lowerbound: number, upperbound: number) {
	if (!str) {
		return lowerbound;
	}

	let result = 0;
	for (let i = 0; i < str.length; i++) {
		result = result + str.charCodeAt(i);
	}

	if (!lowerbound) lowerbound = 0;
	if (!upperbound) upperbound = 500;

	return (result % (upperbound - lowerbound)) + lowerbound;
}

const srcResource = computed(() => {
	if (!props.info) {
		return '';
	}

	// not set terminusName
	if (!props.info.olaresId && !props.info.terminusName) {
		return `https://app.cdn.olares.com/avatar3/1.png`;
	}

	if (!props.info.avatar) {
		let id = ref(
			stringToIntHash(props.info.olaresId || props.info.terminusName, 1, 36)
		);
		if (props.useGlobalCDN) {
			return `https://app.cdn.olares.com/avatar3/${id.value}.png`;
		} else {
			return (
				`https://` +
				(props.info.olaresId || props.info.terminusName).replace('@', '.') +
				`/avatar/${id.value}.png`
			);
		}
	}

	if (props.info.avatar.startsWith('http')) {
		return props.info.avatar;
	} else {
		const re = new RegExp('^[1-3]?[0-9]\\.png');
		if (re.test(props.info.avatar)) {
			console.log('re test true');
			if (props.isMe) {
				if (props.useGlobalCDN) {
					return 'https://app.cdn.olares.com/avatar3/' + props.info.avatar;
				} else {
					return (
						`https://` +
						(props.info.olaresId || props.info.terminusName).replace('@', '.') +
						`/avatar/` +
						props.info.avatar
					);
				}
			} else {
				if (props.useGlobalCDN) {
					return `https://app.cdn.olares.com/avatar3/${props.info.avatar}`;
				} else {
					return (
						`https://` +
						(props.info.olaresId || props.info.terminusName).replace('@', '.') +
						`/avatar/` +
						props.info.avatar
					);
				}
			}
		} else {
			return getNftUrl(props.info?.avatar).url;
		}
	}
});
</script>

<style lang="scss">
.terminus-avatar-root {
	position: relative;
	background: transparent;

	.terminus-avatar-image-background {
		border-radius: var(--avatarRadius);
		outline: var(--avatarBorder);
		position: relative;
		overflow: hidden;

		.terminus-avatar-image {
			width: 100%;
			height: 100%;
			border-radius: var(--avatarRadius);
		}
	}

	.terminus-avatar-board {
		position: absolute;
		top: 0;
	}
}
</style>
