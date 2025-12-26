<template>
	<div
		class="vc-card"
		:style="{
			'--color': cardStyleRef.bgColor,
			'--cardHeight': showDetails ? 'auto' : '106px'
		}"
	>
		<div class="row justify-between items-start" style="width: 100%">
			<div
				class="column justify-start items-start"
				style="width: calc(100% - 72px)"
			>
				<img class="vc-card__brand" :src="brandRef" />
				<div class="vc-card__name text-subtitle2">{{ nameRef }}</div>
			</div>
			<terminus-avatar-v2 :size="minSize ? 32 : 40" :image-uri="avatarRef" />

			<q-separator class="vc-card__separator" v-if="showDetails" />
			<div
				class="row justify-start items-start"
				style="width: 100%; padding: 8px"
				v-if="showDetails"
			>
				<div
					class="row justify-start items-start"
					v-for="(field, index) in fields"
					:key="index"
					:class="field.matchParent ? 'vc-card-match_item' : 'vc-card-item'"
				>
					<div class="column justify-start items-start">
						<div class="text-body3 vc-card__label">
							{{ field.label }}
						</div>
						<div class="text-body2 vc-card__value">
							{{
								field.format
									? field.format(subjectRef[field.key])
									: subjectRef[field.key]
							}}
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { base64ToString } from '@didvault/sdk/src/core';
import { PropType, ref } from 'vue';
import { VCCardItem } from '../../../utils/vc';
import TerminusAvatarV2 from '../../../components/TerminusAvatar.vue';
import moment from 'moment';

interface VCCardStyle {
	type: string;
	brand: string;
	bgColor: string;
	fields: CardField[];
}

class CardField {
	label: string;
	key: string;
	matchParent: boolean;
	format: ((date: string) => string) | undefined;

	constructor(
		label: string,
		key = label,
		matchParent = true,
		format?: (date: string) => string
	) {
		this.label = label;
		this.key = key;
		this.matchParent = matchParent;
		this.format = format;
	}
}

const cardList: VCCardStyle[] = [
	{
		type: 'Facebook',
		brand: 'facebook_brand.svg',
		bgColor: '#80A3FF',
		fields: [new CardField('id', 'facebook_id'), new CardField('email')]
	},
	{
		type: 'Twitter',
		brand: 'twitter_brand.svg',
		bgColor: '#6AD2FF',
		fields: [
			new CardField('id', 'twitter_id'),
			new CardField('create at', 'create_at', true, (date: string) => {
				return utcToTime(date);
			}),
			new CardField('following', 'following', false),
			new CardField('followers', 'followers', false),
			new CardField('tweet', 'tweet', false),
			new CardField('listed', 'listed', false)
		]
	},
	{
		type: 'Google',
		brand: 'google_brand.svg',
		bgColor: '#FFFFFF',
		fields: [new CardField('email')]
	},
	{
		type: 'Channel',
		brand: 'google_brand.svg',
		bgColor: '#FFD3ED',
		fields: [
			new CardField('id'),
			new CardField('name'),
			new CardField('address'),
			new CardField('birthday'),
			new CardField('country'),
			new CardField('email'),
			new CardField('description', 'description', true)
		]
	}
];

const props = defineProps({
	item: {
		type: Object as PropType<VCCardItem>,
		required: true
	},
	showDetails: {
		type: Boolean,
		required: false,
		default: false
	},
	minSize: {
		type: Boolean,
		required: false
	}
});

const cardStyleRef = ref();
const brandRef = ref();
const nameRef = ref();
const avatarRef = ref();
const fields = ref();
const subjectRef = ref();

const vc = JSON.parse(
	base64ToString(props.item.verifiable_credential.split('.')[1])
).vc;

subjectRef.value = vc.credentialSubject;
nameRef.value = subjectRef.value[`${props.item.type.toLowerCase()}_name`];
avatarRef.value = subjectRef.value['profile_image'];

if (props.item) {
	const card = cardList.find((value) => {
		return value.type == props.item.type;
	});

	if (card) {
		cardStyleRef.value = card;
		fields.value = card.fields;
		brandRef.value = require('../../../assets/secret/' + card.brand);
	}
}

const utcToTime = (utc_datetime: string) => {
	const T_pos = utc_datetime.indexOf('T');
	const Z_pos = utc_datetime.indexOf('Z');
	const year_month_day = utc_datetime.substr(0, T_pos);
	const hour_minute_second = utc_datetime.substr(T_pos + 1, Z_pos - T_pos - 1);
	const new_datetime = year_month_day + ' ' + hour_minute_second;
	const date = new Date(Date.parse(new_datetime));
	return moment(date).format('YYYY/MM/DD hh:mm');
};
</script>

<style lang="scss" scoped>
.vc-card {
	width: 100%;
	height: var(--cardHeight);
	box-shadow: 0px 2px 20px rgba(13, 13, 13, 0.2);
	background: var(--color, #26e2c7);
	padding: 16px;
	backdrop-filter: blur(6.07811px);
	border-radius: 8px;

	.vc-card-item {
		width: 50%;
	}

	.vc-card-match_item {
		width: 100%;
	}

	&__brand {
		height: 20px;
		width: auto;
	}

	&__name {
		margin-top: 8px;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	&__separator {
		margin-top: 20px;
		height: 0.5px;
		width: 100%;
		margin-bottom: 10px;
	}

	&__label {
		overflow: hidden;
		white-space: nowrap;
		margin-top: 20px;
		text-overflow: ellipsis;
		text-transform: capitalize;
	}

	&__value {
		margin-top: 8px;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
}
</style>
