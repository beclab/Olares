<template>
	<div
		v-if="isLoading"
		:class="
			readerStore.subscribed
				? 'drawer-feed-subscribed-loading'
				: 'drawer-feed-unSubscribe-loading'
		"
		class="text-body3"
	>
		<bt-loading :loading="isLoading" />
	</div>
	<div
		v-else
		:class="
			readerStore.subscribed
				? 'drawer-feed-subscribed'
				: 'drawer-feed-unSubscribe'
		"
		class="text-body3 cursor-pointer"
		@click="setCurrentSubscribe"
	>
		<bt-loading :loading="isLoading" />
		{{
			readerStore.subscribed
				? t('base.unsubscribe')
				: t('base.subscribe_this_feed')
		}}
	</div>
</template>

<script lang="ts" setup>
import BtLoading from '../base/BtLoading.vue';
import BaseCheckBoxDialog from '../base/BaseCheckBoxDialog.vue';
import { useReaderStore } from 'src/stores/rss-reader';
import { useConfigStore } from 'src/stores/rss-config';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useAbilityStore } from 'src/stores/rss-ability';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';

const { t } = useI18n();
const $q = useQuasar();
const readerStore = useReaderStore();
const configStore = useConfigStore();
const abilityStore = useAbilityStore();
const isLoading = computed(() => {
	if (readerStore.readingFeed) {
		return readerStore.subscribedLoadingSet.has(readerStore.readingFeed.id);
	} else {
		return false;
	}
});

const setCurrentSubscribe = async () => {
	if (isLoading.value) {
		return;
	}
	await abilityStore.getAbiAbility();
	if (!readerStore.subscribed && !abilityStore.rssubscribe) {
		notifyFailed(t('Rss Subscribe not installed'));
		return;
	}
	if (readerStore.subscribed) {
		const remove = configStore.feedRemoveWithFile;
		$q.dialog({
			component: BaseCheckBoxDialog,
			componentProps: {
				label: t('dialog.remove_subscription'),
				content: t('dialog.remove_subscription_desc'),
				modelValue: remove,
				showCheckbox: true,
				boxLabel: t('dialog.delete_the_files_if_present')
			}
		})
			.onOk(async (selected) => {
				console.log('remove task ok');
				configStore.setFeedRemoveWithFile(selected);
				readerStore.setCurrentSubscribe(selected);
			})
			.onCancel(() => {
				console.log('remove task cancel');
			});
	} else {
		readerStore.setCurrentSubscribe();
	}
};
</script>

<style scoped lang="scss">
.drawer-feed-unSubscribe-loading {
	margin-top: 8px;
	border-radius: 8px;
	border: 1px solid $orange-default;
	width: 100%;
	height: 32px;
	padding-top: 8px;
	text-align: center;
	color: $orange-default;
}
.drawer-feed-unSubscribe {
	@extend .drawer-feed-unSubscribe-loading;

	&:hover {
		background: $btn-bg-hover;
	}

	&:active {
		background: $btn-bg-pressed;
	}
}

.drawer-feed-subscribed-loading {
	margin-top: 8px;
	border-radius: 8px;
	border: 1px solid $btn-stroke;
	width: 100%;
	padding-top: 8px;
	text-align: center;
	height: 32px;
	color: $ink-2;
}

.drawer-feed-subscribed {
	@extend .drawer-feed-subscribed-loading;

	&:hover {
		background: $btn-bg-hover;
	}

	&:active {
		background: $btn-bg-pressed;
	}
}
</style>
