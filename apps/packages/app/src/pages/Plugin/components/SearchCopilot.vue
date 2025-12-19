<template>
	<div class="search-copilot-container">
		<SimpleCard :avatar="stacks" title="Sources">
			<div class="row q-col-gutter-md">
				<div v-for="(item, index) in list" :key="index" class="col-6">
					<SimpleInfo :data="item"></SimpleInfo>
				</div>
			</div>
		</SimpleCard>

		<SimpleCard
			:avatar="speaker_notes"
			title="Answer"
			class="answer gutter-top"
		>
			<div class="text-body2 text-grey-10">
				With so much content to display at once, and often so little screen
				real-estate, Cards have fast become the design pattern of choice for
				many companies, including the likes of Google and Twitter.
			</div>
		</SimpleCard>
		<OthreOperation class="gutter-top"></OthreOperation>
		<q-card class="more-search gutter-top" flat bordered>
			<div class="more-search-content">
				<div class="text-subtitle2 text-grey-10">
					Use Search Agent for more comprehensive answers
				</div>
				<div class="text-body3 text-light-blue-6 more-btn">
					<span>Body3</span>
					<q-icon name="arrow_right_alt" />
				</div>
			</div>
		</q-card>
	</div>
</template>

<script setup lang="ts">
import SimpleInfo from './SimpleInfo.vue';
import SimpleCard from './SimpleCard.vue';
import OthreOperation from './OthreOperation.vue';
import image69 from 'assets/images/image69.png';
import image70 from 'assets/images/image70.png';
import image71 from 'assets/images/image71.png';
import image72 from 'assets/images/image72.png';
import image73 from 'assets/images/image73.png';
import speaker_notes from 'assets/images/speaker_notes.svg';
import stacks from 'assets/images/stacks.svg';

const list = [
	{
		title: 'BBC',
		text: 'BBC Home - Breaking News, World News, U.S...',
		img: image69,
		bordered: true
	},
	{
		title: 'YouTube',
		text: 'BBC News',
		img: image70
	},
	{
		title: 'You Tube',
		text: 'BBC News',
		img: image70
	},
	{
		title: 'Wikipedia',
		text: 'BBC',
		img: image71
	},
	{
		title: 'X',
		text: 'BBC Breaking News (@ BBCBreaking)/X',
		img: image72
	},
	{
		title: 'Facebook',
		text: 'BBC News | London',
		img: image73
	}
];
</script>

<style lang="scss" scoped>
.search-copilot-container {
	.gutter-top {
		margin-top: 20px;
	}

	.more-search {
		border: 1px solid $separator;
		border-radius: 12px;
		.more-search-content {
			padding: 16px;
			background-image: linear-gradient(
				128deg,
				#fffef7 4.22%,
				rgba(249, 254, 199, 0.5) 47.23%,
				rgba(243, 254, 194, 0.5) 81.06%
			);

			.more-btn {
				margin-top: 6px;
				cursor: pointer;
			}
		}
	}
}
</style>
