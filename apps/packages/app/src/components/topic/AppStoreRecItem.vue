<template>
	<div
		class="modal_application_single row justify-between items-center"
		@click="clickAppDetails(item)"
	>
		<div class="modal_application_top">
			<img :src="item ? item.icon : ''" class="application_top_logo" />
		</div>
		<div class="modal_application_text">
			<p class="modal_application_title">
				{{ item ? getAppFieldI18n(item, APP_FIELD.TITLE) : '' }}
			</p>
			<p class="modal_application_content">
				{{ item ? getAppFieldI18n(item, APP_FIELD.DESCRIPTION) : '' }}
			</p>
		</div>
		<div class="modal_application_install">
			<div
				v-if="item ? item.status === 'uninstalled' : ''"
				style="margin-left: auto"
			>
				<q-btn
					dense
					class="application_install_btn"
					outline
					no-caps
					label="Install"
					@click="appStore.installApp(item, false, $q)"
				/>
			</div>
			<div
				v-else-if="item ? item.status === 'installed' : ''"
				style="margin-left: auto"
			>
				<q-btn
					dense
					class="application_install_btn"
					outline
					no-caps
					label="Open"
				/>
			</div>
			<div v-else />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { useAppStore } from '@apps/market/src/stores/market/app';
import {
	APP_FIELD,
	AppStoreInfo,
	getAppFieldI18n
} from '@apps/market/src/constant/constants';
import { useQuasar } from 'quasar';

defineProps({
	item: {
		type: Object as PropType<AppStoreInfo>,
		required: true
	}
});
const emit = defineEmits(['clickAppDetails']);
const appStore: any = useAppStore();
const $q = useQuasar();

const clickAppDetails = (item: AppStoreInfo) => {
	emit('clickAppDetails', item);
};
</script>
<style lang="scss" scoped>
.modal_application_single {
	width: 100%;
	height: 114px;
	padding: 20px;
	background: rgba(0, 0, 0, 0.3);
	overflow: hidden;
	cursor: pointer;
	border-radius: 8px;
	backdrop-filter: blur(10px);
	.modal_application_top {
		width: 54px;
		height: auto;
		.application_top_logo {
			width: 54px;
			height: 54px;
			filter: drop-shadow(0px 0px 6px #d3d9ea);
			border-radius: 9px;
			overflow: hidden;
		}
	}
	.modal_application_text {
		width: calc(100% - 160px);
		flex-wrap: wrap;
		.modal_application_title {
			width: 100%;
			font-family: 'Source Han Sans CN';
			font-style: normal;
			font-weight: 700;
			font-size: 14px;
			line-height: 14px;
			color: #ffffff;
			margin-bottom: 0px;
		}
		.modal_application_content {
			width: 100%;
			height: 40px;
			margin-top: 7px;
			max-height: 34px;
			font-family: 'Source Han Sans CN';
			font-style: normal;
			font-weight: 400;
			font-size: 12px;
			color: #ffffff;
			margin-bottom: 0px;
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			line-height: 17px;
			-webkit-line-clamp: 2;
			-webkit-box-orient: vertical;
		}
	}
	.modal_application_install {
		width: 69px;
		height: 22px;

		.application_install_btn {
			width: 69px;
			height: 22px;
			background: #f1f4fd !important;
			border-radius: 10px;
			font-family: 'Source Han Sans CN';
			line-height: 12px;
			font-style: normal;
			font-weight: 400;
			font-size: 12px;
			color: #5568e8;
		}
	}
}
</style>
