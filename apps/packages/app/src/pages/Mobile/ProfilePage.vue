<template>
	<div class="profile-root">
		<back-title-bar :title="t('profile')" />
		<!-- TODO: snowning.com error-->
		<iframe
			class="iframe-profile"
			:src="`https://${name}` + '.' + TerminusDefaultDomain"
			frameborder="no"
			scrolling="yes"
			:allowtransparency="true"
		/>
	</div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import BackTitleBar from '../../components/common/BackTitleBar.vue';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { TerminusDefaultDomain } from '@bytetrade/core';

const Route = useRoute();

const name = ref(Route.query.name ? (Route.query.name as string) : '');

const { t } = useI18n();
</script>

<style scoped lang="scss">
.profile-root {
	width: 100%;
	height: 100%;

	.iframe-profile {
		width: 100%;
		height: calc(100% - 44px);

		position: relative;
		right: 0;
		bottom: 0;
		left: 0;
		top: 0;
		-webkit-overflow-scrolling: touch;
		overflow-y: scroll;
	}
}
</style>
