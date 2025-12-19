<template>
	<div class="q-gutter-y-md">
		<div class="row items-center justify-between">
			<div class="text-ink-2">{{ $t('cookie_action.upload_olares') }}</div>
			<div class="row items-center">
				<div>
					<span class="text-ink-2">{{ $t('cookie_auto_sync_label') }}</span>
					<QToggleStyle style="margin: 0 2px 0 -6px">
						<q-toggle
							size="35px"
							v-model="auto_sync"
							color="yellow-default"
							:disable="!hasCookie"
							@update:model-value="autoSyncHandler"
						/>
					</QToggleStyle>
				</div>
				<QBtnStyle size="md">
					<q-btn
						dense
						outline
						no-caps
						color="ink-2"
						:disable="!hasCookie || !appAbilitiesStore.wise.running"
						:label="$t('cookie_upload_label')"
						@click="pushCookie"
					/>
				</QBtnStyle>
			</div>
		</div>
		<message-item :data="message.no_data" v-if="!hasCookie"></message-item>
		<div
			v-else
			class="message-container q-gutter-y-sm q-pa-md"
			:class="[
				current_message.type === 'info' ? 'bg-blue-soft' : 'bg-red-soft',
				`text-${current_message.type}`
			]"
			v-show="current_message.show"
		>
			<div class="row items-center q-gutter-x-sm">
				<q-icon name="sym_r_error" size="20px" />
				<span class="text-body3">{{ current_message.title }}</span>
			</div>
			<div class="text-body3">
				{{ current_message.content }}
			</div>
			<div class="text-body3 message-footer" @click="messageHandler">
				{{ current_message.footer }}
			</div>
		</div>
		<div v-for="item in cookiesList" :key="item.domain">
			<q-list class="rounded-borders" color="text-ink-2">
				<q-expansion-item
					:label="`${item.domain} (${item.records.length})`"
					:default-opened="item.domain === domain"
					class="cookie-expansion-item"
					header-class="text-ink-2"
					expand-icon-class="text-ink-3"
					dense
					dense-toggle
				>
					<div class="q-py-md">
						<chip-list v-model="item.records"></chip-list>
					</div>
				</q-expansion-item>
			</q-list>
			<div class="text-body3 text-ink-3 q-px-xs">{{ uploadTime }}</div>
		</div>
	</div>
	<q-inner-loading :showing="loading"> </q-inner-loading>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import ChipList from 'src/components/ChipList.vue';
import MessageItem from './MessageItem.vue';
import QToggleStyle from 'src/components/style/QToggleStyle.vue';
import QBtnStyle from 'src/components/style/QBtnStyle.vue';
import { useCookieContent } from 'src/composables/mobile/useCookieContent';
import { useAppAbilitiesStore } from '../../../stores/appAbilities';
const appAbilitiesStore = useAppAbilitiesStore();

const {
	message,
	cookiesList,
	domain,
	loading,
	auto_sync,
	uploadTime,
	current_message,
	hasCookie,
	messageHandler,
	getAllCookies,
	pushCookie,
	autoSyncHandler,
	initLocalMessage
} = useCookieContent();

onMounted(() => {
	initLocalMessage();
	getAllCookies();
});
</script>

<style lang="scss" scoped>
.message-container {
	border-radius: 12px;
}
.message-footer {
	text-decoration-line: underline;
	cursor: pointer;
}
::v-deep(.cookie-expansion-item .q-item) {
	padding-left: 4px;
	padding-right: 4px;
}
</style>
