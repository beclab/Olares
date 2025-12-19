<template>
	<div class="float-wrap">
		<div
			class="float"
			:style="{ background: !failed ? '#93e1ff' : '#FFBBC8' }"
		></div>
		<div
			class="float"
			:style="{ background: !failed ? '#faffd9' : '#FFBBC8' }"
		></div>
	</div>

	<div v-if="!failed" class="column items-center">
		<div class="boot_justify">
			<animationPage
				:picture="dns_waikuang_image"
				:certificate="dns_image"
				:isAnimated="true"
			/>
		</div>
		<div class="wizard-content__title">{{ t('dns_resolution') }}</div>
		<div class="wizard-content__detail" v-html="reminderInfo" />
	</div>
	<div v-else class="column items-center">
		<div class="boot_justify">
			<q-img
				src="../../../../assets/wizard/failed.svg"
				class="wizard-content__image"
			/>
		</div>
		<div class="wizard-content__title">
			{{ t('dns_resolution_failed') }}
		</div>
		<div class="wizard-content__detail">
			{{ t('dns_resolution_failed_please_try_again') }}
		</div>

		<div class="wizard-content__retry">
			<q-btn
				class="wizard-content__retry__btn row items-center justify-center"
				flat
				padding="0px"
				no-caps
				@click="onFailed"
			>
				<q-icon name="sync" size="15px" class="q-mr-xs" />
				<div>
					{{ t('retry') }}
				</div>
			</q-btn>
		</div>
	</div>
</template>

<script lang="ts" setup>
import animationPage from './animation.vue';
import dns_image from '../../../../assets/wizard/dns.png';
import dns_waikuang_image from '../../../../assets/wizard/dns_waikuang.png';
import { useI18n } from 'vue-i18n';
import { i18n } from '../../../../boot/i18n';

defineProps({
	failed: {
		type: Boolean,
		required: false,
		default: false
	},
	reminderInfo: {
		type: String,
		required: false,
		default: i18n.global.t('dns_resolution_introduce')
	}
});
const { t } = useI18n();

const onFailed = () => {
	emit('dnsRetry');
};

const emit = defineEmits(['dnsRetry']);
</script>
