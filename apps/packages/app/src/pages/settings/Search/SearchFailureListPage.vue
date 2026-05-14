<template>
	<page-title-component :show-back="true" :title="t('Failed File List')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<BtLoading
			:show="true"
			v-if="isLoading"
			textColor="#4999ff"
			color="#4999ff"
			text=""
			backgroundColor="rgba(0, 0, 0, 0)"
		>
		</BtLoading>
		<bt-list v-else-if="filedList.length > 0">
			<bt-form-item
				v-for="(item, index) in filedList"
				:key="item.resource_uri"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="index !== filedList.length - 1"
			>
				<template v-slot:all>
					<div class="column q-px-lg">
						<div class="text-body1 text-ink-1">{{ item.resource_uri }}</div>
						<div class="text-body3 text-negative">
							{{ item.extract_error_message }}
						</div>
					</div>
				</template>
			</bt-form-item>
		</bt-list>
		<app-menu-empty
			v-else
			:title="t('No failed file')"
			message=""
			image="settings/imgs/root/file.svg"
		/>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { getFailedPathList } from 'src/api/settings/search';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const filedList = ref([]);
const isLoading = ref(false);

onMounted(() => {
	isLoading.value = true;
	getFailedPathList()
		.then((pathList) => {
			filedList.value = pathList;
		})
		.catch((e) => {
			notifyFailed(e.response.data.message || e.message);
		})
		.finally(() => {
			isLoading.value = false;
		});
});
</script>

<style scoped lang="scss"></style>
