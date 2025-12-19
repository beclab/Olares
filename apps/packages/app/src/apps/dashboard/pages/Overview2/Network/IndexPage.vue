<template>
	<FullPageWithBack :title="$t('NETWORK_DETAILS')">
		<div class="column no-wrap flex-gap-y-xl">
			<MyCard v-for="(item, index) in netStore.list" :key="item.iface">
				<div class="row items-center justify-between">
					<div class="row items-center flex-gap-lg">
						<q-img
							:src="netIcon"
							:ratio="1"
							spinner-size="0px"
							width="32px"
							height="32px"
							style="align-self: flex-start"
						/>
						<div class="">
							<div class="text-h6 text-ink-1">
								<span>{{
									$t('NET_OP.NETWORK_PORT', { name: index + 1 })
								}}</span>
								<span>&nbsp;({{ item.iface }})</span>
							</div>
							<template v-if="item.isHostIp">
								<div class="text-body3 text-ink-2 q-mt-sm">
									<span>{{ $t('NET_OP.USE_STATUS') }}:&nbsp;</span>
									<span class="text-green-default">{{
										$t('NET_OP.NET_USED')
									}}</span>
								</div>
							</template>
						</div>
					</div>
					<div class="row no-wrap flex-gap-x-xl">
						<div class="row text-subtitle2 text-ink-1">
							<div class="row items-center">
								<span>{{ networkRate(item.txRate) }}</span>
								<q-icon
									name="sym_r_arrow_upward_alt"
									color="positive"
									size="20px"
								/>
							</div>
							<q-separator class="q-mx-md" vertical color="separator-2" />
							<div class="row items-center">
								<span>{{ networkRate(item.rxRate) }}</span>
								<q-icon
									name="sym_r_arrow_downward_alt"
									color="negative"
									size="20px"
								/>
							</div>
						</div>
						<div
							class="row items-center flex-gap-md text-body3"
							:class="[connectFormat(item.internetConnected).color]"
						>
							<div
								class="dot"
								:class="[connectFormat(item.internetConnected).bg]"
							></div>
							<span>{{ connectFormat(item.internetConnected).label }}</span>
						</div>
					</div>
				</div>

				<template v-if="item.internetConnected">
					<q-separator class="q-mt-lg q-mb-xl" color="separator" />
					<Descriptions :data="item.contentList" colWidth="362px">
					</Descriptions>
					<Descriptions
						:data="item.contentList2"
						colWidth="362px"
						class="q-mt-xl"
					>
					</Descriptions>
				</template>
			</MyCard>
		</div>
		<q-inner-loading :showing="netStore.loading"> </q-inner-loading>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack2.vue';
import MyCard from '@apps/dashboard/components/MyCard.vue';
import { getThroughput } from '@apps/dashboard/src/utils/memory';
import netIcon from '@apps/dashboard/assets/net.svg';
import Descriptions from '@apps/control-panel-common/components/Descriptions.vue';
import { useI18n } from 'vue-i18n';
import { useNetStore } from '@apps/dashboard/stores/Net';
import { onBeforeUnmount } from 'vue';

const { t } = useI18n();
const netStore = useNetStore();

const connectFormat = (internetConnected) => {
	if (internetConnected) {
		return {
			label: t('NET_OP.IS_CONNECTED'),
			color: 'text-positive',
			bg: 'bg-positive'
		};
	} else {
		return {
			label: t('NET_OP.NOT_CONNECTED'),
			color: 'text-red',
			bg: 'bg-red'
		};
	}
};

const networkRate = (value) => {
	return getThroughput(value);
};

onBeforeUnmount(() => {
	netStore.clearLocker();
});
</script>

<style lang="scss" scoped>
.dot {
	width: 8px;
	height: 8px;
	border-radius: 50%;
	background: $status-pending;
}
.net-tab-wrapper {
	border-radius: 4px;
}
</style>
