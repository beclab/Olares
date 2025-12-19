<template>
	<div
		v-if="social && social.data"
		class="simple-social-view"
		:style="{ justifyContent: justify }"
	>
		<template v-for="item in social.data" :key="item.url">
			<div
				class="bio-social"
				:class="clickable ? 'cursor-pointer' : ''"
				:style="{ color: color }"
				@click="getSocialUrl(item)"
			>
				<social-svg
					:width="socialSize"
					:height="socialSize"
					:platform="item.platform"
				/>
			</div>
		</template>
	</div>
</template>

<script lang="ts" setup>
import SocialSvg from '@apps/profile/src/components/profile/social/SocialSvg';
import { computed, PropType } from 'vue';
import { SIZE_TYPE, Social, UserSocial } from '@apps/profile/src/types/User';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
const props = defineProps({
	social: {
		type: Object as PropType<UserSocial>,
		required: true
	},
	color: {
		type: String,
		require: true
	},
	clickable: {
		type: Boolean,
		require: true
	},
	justify: {
		type: String,
		default: 'center'
	}
});

const { t } = useI18n();

const socialSize = computed(() => {
	switch (props.social.size) {
		case SIZE_TYPE.SMALL:
			return '2em';
		case SIZE_TYPE.MEDIUM:
			return '2.5em';
		case SIZE_TYPE.LARGER:
			return '3em';
		default:
			return '2.5em';
	}
});

function getSocialUrl(item: Social) {
	if (!props.clickable) {
		return;
	}
	try {
		if (item.username) {
			if (item.url.includes('${username}')) {
				const socialUrl = item.url.replace('${username}', item.username);
				window.open(socialUrl);
			} else if (item.username) {
				const hasProtocol = /^https?:\/\//i.test(item.username);
				const absoluteUrl = hasProtocol
					? item.username
					: 'https://' + item.username;
				window.open(absoluteUrl);
			}
		}
	} catch (e) {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('social.failed_to_open_social_link')
		});
	}
}
</script>

<style scoped lang="scss">
.simple-social-view {
	width: 100%;
	display: flex;
	margin-top: 0.75em;
	flex-wrap: wrap;

	.bio-social {
		margin: 0 0.375em;
	}
}
</style>
