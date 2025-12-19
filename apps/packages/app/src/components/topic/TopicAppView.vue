<template>
	<div class="column justify-start items-start cursor-pointer">
		<q-img
			@click="handleImgClick"
			class="topic-item-img"
			:src="item.iconimg ? item.iconimg : '../appIntro.svg'"
			:alt="item.iconimg"
			ratio="1.6"
		>
			<template v-slot:loading>
				<q-skeleton class="topic-item-img" />
			</template>
		</q-img>
		<recommend-app-card
			:app-name="item?.apps[0]"
			layout="column"
			:source-id="settingStore.marketSourceId"
			ref="cardRef"
			:is-last-line="true"
		/>
	</div>
</template>

<script setup lang="ts">
import RecommendAppCard from '../appcard/RecommendAppCard.vue';
import { useSettingStore } from '../../stores/market/setting';
import { TopicInfo } from '../../constant/constants';
import { PropType, ref } from 'vue';

defineProps({
	item: {
		type: Object as PropType<TopicInfo>,
		require: false
	}
});
const cardRef = ref();
const settingStore = useSettingStore();

const handleImgClick = () => {
	cardRef.value.goAppDetails();
};
</script>

<style scoped lang="scss">
.topic-item-img {
	border-radius: 12px;
	width: 100%;
	height: 100%;
}
</style>
