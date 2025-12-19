<template>
	<div class="account-list-root">
		<terminus-title-bar :title="t('website_manager')" />
		<div v-if="connectWebSiteRef.length > 0" class="account-list">
			<terminus-page-title :desc="t('website_manager_desc')" />
			<q-scroll-area class="scroll-area-conf">
				<terminus-item
					v-for="(website, index) in connectWebSiteRef"
					:key="'ii' + index"
					@click="choose(website.origin)"
					style="margin-top: 16px"
					:image-uri="website.icon"
				>
					<template v-slot:title>{{ website.name }}</template>
					<template v-slot:detail>{{ website.origin }}</template>
				</terminus-item>
			</q-scroll-area>
		</div>
		<div v-else>
			<terminus-empty
				img-path="message/message_empty.svg"
				:info="t('no_information')"
				style="width: 100%; height: 100%"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useUserStore } from '../../../stores/user';
import { useI18n } from 'vue-i18n';
import { useBexStore } from '../../../stores/bex';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusPageTitle from '../../../components/common/TerminusPageTitle.vue';
import { onMounted, ref } from 'vue';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import TerminusDialog from '../../Items/dialog/TerminusDialog.vue';
import { useQuasar } from 'quasar';
import TerminusEmpty from '../../../components/common/TerminusEmpty.vue';

const userStore = useUserStore();
const { t } = useI18n();
const bexStore = useBexStore();
const connectWebSiteRef = ref([] as any[]);
const $q = useQuasar();
const didKey = userStore.current_id;

onMounted(async () => {
	if (didKey) {
		connectWebSiteRef.value = await bexStore.controller.getConnectedSite();
	}
});

const choose = async (origin: string) => {
	if (didKey) {
		$q.dialog({
			component: TerminusDialog,
			componentProps: {
				title: t('disconnect_warning'),
				content: t('disconnect_warning_content'),
				leftText: t('cancel'),
				rightText: t('confirm')
			}
		}).onOk(async () => {
			await bexStore.controller.removeConnectedSite(origin);
			connectWebSiteRef.value = await bexStore.controller.getConnectedSite();
		});
	}
};
</script>

<style lang="scss" scoped>
.account-list-root {
	width: 100%;
	height: 100%;

	.account-list {
		height: 100%;
		width: 100%;
		padding-top: 20px;
		padding-left: 24px;
		padding-right: 24px;

		.scroll-area-conf {
			height: calc(100% - 106px);
			width: 100%;
			padding-top: 20px;
		}
	}
}
</style>
