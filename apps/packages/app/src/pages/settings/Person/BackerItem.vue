<template>
	<div
		class="backer-item q-pa-md row justify-between items-center"
		:class="!!backer.uid ? 'cursor-pointer' : ''"
		@click="onItemClick"
	>
		<div class="row justify-start items-center">
			<div class="backer-avatar text-white text-subtitle3" :style="avatarStyle">
				{{ getFirstChar(backer.name) }}
			</div>
			<div
				class="backer-name text-subtitle3 q-ml-sm"
				v-html="highlightedName"
			/>
		</div>

		<div class="backer-index q-px-sm q-py-xs text-caption text-ink-1">
			#{{ backer.number }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, PropType } from 'vue';
import { Backer } from 'src/constant/constants';

const props = defineProps({
	backer: {
		type: Object as PropType<Backer>,
		required: true
	},
	searchKeyword: {
		type: String,
		default: ''
	}
});

const avatarStyle = computed(() => {
	const styleStr = generateAvatarBackground(props.backer.name);
	return parseStyleString(styleStr);
});

const getFirstChar = (name: string) => {
	if (!name) return '?';
	return name.trim().charAt(0).toUpperCase();
};

const parseStyleString = (styleStr: string) => {
	const styleObj: Record<string, string> = {};
	styleStr.split(';').forEach((item) => {
		const [key, value] = item.split(':').map((part) => part.trim());
		if (key && value) {
			const camelKey = key.replace(/-([a-z])/g, (_, letter) =>
				letter.toUpperCase()
			);
			styleObj[camelKey] = value;
		}
	});
	return styleObj;
};

const generateAvatarBackground = (name: string) => {
	const stylePresets = [
		`background: linear-gradient(180deg, #FF7E15, #FFDF82);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: linear-gradient(68.8deg, #5b73ff, rgba(255, 91, 94, 0.2));`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: linear-gradient(180deg, #334eff, #59b1ff);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: radial-gradient(72.76% 72.76% at 50% 50%, rgba(255, 250, 119, 0.5), rgba(255, 250, 119, 0)),radial-gradient(97.58% 88.18% at -5.57% -17.26%, #8affb9, rgba(138, 255, 185, 0)),radial-gradient(65.6% 56.26% at 108.55% 109.84%, #ff6200, rgba(255, 98, 0, 0)),linear-gradient(129.71deg, #ffaa17 0.22%, #ff7717);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: radial-gradient(129.69% 76.56% at 50% 23.44%, #b00b0b, #ff7777);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: linear-gradient(180deg, #2288f9, #5f26f7);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: linear-gradient(180deg, #ffb61f, #ff4a16);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: linear-gradient(180deg, #1fcf0e, #d6df0d);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: linear-gradient(180deg, #019fe3, #06d9ae);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: radial-gradient(428.82% 168.61% at 100% 142.19%, #d2ff60, #ff7759 58.17%, #ffae50), linear-gradient(#fff, #fff);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: linear-gradient(180deg, #0046f7, #d2a2ff);`
			.replace(/\s+/g, ' ')
			.trim(),
		`background: radial-gradient(85.34% 85.34% at 53.04% 85.34%, #171b32, rgba(112, 135, 255, 0.64)), linear-gradient(#333, #333);`
			.replace(/\s+/g, ' ')
			.trim()
	];

	if (!name || name.trim() === '') {
		return stylePresets[0];
	}

	const hash = getStringHash(name.trim());

	// const index = Math.abs(props.index % stylePresets.length);
	const index = Math.abs(hash % stylePresets.length);
	// console.log(index);
	return stylePresets[index];
};

function getStringHash(str: string): number {
	let hash = 0;
	for (let i = 0; i < str.length; i++) {
		const charCode = str.charCodeAt(i);
		hash = (hash << 5) - hash + charCode;
		hash = hash & hash;
	}
	return hash;
}

const highlightedName = computed(() => {
	try {
		if (
			!props.searchKeyword ||
			props.searchKeyword.trim() === '' ||
			!props.backer ||
			!props.backer.name
		) {
			return `<span class="text-ink-1">${props.backer.name}</span>`;
		}

		const escapedKeyword = props.searchKeyword
			.trim()
			.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
		const regex = new RegExp(`(${escapedKeyword})`, 'gi');
		return props.backer.name.replace(regex, (match) => {
			return `<span class="highlight-keyword">${match}</span>`;
		});
	} catch (e) {
		return `<span class="text-ink-1">${props.backer.name}</span>`;
	}
});

const onItemClick = () => {
	if (props.backer?.uid) {
		window.open(`https://www.kickstarter.com/profile/${props.backer?.uid}`);
	}
};
</script>

<style scoped lang="scss">
.backer-item {
	border: 1px solid $separator;
	border-radius: 8px;

	&:hover {
		background: $blue-alpha;
		border: 1px solid $info;

		.backer-index {
			background: $background-1;
		}
	}

	.backer-avatar {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.backer-name {
		flex: 1;
		min-width: 0;
		word-break: break-word;
		max-width: 170px;
	}

	.backer-index {
		border-radius: 4px;
		background: $background-3;
		flex-shrink: 0;
	}
}

:deep(.highlight-keyword) {
	color: $blue-default !important;
}
</style>
