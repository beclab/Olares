<template>
	<q-list class="my-routebackbar-container" @click="routeBack">
		<q-item>
			<q-item-section avatar class="test">
				<img :src="arrowLeftIcon" alt="" />
			</q-item-section>
			<q-item-section>
				<q-item-label lines="1">{{ title }}</q-item-label>
				<q-item-label caption v-if="subTitle">{{ subTitle }}</q-item-label>
			</q-item-section>
		</q-item>
	</q-list>
</template>

<script setup lang="ts">
import arrowLeftIcon from '@apps/control-hub/assets/arrow-left.svg';
import { toRefs } from 'vue';
import { useRoute, useRouter } from 'vue-router';

interface Props {
	title?: string;
	subTitle?: string;
}
const props = withDefaults(defineProps<Props>(), {});
const { title, subTitle } = toRefs(props);

const router = useRouter();

const routeBack = () => {
	router.back();
};
</script>

<style lang="scss" scoped>
.my-routebackbar-container {
	padding-left: 24px;
	font-size: 12px;
	line-height: 16px;
	color: #414141;
	background: #ffffff;
	cursor: default;
	.my-icon {
		color: #a9a9a9;
	}
	::v-deep(.test) {
		margin: 0;
		padding-right: 0;
		min-width: 16px;
		margin-right: 20px;
	}
}
</style>
