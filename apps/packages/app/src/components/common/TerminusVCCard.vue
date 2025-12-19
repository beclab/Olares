<template>
	<div
		class="vc-card row justify-between items-center"
		:style="{ '--color': cardStyleRef.bgColor, width: '100%' }"
	>
		<div class="vc-card__layout column justify-center items-start">
			<img class="vc-card__brand" :src="brandRef" />
			<div
				class="vc-card__name q-mt-sm"
				:class="item ? 'text-ink-1 text-h6' : 'login-ink-2'"
			>
				{{ nameRef }}
			</div>
		</div>
		<terminus-avatar v-if="item" :size="40" :image-uri="avatarRef" />
		<div class="vc-card__avatar row justify-center items-center" v-else>
			<q-icon name="sym_r_add" size="20px" color="text-ink-2" />
		</div>
	</div>
</template>

<script setup lang="ts">
import { base64ToString } from '@didvault/sdk/src/core';
import { PropType, ref } from 'vue';
import { VCCardItem } from '../../utils/vc';
import TerminusAvatar from '../TerminusAvatar.vue';
import { TERMINUS_VC_TYPE } from '../../utils/constants';
import { useI18n } from 'vue-i18n';
import { getRequireImage } from '../../utils/imageUtils';

interface VCCardStyle {
	type: TERMINUS_VC_TYPE;
	brand: string;
	bgColor: string;
}

const cardList: VCCardStyle[] = [
	{
		type: TERMINUS_VC_TYPE.FACEBOOK,
		brand: 'facebook_brand.svg',
		bgColor:
			'linear-gradient(180deg, rgba(222, 235, 255, 0.4) 0%, rgba(134, 183, 255, 0.4) 100%),linear-gradient(91.55deg, #EBF3FF 1.07%, #C5DDFF 22.09%, #D9E7FF 53.16%, #BAD6FF 83.15%, #DAE8FD 99.07%)'
	},
	{
		type: TERMINUS_VC_TYPE.TWITTER,
		brand: 'twitter_brand.svg',
		bgColor:
			'linear-gradient(180deg, rgba(241, 251, 255, 0.3) 0%, rgba(133, 211, 255, 0.3) 100%),linear-gradient(91.55deg, #F5FAFF 1.07%, #C9F0FF 22.09%, #D6F3FF 53.16%, #BEEEFF 83.15%, #E5F9FF 99.07%)'
	},
	{
		type: TERMINUS_VC_TYPE.GOOGLE,
		brand: 'google_brand.svg',
		bgColor:
			'linear-gradient(91.55deg, #FEFFEC 1.07%, #FEFFD3 22.09%, #FBFEE0 53.16%, #FAFFC5 83.15%, #F9FFE1 99.07%),linear-gradient(180deg, rgba(255, 255, 255, 0.3) 0%, rgba(244, 255, 179, 0.3) 100%)'
	},
	{
		type: TERMINUS_VC_TYPE.CHANNEL,
		brand: 'google_brand.svg',
		bgColor: '#FFD3ED'
	}
];

const props = defineProps({
	item: {
		type: Object as PropType<VCCardItem>,
		required: false
	},
	type: {
		type: Object as PropType<TERMINUS_VC_TYPE>,
		required: false
	},
	subText: {
		type: String,
		required: false,
		default: ''
	}
});

const cardStyleRef = ref();
const brandRef = ref();
const nameRef = ref();
const avatarRef = ref();
const subjectRef = ref();
const { t } = useI18n();

if (props.item) {
	const vc = JSON.parse(
		base64ToString(props.item.verifiable_credential.split('.')[1])
	).vc;

	subjectRef.value = vc.credentialSubject;

	nameRef.value = subjectRef.value[`${props.item.type.toLowerCase()}_name`];
	avatarRef.value = subjectRef.value['profile_image'];

	const card = cardList.find((value) => {
		return value.type === props.item.type;
	});

	if (card) {
		cardStyleRef.value = card;
		brandRef.value = getRequireImage(`secret/${card.brand}`);
	}
}

if (props.type) {
	nameRef.value = props.subText || t('add_vc');
	avatarRef.value = '';

	const card = cardList.find((value) => {
		return value.type === props.type;
	});

	if (card) {
		cardStyleRef.value = card;
		brandRef.value = getRequireImage(`secret/${card.brand}`);
	}
}
</script>

<style lang="scss" scoped>
.vc-card {
	width: 100%;
	height: 92px;
	background: var(--color, #26e2c7);
	padding: 20px;
	border-radius: 8px;

	&__layout {
		width: calc(100% - 40px);
		height: 100%;
	}

	&__brand {
		height: 20px;
		width: auto;
	}

	&__name {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		color: $ink-1;
	}

	&__avatar {
		height: 40px;
		background: $background-3;
		width: 40px;
		border-radius: 50%;
	}
}
</style>
