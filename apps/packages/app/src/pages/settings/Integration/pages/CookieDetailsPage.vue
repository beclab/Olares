<template>
	<page-title-component :show-back="true" :title="t('Cookie Management')">
		<template v-slot:end>
			<CustomButton
				:label="t('Delete All')"
				icon="sym_r_delete"
				class="q-mr-md"
				@click="deleteAggregatedCookie"
			/>
		</template>
	</page-title-component>
	<bt-scroll-area
		v-if="domainCookieList.length > 0"
		class="full-width q-px-lg"
		style="height: calc(100vh - 56px - 42px - 12px)"
	>
		<div v-for="item in domainCookieList" :key="item.domain">
			<bt-list color="text-ink-2">
				<custom-expansion-item>
					<template v-slot:header="{ isOpen }">
						<div class="row justify-between items-center full-width">
							<div class="column">
								<div class="row justify-start items-center">
									<div class="text-ink-1 text-subtitle2">
										{{ item.domain }}
									</div>
									<q-icon
										v-if="item.hasExpiredRecords()"
										class="q-ml-xs"
										name="sym_r_error"
										color="negative"
									/>
								</div>
								<div class="text-overline text-ink-3">
									{{ item.records.length }} cookies ·
									{{ t('last update time') }}
									{{ date.formatDate(item.updateTime, 'YYYY.MM.DD HH:mm:ss') }}
								</div>
							</div>
							<div class="row justify-end items-center">
								<div
									class="arrow q-mr-md"
									:style="{
										transform: isOpen ? 'rotate(180deg)' : 'rotate(0)'
									}"
								>
									<q-icon
										class="text-ink-2"
										size="24px"
										name="sym_r_keyboard_arrow_up"
									/>
								</div>
								<q-icon
									size="24px"
									name="sym_r_delete"
									class="cursor-pointer text-ink-2"
									@click.stop="deleteCookie(item.domain)"
								/>
							</div>
						</div>
					</template>
					<chip-list
						:model-value="item.records"
						@update:modelValue="
							(data) => {
								updateDomainCookieRecords(data, item);
							}
						"
					/>
				</custom-expansion-item>
			</bt-list>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import CustomExpansionItem from 'src/components/CustomExpansionItem.vue';
import CustomButton from 'src/components/settings/CustomButton.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import ChipList from 'src/components/ChipList.vue';
import { useCookieStore } from 'src/stores/settings/cookie';
import { BtDialog, BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { DomainCookie, DomainCookieRecord } from 'src/constant/constants';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { date } from 'quasar';

const domainCookieList = computed(() => {
	return cookieStore.getDomainCookies(route.params.mainDomain);
});
const cookieStore = useCookieStore();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const deleteAggregatedCookie = () => {
	BtDialog.show({
		title: t('Delete All'),
		message: t('Are you sure you want to delete all cookies for the domain', {
			domain: route.params.mainDomain
		}),
		okStyle: {
			background: 'yellow-default',
			color: '#1F1F1F'
		},
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel')
	})
		.then((res: any) => {
			if (res) {
				cookieStore
					.deleteAggregatedDomain(route.params.mainDomain)
					.then(() => {
						BtNotify.show({
							type: NotifyDefinedType.SUCCESS,
							message: t('success')
						});
						router.back();
					})
					.catch((err) => {
						console.error(err);
						BtNotify.show({
							type: NotifyDefinedType.FAILED,
							message: err?.response?.data || t('failed')
						});
					});
			}
		})
		.catch((err: Error) => {
			console.log('click cancel', err);
		});
};

const deleteCookie = (subDomain: string) => {
	BtDialog.show({
		title: t('Delete a single domain'),
		message: t('Are you sure you want to delete all cookies for the domain', {
			domain: subDomain
		}),
		okStyle: {
			background: 'yellow-default',
			color: '#1F1F1F'
		},
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel')
	})
		.then((res: any) => {
			if (res) {
				if (domainCookieList.value.length > 1) {
					cookieStore
						.deleteDomainCookie(route.params.mainDomain, subDomain)
						.then(() => {
							BtNotify.show({
								type: NotifyDefinedType.SUCCESS,
								message: t('success')
							});
						})
						.catch((err) => {
							console.error(err);
							BtNotify.show({
								type: NotifyDefinedType.FAILED,
								message: err?.response?.data || t('failed')
							});
						});
				} else {
					cookieStore
						.deleteAggregatedDomain(route.params.mainDomain)
						.then(() => {
							BtNotify.show({
								type: NotifyDefinedType.SUCCESS,
								message: t('success')
							});
							router.back();
						})
						.catch((err) => {
							console.error(err);
							BtNotify.show({
								type: NotifyDefinedType.FAILED,
								message: err?.response?.data || t('failed')
							});
						});
				}
			}
		})
		.catch((err: Error) => {
			console.log('click cancel', err);
		});
};

const updateDomainCookieRecords = (
	record: DomainCookieRecord,
	cookie: DomainCookie
) => {
	cookieStore
		.updateDomainCookie(cookie.domain, record)
		.then(() => {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('success')
			});
		})
		.catch((err) => {
			console.error(err);
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: err?.response?.data || t('failed')
			});
		});
};
</script>

<style scoped lang="scss">
.arrow {
	transition: transform 0.3s ease;
	font-size: 12px;
}
</style>
