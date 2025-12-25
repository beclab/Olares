<template>
	<title-right-layout>
		<q-btn
			class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
			color="ink-2"
			outline
			no-caps
			icon="sym_r_keyboard_double_arrow_left"
			:disable="!readerStore.canPreEntryRoute(readerStore.readingEntry)"
			@click="
				updateEntry(readerStore.getPreEntryId(readerStore.readingEntry), 'back')
			"
		>
			<bt-tooltip :label="t('base.previous')" :hotkey="WISE_HOTKEY.ENTRY.PRE" />
		</q-btn>
		<q-btn
			class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
			color="ink-2"
			outline
			no-caps
			icon="sym_r_keyboard_double_arrow_right"
			:disable="!readerStore.canNextEntryRoute(readerStore.readingEntry)"
			@click="updateEntry(readerStore.getNextEntryId(readerStore.readingEntry))"
		>
			<bt-tooltip :label="t('base.next')" :hotkey="WISE_HOTKEY.ENTRY.NEXT" />
		</q-btn>
		<q-btn
			class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
			color="ink-2"
			outline
			no-caps
			icon="sym_r_step"
			:disable="
				!readerStore.canNextEntryRoute(
					readerStore.readingEntry,
					(entry) => entry.unread
				)
			"
			@click="
				updateEntry(
					readerStore.getNextEntryId(
						readerStore.readingEntry,
						(entry) => entry.unread
					)
				)
			"
		>
			<bt-tooltip
				:label="t('base.next_unread')"
				:hotkey="WISE_HOTKEY.ENTRY.NEXT_UNREAD"
			/>
		</q-btn>
		<q-btn
			v-if="configStore.menuChoice.type !== MenuType.Trend"
			class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
			color="ink-2"
			outline
			no-caps
			:icon="
				readerStore.unread
					? 'sym_r_playlist_add_check'
					: 'sym_r_playlist_remove'
			"
			@click="readerStore.setCurrentReadChange(!readerStore.unread)"
		>
			<bt-tooltip
				:label="
					readerStore.unread ? t('main.mask_as_seen') : t('main.mask_as_unseen')
				"
				:hotkey="
					readerStore.unread ? WISE_HOTKEY.ENTRY.READ : WISE_HOTKEY.ENTRY.UNREAD
				"
			/>
		</q-btn>
		<q-btn
			class="my-custom-button q-mr-sm btn-size-sm btn-no-text btn-no-border"
			color="ink-2"
			outline
			no-caps
			:loading="readerStore.inboxLoading"
			icon="sym_r_inbox"
			:disable="readerStore.inbox"
			@click="setReadLater(false)"
		>
			<bt-tooltip
				:label="t('main.inbox')"
				:hotkey="WISE_HOTKEY.ENTRY.ADD_INBOX"
			/>
			<template v-slot:loading>
				<bt-loading :loading="readerStore.inboxLoading" />
			</template>
		</q-btn>
		<q-btn
			class="my-custom-button q-mr-sm btn-size-sm btn-no-text btn-no-border"
			color="ink-2"
			outline
			no-caps
			:loading="readerStore.readLaterLoading"
			icon="sym_r_schedule"
			:disable="readerStore.readLater"
			@click="setReadLater(true)"
		>
			<bt-tooltip
				:label="t('main.read_later')"
				:hotkey="WISE_HOTKEY.ENTRY.ADD_LATER"
			/>
			<template v-slot:loading>
				<bt-loading :loading="readerStore.readLaterLoading" />
			</template>
		</q-btn>
		<q-btn
			v-if="!isPDF"
			class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
			color="ink-2"
			outline
			no-caps
			icon="sym_r_open_in_browser"
			@click="onShare"
		>
			<bt-tooltip
				:label="t('main.open_origin')"
				:hotkey="WISE_HOTKEY.ENTRY.ORIGIN"
			/>
		</q-btn>
		<q-btn
			v-if="!isPDF"
			class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
			color="ink-2"
			outline
			:disable="!readerStore.readingFeed"
			:loading="subscribeLoading"
			no-caps
			:icon="
				readerStore.subscribed ? 'sym_r_bookmark_added' : 'sym_r_bookmark_add'
			"
			@click="setSubscribe"
		>
			<bt-tooltip
				:label="
					readerStore.subscribed
						? t('base.unsubscribe')
						: t('base.subscribe_this_feed')
				"
				:hotkey="WISE_HOTKEY.ENTRY.SUBSCRIBE"
			/>
			<template v-slot:loading>
				<bt-loading :loading="subscribeLoading" />
			</template>
		</q-btn>
		<!--		<q-btn-->
		<!--			v-if="-->
		<!--				configStore.menuChoice.type === MenuType.Trend &&-->
		<!--				readerStore.readingEntry &&-->
		<!--				readerStore.readingEntry.extra &&-->
		<!--				readerStore.readingEntry.extra.reason_data &&-->
		<!--				readerStore.readingEntry.extra.reason_data.length > 0-->
		<!--			"-->
		<!--			class="q-mr-sm btn-size-sm btn-no-text btn-no-border"-->
		<!--			color="ink-2"-->
		<!--			outline-->
		<!--			no-caps-->
		<!--			icon="sym_r_reviews"-->
		<!--			@click="showReasonDialog"-->
		<!--		>-->
		<!--			<bt-tooltip :label="t('base.recommend_reason')" />-->
		<!--		</q-btn>-->
		<!--		<q-btn-->
		<!--			class="q-mr-sm btn-size-sm btn-no-text btn-no-border"-->
		<!--			color="ink-2"-->
		<!--			outline-->
		<!--			no-caps-->
		<!--			v-if="showDebugButton"-->
		<!--			icon="sym_r_settings_alert"-->
		<!--			@click="showDebugDialog"-->
		<!--		>-->
		<!--			<bt-tooltip :label="t('base.debug_info')" />-->
		<!--		</q-btn>-->
	</title-right-layout>
</template>

<script setup lang="ts">
// import RecommendReasonDialog from '../../../../components/rss/dialog/RecommendReasonDialog.vue';
// import RecommendDebugDialog from '../../../../components/rss/dialog/RecommendDebugDialog.vue';
import BaseCheckBoxDialog from '../../../../components/base/BaseCheckBoxDialog.vue';
import TitleRightLayout from '../../../../components/base/TitleRightLayout.vue';
import BtTooltip from '../../../../components/base/BtTooltip.vue';
import BtLoading from '../../../../components/base/BtLoading.vue';
import { notifyFailed } from '../../../../utils/settings/btNotify';
import hotkeyManager from '../../../../directives/hotkeyManager';
import { useAbilityStore } from '../../../../stores/rss-ability';
import { WISE_HOTKEY } from '../../../../directives/wiseHotkey';
import { useConfigStore } from '../../../../stores/rss-config';
import { useReaderStore } from '../../../../stores/rss-reader';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { computed, onBeforeUnmount, onMounted } from 'vue';
import { FILE_TYPE } from '../../../../utils/rss-types';
import { MenuType } from '../../../../utils/rss-menu';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { useRouter } from 'vue-router';

const { t } = useI18n();
const $q = useQuasar();
const router = useRouter();
const configStore = useConfigStore();
const readerStore = useReaderStore();
const abilityStore = useAbilityStore();
let hotkeyMap;

// const showReasonDialog = () => {
// 	if (
// 		!readerStore.readingEntry ||
// 		!readerStore.readingEntry.extra ||
// 		!readerStore.readingEntry.extra.reason_data
// 	) {
// 		return;
// 	}
// 	$q.dialog({
// 		component: RecommendReasonDialog,
// 		componentProps: {
// 			extra: readerStore.readingEntry.extra
// 		}
// 	});
// };

// const showDebugButton = computed(() => {
// 	return (
// 		configStore.menuChoice.type === MenuType.Trend &&
// 		readerStore.readingEntry &&
// 		readerStore.readingEntry.debug_recommend_info
// 	);
// });
//
// const showDebugDialog = () => {
// 	if (!readerStore.readingEntry.debug_recommend_info) {
// 		return;
// 	}
// 	$q.dialog({
// 		component: RecommendDebugDialog,
// 		componentProps: {
// 			data: readerStore.readingEntry.debug_recommend_info
// 		}
// 	});
// };

const subscribeLoading = computed(() => {
	if (readerStore.readingFeed) {
		return readerStore.subscribedLoadingSet.has(readerStore.readingFeed.id);
	} else {
		return false;
	}
});

const isPDF = computed(() => {
	return (
		readerStore.readingEntry &&
		readerStore.readingEntry.file_type === FILE_TYPE.PDF
	);
});

const onShare = () => {
	if (readerStore.readingEntry && readerStore.readingEntry.url) {
		window.open(readerStore.readingEntry.url);
	} else {
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: t('base.failed_to_open_link')
		});
	}
};

const setSubscribe = async () => {
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
				readerStore.setCurrentSubscribe(selected).then(() => {
					router.back();
				});
			})
			.onCancel(() => {
				console.log('remove task cancel');
			});
	} else {
		readerStore.setCurrentSubscribe();
	}
};

const setReadLater = (readLater: boolean) => {
	readerStore.setCurrentReadLater(readLater).then(() => {
		router.back();
	});
};

const updateEntry = async (entryId: string, routerAction = '') => {
	if (entryId) {
		router.replace({
			name: 'entry',
			params: {
				id: entryId,
				action: routerAction
			}
		});
	}
};

onMounted(() => {
	hotkeyMap = {
		[WISE_HOTKEY.ENTRY.PRE]: () => {
			if (readerStore.canPreEntryRoute(readerStore.readingEntry)) {
				updateEntry(
					readerStore.getPreEntryId(readerStore.readingEntry),
					'back'
				);
			}
		},
		[WISE_HOTKEY.ENTRY.NEXT]: () => {
			if (readerStore.canNextEntryRoute(readerStore.readingEntry)) {
				updateEntry(readerStore.getNextEntryId(readerStore.readingEntry));
			}
		},
		[WISE_HOTKEY.ENTRY.NEXT_UNREAD]: () => {
			if (configStore.menuChoice.type === MenuType.Trend) {
				return;
			}
			if (
				readerStore.canNextEntryRoute(
					readerStore.readingEntry,
					(entry) => entry.unread
				)
			) {
				updateEntry(
					readerStore.getNextEntryId(
						readerStore.readingEntry,
						(entry) => entry.unread
					)
				);
			}
		},
		[WISE_HOTKEY.ENTRY.READ]: () => {
			if (configStore.menuChoice.type === MenuType.Trend) {
				return;
			}
			readerStore.setCurrentReadChange(false);
		},
		[WISE_HOTKEY.ENTRY.UNREAD]: () => {
			if (configStore.menuChoice.type === MenuType.Trend) {
				return;
			}
			readerStore.setCurrentReadChange(true);
		},
		[WISE_HOTKEY.ENTRY.ADD_INBOX]: () => {
			if (readerStore.inboxLoading) {
				return;
			}
			if (readerStore.inbox) {
				return;
			}
			setReadLater(false);
		},
		[WISE_HOTKEY.ENTRY.ADD_LATER]: () => {
			if (readerStore.readLaterLoading) {
				return;
			}
			if (readerStore.readLater) {
				return;
			}
			setReadLater(true);
		},
		[WISE_HOTKEY.ENTRY.ORIGIN]: () => {
			if (isPDF.value) {
				return;
			}
			onShare();
		},
		[WISE_HOTKEY.ENTRY.SUBSCRIBE]: () => {
			if (isPDF.value) {
				return;
			}

			if (subscribeLoading.value) {
				return;
			}

			if (!readerStore.readingFeed) {
				return;
			}
			setSubscribe();
		}
	};
	hotkeyManager.registerHotkeys(hotkeyMap);
});

onBeforeUnmount(() => {
	if (hotkeyMap) {
		hotkeyManager.unregisterHotkeys(hotkeyMap);
	}
});
</script>

<style scoped lang="scss"></style>
