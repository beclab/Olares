<template>
	<BindTerminusVCContent style="height: 100%" :has-btn="false">
		<template v-slot:desc>
			{{ t('domain_list_under_account_desc') }}
		</template>
		<template v-slot:content>
			<q-list
				class="q-mt-md"
				v-if="cloudStore.domains && cloudStore.domains.length > 0"
			>
				<q-item
					dense
					clickable
					class="row items-center justify-between domain-item"
					@click="onClick(domain)"
					v-for="(domain, index) in cloudStore.domains"
					:key="index"
					:class="index > 0 ? 'q-mt-md' : ''"
				>
					<div class="row items-center" style="width: calc(100% - 30px)">
						<div
							class="row items-center justify-center bg-yellow"
							style="border-radius: 32px; width: 32px; height: 32px"
						>
							<q-icon name="sym_r_home_work" size="20px" color="grey-10" />
						</div>
						<div class="column q-ml-sm" style="width: calc(100% - 40px)">
							<div
								class="domain-item__title text-subtitle2"
								style="word-break: break-all; width: 100%"
							>
								{{ domain.domain }}
							</div>
							<div
								class="domain-item__detail text-body3"
								style="word-break: break-all; width: 100%"
							>
								{{ getDomainDesc(domain.status) }}
							</div>
						</div>
					</div>
					<div class="row items-center justify-end">
						<q-icon name="sym_r_keyboard_arrow_right" size="20px" />
					</div>
				</q-item>
			</q-list>
			<div
				v-else
				class="column items-center justify-center empty"
				style="width: 100%; margin-top: 80px"
			>
				<img src="../../../../assets/nodata.svg" alt="empty" />
				<span class="text-body2">{{ t('there_is_nothing_for_now') }}</span>
				<div
					class="row items-center justify-center"
					style="margin-top: 40px"
					@click="refresh"
				>
					<q-icon name="sym_r_sync" size="20px" color="blue-4" />
					<div class="domain-list-refresh text-body2 q-ml-xs">
						{{ t('refresh') }}
					</div>
				</div>
			</div>
		</template>
	</BindTerminusVCContent>
</template>

<script setup lang="ts">
import BindTerminusVCContent from '../../login/vc/BindTerminusVCContent.vue';
import { onMounted } from 'vue';
import { Domain, useCloudStore, getDomainDesc } from '../../../../stores/cloud';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { Encoder } from '@bytetrade/core';
import { useQuasar } from 'quasar';
import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';

const cloudStore = useCloudStore();

const { t } = useI18n();

const router = useRouter();
const $q = useQuasar();

onMounted(async () => {
	await cloudStore.getDomains();
});

const refresh = async () => {
	try {
		$q.loading.show();
		await cloudStore.getDomains();
		$q.loading.hide();
	} catch (error) {
		$q.loading.hide();
		if (error && error.message) {
			notifyFailed(error.message);
		}
	}
};

async function onClick(domain: Domain) {
	router.push({
		path: '/DomainDetail/' + Encoder.stringToBase64Url(domain.domain)
	});
}
</script>

<style scoped lang="scss">
.domain-list {
	height: 24px;
}

.domain-list-refresh {
	text-align: right;
	color: $blue-4;
}

.domain-item {
	// height: 60px;
	width: 100%;
	border: 1px solid $separator;
	border-radius: 8px;
	padding: 12px 10px;

	&__title {
		color: $ink-1;
	}

	&__detail {
		text-align: left;
		color: $ink-2;
	}
}
.empty {
	width: 100%;
	height: calc(100% - 80px);

	img {
		width: 120px;
		height: 120px;
		margin-bottom: 40px;
	}

	span {
		color: $grey-14;
	}
}
</style>
