<template>
	<div
		class="row justify-evenly account_box"
		:style="isResolution ? '' : 'min-height:auto;'"
	>
		<div
			class="q-pa-md wrap account_center"
			style="word-wrap: break-word"
			v-if="isResolution"
		>
			<div class="boot_justify">
				<animationPage
					:picture="step8Img"
					:certificate="step8CertificateImg"
					:isAnimated="true"
				/>
			</div>
			<p class="get-text">{{ t('dns_title') }}</p>
			<p class="account_please_text please_text">
				{{ t('dns_content_1') }}<br />
				{{ t('dns_content_2') }}
				<span>{{ tokenStore.get_terminus_url }}</span>
			</p>
		</div>
		<div
			class="q-pa-md;wrap account_center"
			style="word-wrap: break-word"
			v-else
		>
			<div class="boot_justify1">
				<q-img
					src="../../../../assets/wizard-web/wizard_failed.jpg"
					class="account_img"
				/>
			</div>
			<p class="get-text">{{ t('dns_failed_title') }}</p>
			<p class="account_please_text please_text">
				{{ t('dns_failed_content') }}
			</p>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import animationPage from '../animation.vue';
import { useTokenStore } from '../../../../stores/wizard-step';

export default defineComponent({
	name: 'CheckDNS',
	props: {
		step8Img: {
			type: String,
			default: '',
			required: false
		},
		step8CertificateImg: {
			type: String,
			default: '',
			required: false
		}
	},
	components: {
		animationPage
	},
	setup() {
		const { t } = useI18n();
		const tokenStore = useTokenStore();

		const isResolution = ref<boolean>(true);

		return {
			isResolution,

			tokenStore,
			t
		};
	}
});
</script>
