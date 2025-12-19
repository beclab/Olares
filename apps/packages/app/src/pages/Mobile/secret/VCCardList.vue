<template>
	<div class="vc-card-list">
		<back-title-bar
			:title="notificationRef ? t('selected_vc_sign_vp') : 'VC'"
		/>
		<terminus-scroll-area class="vc-card-list__scroll">
			<template v-slot:content>
				<div class="vc-card-list__layout">
					<v-c-card-container
						class="vc-card-list__item"
						v-for="item in cardItemsRef"
						:item="item"
						@click="showDetails(item)"
						:key="item.id"
					/>
				</div>
				<terminus-item
					@click="addNew"
					class="vc-card-list__item"
					:whole-picture-size="36"
					title-classes="text-ink-2"
					:icon-size="12"
					v-if="!notificationRef"
					image-path="secret/add.svg"
				>
					<template v-slot:title>
						{{ t('add_vc') }}
					</template>
				</terminus-item>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import BackTitleBar from '../../../components/common/BackTitleBar.vue';
import { onMounted, ref } from 'vue';
import { getVCCardItemList } from '../items/VCItemUtil';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';

import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import VCCardDialog from '../items/VCCardDialog.vue';
import VCCardContainer from '../items/VCCardContainer.vue';
import { getUiType } from '../../../utils/utils';
import { useApproval } from '../wallet/approval';
import { useI18n } from 'vue-i18n';
import { VCCardItem } from '../../../utils/vc';

const cardItemsRef = ref([] as any[]);
const router = useRouter();
const $q = useQuasar();
const { t } = useI18n();

onMounted(() => {
	cardItemsRef.value = getVCCardItemList();
});

const notificationRef = ref(getUiType().isNotification);

const addNew = () => {
	router.replace({
		path: '/addVC'
	});
};

const showDetails = (item: VCCardItem) => {
	if (notificationRef.value) {
		approveAction(item);
	} else {
		$q.dialog({
			component: VCCardDialog,
			componentProps: {
				item
			}
		});
	}
};

const approveAction = async (item: VCCardItem) => {
	const { resolveApproval } = useApproval(router);
	await resolveApproval({ vc: item.verifiable_credential });
};
</script>

<style scoped lang="scss">
.vc-card-list {
	height: 100%;
	width: 100%;

	&__scroll {
		height: calc(100% - 44px);
		width: 100%;
		padding-left: 24px;
		padding-right: 24px;
	}

	&__item {
		margin-top: 16px;
		margin-bottom: 16px;
	}

	&__layout {
		width: 100%;
		height: 100%;
	}
}
</style>
