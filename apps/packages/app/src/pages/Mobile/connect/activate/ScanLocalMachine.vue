<template>
	<div class="column items-center" style="width: 100%">
		<div class="boot_justify">
			<q-img :src="getRequireImage(icon)" class="wizard-content__image" />
		</div>
		<div class="wizard-content__title">{{ titleStr }}</div>
		<div class="wizard-content__detail" v-html="reminderContentStr" />
	</div>
	<!-- <div v-else class="column items-center" style="width: 100%">
		<div class="boot_justify">
			<q-img
				src="../../../../assets/wizard/machine-not-found.svg"
				class="wizard-content__image"
			/>
		</div>
		<div class="wizard-content__title">
			{{ t('Olares not found') }}
		</div>
		<div class="wizard-content__detail">
			{{
				reminderContent && reminderContent.length > 0
					? reminderContent
					: t(
							'Make sure your Olares device is powered on and connected to the same wi-Fi network as your phone'
					  )
			}}
		</div>
	</div> -->
</template>

<script lang="ts" setup>
import { computed } from 'vue';

import { getRequireImage } from '../../../../utils/imageUtils';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	title: {
		type: String,
		default: '',
		required: false
	},
	reminderContent: {
		type: String,
		requred: false,
		default: ''
	},
	icon: {
		type: String,
		default: 'wizard/machine-scaning.svg',
		required: false
	}
});

const { t } = useI18n();

const titleStr = computed(() => {
	if (!!props.title) {
		return props.title;
	}
	return t('Scanning Olares in LAN');
});

const reminderContentStr = computed(() => {
	if (!!props.reminderContent) {
		return props.reminderContent;
	}
	return t('Your Phone and Olares must be in the same network');
});
</script>
