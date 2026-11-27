<template>
	<bio-popup
		@hide="index = 1"
		:title="title"
		:label="label"
		:back="back"
		@on-back-click="index = 1"
	>
		<q-list v-if="index === 1">
			<bio-share-item
				@click="shareToTwitter"
				v-close-popup="true"
				:platform="SOCIAL_TYPE.TWITTER"
				:label="t('share_on_x')"
			/>
			<bio-share-item
				@click="shareToFacebook"
				v-close-popup="true"
				:platform="SOCIAL_TYPE.FACEBOOK"
				:label="t('share_on_facebook')"
			/>
			<bio-share-item
				@click="shareToLinkedin"
				v-close-popup="true"
				:platform="SOCIAL_TYPE.LINKEDIN"
				:label="t('share_on_linkedin')"
			/>
			<bio-share-item
				@click="index = 2"
				v-close-popup="false"
				icon="sym_r_qr_code"
				:label="t('share_via_qr_code')"
			/>
			<bio-share-item
				@click="index = 3"
				v-close-popup="false"
				icon="sym_r_imagesmode"
				:label="t('share_profile_picture')"
			/>
		</q-list>
		<div
			v-else-if="index === 2"
			class="column items-center"
			style="width: 300px"
		>
			<!--
				The id must live on a real DOM node. qrcode.vue declares `id` as a
				component prop, so passing it to <qrcode-vue> never reaches the
				rendered <canvas> and getElementById() would return null, breaking
				the PNG download. Wrap the canvas in a div that carries the id.
			-->
			<div id="share-qrcode" class="bio-share-qrcode">
				<qrcode-vue :value="url" size="136" />
			</div>
			<png-download
				element-id="share-qrcode"
				:png-name="`${name}_qrcode.png`"
			/>
		</div>
		<div
			v-else-if="index === 3"
			class="column items-center"
			style="width: 300px"
		>
			<bio-profile-picture
				class="bio-share-picture"
				id="share-picture"
				:name="name"
				:url="url"
			/>
			<png-download element-id="share-picture" :png-name="`${name}_card.png`" />
		</div>
	</bio-popup>
</template>

<script lang="ts" setup>
import BioProfilePicture from '@apps/profile/src/components/profile/share/BioProfilePicture.vue';
import BioShareItem from '@apps/profile/src/components/profile/share/BioShareItem.vue';
import PngDownload from '@apps/profile/src/components/profile/base/PngDownload.vue';
import BioPopup from '@apps/profile/src/components/profile/base/BioPopup.vue';
import { SOCIAL_TYPE } from '@apps/profile/src/types/SocialProps';
import { ref, watch } from 'vue';
import QrcodeVue from 'qrcode.vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	name: String,
	url: String
});

const index = ref(1);
const back = ref(false);
const title = ref();
const label = ref();
const { t } = useI18n();

watch(
	() => index.value,
	(newValue) => {
		switch (newValue) {
			case 1:
				back.value = false;
				title.value = t('share_your_profile');
				label.value = t('share_your_profile_desc');
				break;
			case 2:
				back.value = true;
				title.value = t('share_via_qr_code');
				label.value = t('share_via_qr_code_desc');
				break;
			case 3:
				back.value = true;
				title.value = t('share_profile_picture');
				label.value = '';
				break;
		}
	},
	{
		immediate: true
	}
);

const shareToFacebook = () => {
	const url = encodeURIComponent(props.url);
	const facebook = `https://www.facebook.com/sharer/sharer.php?u=${url}`;
	window.open(facebook, '_blank');
};

// Only Twitter supports free text.
const shareToTwitter = () => {
	const content = t('share_profile_message');
	// Shows "via @<account>" in shared content.
	const via = 'Olares';
	const params = new URLSearchParams({
		text: `${content} ${props.url}`,
		via
	});
	const twitter = `https://twitter.com/intent/tweet?${params.toString()}`;
	window.open(twitter, '_blank');
};

const shareToLinkedin = () => {
	const url = encodeURIComponent(props.url);
	const linkedin = `https://www.linkedin.com/sharing/share-offsite/?url=${url}`;
	window.open(linkedin, '_blank');
};
</script>

<style scoped lang="scss">
.bio-share-qrcode {
	margin-top: 10px;
	padding: 8px;
	margin-bottom: 18px;
	border-radius: 8px;
	background-color: white;
	border: 1px solid $separator;
	// The wrapped <canvas> is inline; drop the baseline gap so the white box
	// hugs the QR code (and the downloaded PNG has no stray whitespace).
	font-size: 0;
	line-height: 0;
}

.bio-share-picture {
	margin-top: 12px;
	margin-bottom: 12px;
}
</style>
