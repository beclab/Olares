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
				:label="t('share.share_on_x')"
			/>
			<bio-share-item
				@click="shareToFacebook"
				v-close-popup="true"
				:platform="SOCIAL_TYPE.FACEBOOK"
				:label="t('share.share_on_facebook')"
			/>
			<bio-share-item
				@click="shareToLinkedin"
				v-close-popup="true"
				:platform="SOCIAL_TYPE.LINKEDIN"
				:label="t('share.share_on_linkedin')"
			/>
			<bio-share-item
				@click="index = 2"
				v-close-popup="false"
				icon="sym_r_qr_code"
				:label="t('share.share_via_qr_code')"
			/>
			<bio-share-item
				@click="index = 3"
				v-close-popup="false"
				icon="sym_r_imagesmode"
				:label="t('share.share_profile_picture')"
			/>
		</q-list>
		<div
			v-else-if="index === 2"
			class="column items-center"
			style="width: 300px"
		>
			<qrcode-vue
				class="bio-share-qrcode"
				id="share-qrcode"
				:value="url"
				size="136"
			/>
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
				title.value = t('share.share_your_profile');
				label.value = t('share.share_your_profile_desc');
				break;
			case 2:
				back.value = true;
				title.value = t('share.share_via_qr_code');
				label.value = t('share.share_via_qr_code_desc');
				break;
			case 3:
				back.value = true;
				title.value = t('share.share_profile_picture');
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

//只有twitter能带text
const shareToTwitter = () => {
	const content = 'Welcome to my profile to learn more about me';
	const url = encodeURIComponent(props.url);
	// 分享后会显示 “via @张三”
	const via = 'Terminus';
	const text = `${content} ${url}&via=${via}`;
	const twitter = `https://twitter.com/intent/tweet?text=${text}`;
	window.open(twitter, '_blank');
};

const shareToLinkedin = () => {
	const url = encodeURIComponent(props.url);
	const linkedin = `https://www.linkedin.com/shareArticle?mini=true&url=${url}`;
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
}

.bio-share-picture {
	margin-top: 12px;
	margin-bottom: 12px;
}
</style>
