<template>
	<MyPage>
		<Metadata v-if="metadata" :data="metadata"></Metadata>
	</MyPage>
</template>

<script setup lang="ts">
import Metadata from '@apps/control-panel-common/src/containers/Metadata.vue';
import { UsePod } from '@apps/control-panel-common/src/stores/PodData';
import { toRefs, computed } from 'vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';
import { get } from 'lodash';

const usePod = UsePod();

const metadata = computed(() => get(usePod.data, 'metadata'));
</script>

<style></style>
