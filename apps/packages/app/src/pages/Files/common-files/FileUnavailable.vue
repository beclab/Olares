<template>
	<div class="column items-center justify-center file-unavailable-root">
		<!-- adfkasdkfjadsjfsdkfjaksdkjdkf -->
		<img
			src="../../../assets/images/preview_unavailable.svg"
			width="96"
			height="96"
		/>
		<div class="no-available text-body3 q-mt-md text-color-title">
			{{ t('files.no_preview_available') }}
		</div>
		<!-- <div class="q-px-md q-py-xs text-body3 go-app-store q-mt-md">
			{{ t('files.go_app_store_to_search') }}
		</div> -->
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<style scoped lang="scss">
.file-unavailable-root {
	height: 100%;
	width: 100%;

	.no-available {
		text-align: center;
	}

	.go-app-store {
		border: 1px solid $blue-4;
		border-radius: 8px;
		background: linear-gradient(
			0deg,
			rgba(81, 174, 255, 0.1),
			rgba(81, 174, 255, 0.1)
		);
		color: $blue-4;
	}
}
</style>
