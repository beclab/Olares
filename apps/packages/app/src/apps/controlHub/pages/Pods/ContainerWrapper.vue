<template>
	<MyCard no-content-gap flat :title="t('CONTAINERS')">
		<Containers></Containers>
	</MyCard>
</template>

<script setup lang="ts">
import Containers from '@apps/control-panel-common/src/containers/ContainersList/Containers.vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import { t } from 'src/boot/control-hub-i18n';
</script>

<style></style>
