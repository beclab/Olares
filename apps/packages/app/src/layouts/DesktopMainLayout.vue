<template>
	<q-layout view="hHh lpR fFf">
		<q-page-container>
			<router-view />
		</q-page-container>

		<bt-theme />
	</q-layout>
</template>

<script lang="ts" setup>
import { onUnmounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { Result, FileSearchAIQuestionMessage } from '@bytetrade/core';
import { useAppStore } from '../stores/desktop/app';
import { useSearchStore } from 'src/stores/search';
import { useUpgradeStore } from '../stores/desktop/upgrade';
import { bus } from '..//utils/bus';

const router = useRouter();
const appStore = useAppStore();
const searchState = useSearchStore();
const upgradeStore = useUpgradeStore();
const timer = ref();
const showTab = ref(false);

const updateTab = () => {
	if (router.currentRoute.value.fullPath === '/') {
		showTab.value = true;
	} else {
		showTab.value = false;
	}
};

watch(
	() => router.currentRoute.value,
	() => {
		updateTab();
	}
);

updateTab();

bus.on('close', () => {
	clearInterval(timer.value);
});

bus.on('app_installation_event', () => {
	appStore.update_my_apps_info();
});

bus.on('entrance_state_event', () => {
	appStore.update_my_apps_info();
});

bus.on('system_upgrade_event', (data: any) => {
	upgradeStore.state = data.payload.status;
});

bus.on('ai', (data: any) => {
	if (data.code == 0) {
		searchState.conversationId = data!.conversationId;
		searchState.waiting = true;

		let text = '···';
		timer.value = setInterval(() => {
			let listIndex = searchState.chatList.findIndex(
				(item) => item.messageId === data.messageId
			);
			if (text === '···') {
				text = '·';
			} else {
				text = text + '·';
			}
			const tempObj = {
				text,
				messageId: data.messageId,
				conversationId: data.conversationId,
				model: data.model,
				name: 'Terminus'
			};
			if (listIndex > -1) {
				searchState.chatList[listIndex] = tempObj;
			} else {
				searchState.chatList.push(tempObj);
			}
		}, 500);
	} else {
		searchState.chatList.push({
			text: 'Sorry something went wrong. Please try again.',
			messageId: '',
			conversationId: data!.conversationId,
			model: 'chat_model',
			name: 'Terminus'
		});
		searchState.waiting = false;
		searchState.can_input = true;
		clearInterval(timer.value);
	}
});

bus.on('ai_message', (cdata: any) => {
	clearInterval(timer.value);
	searchState.waiting = false;

	const data: Result<FileSearchAIQuestionMessage> = cdata.data;
	if (data.code == 0) {
		if (data.data && !data.data.text) {
			data.data.text =
				'Ashia can only run on instance with gpu for the time being.';
		}

		let listIndex = searchState.chatList.findIndex(
			(item) => item.messageId === data.data!.messageId
		);
		if (listIndex > -1) {
			searchState.chatList[listIndex] = {
				text: data.data!.text,
				messageId: data.data!.messageId,
				conversationId: data.data!.conversationId,
				model: data.data!.model,
				name: 'Terminus'
			};
		} else {
			searchState.chatList.push({
				text: data.data!.text,
				messageId: data.data!.messageId,
				conversationId: data.data!.conversationId,
				model: data.data!.model,
				name: 'Terminus'
			});
		}

		if (data.data!.done) {
			searchState.can_input = true;
		}
	} else {
		searchState.chatList.push({
			text: 'Sorry something went wrong. Please try again.',
			messageId: '',
			conversationId: data.data!.conversationId,
			model: 'chat_model',
			name: 'Terminus'
		});

		searchState.can_input = true;
	}
});

onUnmounted(() => {
	bus.off('close');
	bus.off('app_installation_event');
	bus.off('entrance_state_event');
	bus.off('system_upgrade_event');
	bus.off('ai');
	bus.off('ai_message');
});
</script>
