<template>
	<page-title-component :show-back="true" :title="t('Cookie Management')">
		<template v-slot:end v-if="searchCookie.length > 0">
			<div
				class="row justify-center items-center cursor-pointer"
				@click="addCookie"
			>
				<q-icon name="sym_r_add" color="ink-1" size="20px" />
			</div>
		</template>
	</page-title-component>
	<div v-if="cookieStore.getAggregatedDomainList.length > 0">
		<div class="row items-center q-px-lg q-pt-md">
			<q-input
				class="cookie-search-input"
				dense
				borderless
				placeholder="Type a domain name"
				v-model="searchDomain"
			/>
			<div
				class="q-ml-md cursor-pointer"
				style="padding: 6px"
				@click="onRefresh"
			>
				<q-icon name="sym_r_refresh" size="20px" />
			</div>
		</div>

		<bt-scroll-area
			class="full-width q-px-lg"
			style="height: calc(100vh - 56px - 42px - 12px)"
		>
			<bt-list v-for="item in searchCookie" :key="item.mainDomain">
				<div
					class="row items-center q-pa-lg cursor-pointer"
					@click="onCookieDetail(item.mainDomain)"
				>
					<div style="padding: 5px">
						<q-icon size="30px" name="sym_r_language" class="text-ink-1" />
					</div>
					<div
						class="row justify-between items-center q-ml-sm"
						style="width: calc(100% - 48px)"
					>
						<div class="column">
							<div class="row">
								<div class="text-subtitle2 text-ink-1">
									{{ item.mainDomain }}
								</div>
								<div
									v-if="
										item.subDomains.length > 0 &&
										item.subDomains.some((sub) => sub.hasExpiredRecords())
									"
									class="text-negative bg-red-soft q-px-sm q-ml-sm q-py-xs text-caption"
								>
									{{ t('Cookie Expired') }}
								</div>
							</div>
							<div class="text-body3 text-ink-3 q-mt-xs">
								{{
									item.subDomains.reduce(
										(sum, subDomain) => sum + subDomain.records.length,
										0
									)
								}}
								Cookies · {{ t('last update time') }}
								{{
									date.formatDate(
										item.subDomains.reduce((maxItem, currentItem) => {
											return (currentItem.updateTime || 0) >
												(maxItem?.updateTime || 0)
												? currentItem
												: maxItem;
										}, {} as any)?.updateTime,
										'YYYY.MM.DD HH:mm:ss'
									)
								}}
							</div>
						</div>

						<q-icon
							size="24px"
							class="text-ink-1"
							name="sym_r_keyboard_arrow_right"
						/>
					</div>
				</div>
			</bt-list>
			<div style="height: 20px" />
		</bt-scroll-area>
	</div>
	<app-menu-empty
		v-else
		:title="t('No cookies found')"
		:button-label="t('Import Cookie')"
		image="settings/imgs/root/cookie.svg"
		@on-button-click="addCookie"
	>
		<template v-slot:message>
			<div
				class="cookie-message text-body2 text-ink-3 text-center q-mt-sm"
				style="max-width: 408px"
				v-html="formattedMessage"
			/>
		</template>
	</app-menu-empty>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import MultiTextDialog from 'src/components/rss/dialog/MultiTextDialog.vue';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useCookieStore } from 'src/stores/settings/cookie';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { date } from 'quasar';

const { t } = useI18n();
const $q = useQuasar();
const router = useRouter();
const cookieStore = useCookieStore();
const searchDomain = ref('');

const searchCookie = computed(() => {
	return cookieStore.getAggregatedDomainList.filter(
		(cookie) => cookie.mainDomain.indexOf(searchDomain.value) > -1
	);
});

const formattedMessage = computed(() => {
	const linkHtml = `<a
    href="https://docs.olares.com/manual/larepass/manage-knowledge.html#collect-content-via-the-larepass-extension"
    target="_blank"
    rel="noopener noreferrer"
    class="link text-blue-default"
    style="text-decoration: underline"
  >${t('upload via LarePass Chrome Extension')}</a>`;

	return t('Please add manually or link', { link: linkHtml });
});

const addCookie = () => {
	$q.dialog({
		component: MultiTextDialog,
		componentProps: {
			title: 'dialog.upload_cookie',
			label: 'dialog.Add cookies',
			link: false
		}
	});
};

const onRefresh = () => {
	cookieStore.getAllCookies().catch((err) => {
		console.error(err);
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: err?.response?.data || t('failed')
		});
	});
};

const onCookieDetail = (domain: string) => {
	router.push({
		path: '/integration/cookie/' + domain
	});
};
</script>

<style scoped lang="scss">
.cookie-search-input {
	width: calc(100% - 44px);
	border: solid 1px $input-stroke;
	border-radius: 8px;
	padding: 0 12px;
}
</style>
