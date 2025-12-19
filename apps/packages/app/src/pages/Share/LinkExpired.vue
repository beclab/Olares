<template>
	<div class="column items-center justify-center text-negative">
		<div
			class="expired_icon_bg bg-background-3 row items-center justify-center"
		>
			<q-img src="share/link_expired.svg " width="44px" height="44px" />
		</div>
		<div class="text-h5 q-mt-sm">
			{{ t('share.Link expired') }}
		</div>
		<div class="text-overline q-mt-sm">
			{{
				t('share.Expired on {time}', {
					time
				})
			}}
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
defineProps({
	time: {
		type: String,
		required: false,
		default: ''
	}
});

const { t } = useI18n();
</script>

<style scoped lang="scss">
.expired_icon_bg {
	height: 88px;
	width: 88px;
	border-radius: 44px;
}
</style>
