<template>
	<div class="carousel_box" v-if="showImageFirst.length > 0">
		<q-carousel
			animated
			v-model="slide"
			:keep-alive="true"
			arrows
			swipeable
			transition-prev="slide-right"
			transition-next="slide-left"
			style="
				border-radius: 20px;
				background: #f7f6fb;
				width: 100%;
				height: auto;
			"
		>
			<q-carousel-slide
				v-if="showImageFirst.length > 0"
				:name="1"
				class="column no-wrap"
			>
				<div class="promote-slide">
					<template v-for="item in showImageFirst" :key="item">
						<q-img class="promote-img" :src="item">
							<template v-slot:loading>
								<q-skeleton class="promote-img" />
							</template>
						</q-img>
					</template>
				</div>
			</q-carousel-slide>
			<q-carousel-slide
				v-if="showImageSecond.length > 0"
				:name="2"
				class="column no-wrap"
			>
				<div class="promote-slide">
					<template v-for="item in showImageSecond" :key="item">
						<q-img class="promote-img" :src="item">
							<template v-slot:loading>
								<q-skeleton class="promote-img" />
							</template>
						</q-img>
					</template>
				</div>
			</q-carousel-slide>
		</q-carousel>
	</div>
</template>

<script lang="ts" setup>
import { onMounted, PropType, Ref, ref, watch } from 'vue';
import { AppStoreInfo } from '@apps/market/src/constant/constants';

const props = defineProps({
	item: {
		type: Object as PropType<AppStoreInfo>,
		required: true
	},
	development: {
		type: Boolean,
		required: false
	}
});
const slide = ref<number>(1);

const showImageFirst = ref<string[]>([]);
const showImageSecond = ref<string[]>([]);
const showLength = 3;

onMounted(() => {
	showImageFirst.value = [];
	showImageSecond.value = [];
	if (props.item.promoteImage) {
		setImageList(showImageFirst, 1);
		setImageList(showImageSecond, 2);
	}
});

watch(
	() => props.item,
	() => {
		showImageFirst.value = [];
		showImageSecond.value = [];
		if (props.item.promoteImage) {
			setImageList(showImageFirst, 1);
			setImageList(showImageSecond, 2);
		}
	}
);

function setImageList(ref: Ref<string[]>, imgName: number) {
	props.item.promoteImage.forEach((item, index) => {
		if (Math.floor(index / showLength) == imgName - 1) {
			ref.value.push(item);
			// console.log(ref.value)
		}
	});
}
</script>

<style lang="scss" scoped>
.carousel_box {
	margin-top: 11px;
	height: auto;
	width: 100%;

	.promote-slide {
		width: 100%;
		display: grid;
		align-items: center;
		justify-items: center;
		justify-content: center;
		grid-template-columns: repeat(3, 32%);
		column-gap: 20px;
		row-gap: 20px;
	}

	.promote-img {
		border-radius: 20px;
		width: 100%;
		height: 100%;
	}
}
</style>
