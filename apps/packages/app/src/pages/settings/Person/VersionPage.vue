<template>
	<page-title-component :show-back="true" :title="t('version')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<bt-form-item
				:title="t('current_version')"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="false"
			>
				{{ adminStore.terminus.osVersion }}
			</bt-form-item>
		</bt-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useI18n } from 'vue-i18n';
import { useAdminStore } from 'src/stores/settings/admin';
const adminStore = useAdminStore();
const { t } = useI18n();
</script>

<style scoped lang="scss">
.list_section {
	min-height: 56px;
}

.upgradeNow {
	border: 1px solid $btn-stroke;
	padding: 8px 12px;
	border-radius: 8px;
	cursor: pointer;
	// background-color: $background-3;

	&:hover {
		background: $background-5;
	}
}

.loader {
	width: 1rem;
	height: 1rem;
	color: inherit;
	vertical-align: middle;
	border: 0.2em solid transparent;
	border-top-color: currentcolor;
	border-bottom-color: currentcolor;
	border-radius: 50%;
	position: relative;
	animation: 1s loader-30 linear infinite;
	display: inline-block;

	&:before,
	&:after {
		content: '';
		display: block;
		width: 0;
		height: 0;
		position: absolute;
		border: 0.2em solid transparent;
		border-bottom-color: currentcolor;
	}

	&:before {
		transform: rotate(135deg);
		right: -0.2em;
		top: -0.05em;
	}

	&:after {
		transform: rotate(-45deg);
		left: -0.2em;
		bottom: -0.05em;
	}
}

@keyframes loader-30 {
	0% {
		transform: rotate(0deg);
	}

	100% {
		transform: rotate(360deg);
	}
}
</style>
