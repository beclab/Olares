<template>
	<page-title-component
		:show-back="false"
		:title="t(`home_menus.${MENU_TYPE.Search.toLowerCase()}`)"
	>
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<bt-form-item
				:title="t('File Search')"
				@click="gotoPage('/search/file')"
				:chevron-right="true"
				:width-separator="false"
			/>
		</bt-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { MENU_TYPE } from 'src/constant';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

const { t } = useI18n();
const router = useRouter();

function gotoPage(path: string) {
	router.push({ path });
}
</script>

<style scoped lang="scss"></style>
