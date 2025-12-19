<template>
	<div
		class="q-py-sm q-pl-md q-pr-sm bg-background-3 cookie-message-container row no-wrap justify-between items-center flex-gap-x-sm"
	>
		<div class="text-negative text-body3">
			{{
				message
					? message
					: t('not installed. Unable to get download file for this link.', {
							app: appName
					  })
			}}
		</div>
		<q-btn
			v-if="appName"
			color="orange-default"
			padding="8px 24px"
			class="btn-wrapper"
			no-caps
			text-color="white"
			@click="onClickHandler"
		>
			<span class="text-body3">{{ t('app.install') }}</span>
		</q-btn>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useUserStore } from 'src/stores/user';
import { useConfigStore } from 'src/stores/rss-config';

const props = defineProps({
	message: {
		type: String,
		default: ''
	},
	appName: {
		type: String,
		default: ''
	}
});

const { t } = useI18n();

const onClickHandler = () => {
	if (process.env.PLATFORM_BEX_ALL) {
		const userStore = useUserStore();
		const url =
			userStore.getModuleSever('market') + '/search?keyword=' + props.appName;
		window.open(url, '_blank');
	} else {
		const configStore = useConfigStore();
		const url =
			configStore.getModuleSever('market') + '/search?keyword=' + props.appName;
		window.open(url, '_blank');
	}
};
</script>

<style lang="scss" scoped>
.cookie-message-container {
	border-radius: 12px;
	.cookie-icon {
		width: 16px;
		height: 16px;
	}
	.submit-icon {
		position: absolute;
		bottom: 0;
		right: 0;
	}
	.btn-wrapper {
		::v-deep(.q-btn__content) {
			line-height: 16px;
		}
	}
}
</style>
