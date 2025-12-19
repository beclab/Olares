<template>
	<div class="text-grey-10 text-h3">
		Hi {{ userStore.user_name }} <br />
		Good {{ getGreetingTime() }}
	</div>
	<!-- <div class="search row items-center justify-start">
		<div class="search-icon row items-center justify-center q-mr-md">
			<img :src="getPluginImageUrl('search.svg')" />
		</div>
		<div>
			<div class="text-grey-10 text-h6">AI Search</div>
			<div class="text-ink-1 text-body3">Smarter and save your time</div>
		</div>
	</div> -->

	<!-- <div class="row items-center justify-between q-mt-md">
		<div class="upload column items-center justify-center">
			<img :src="getPluginImageUrl('upload.svg')" />
			<div class="text-grey-10 text-h6 q-mt-md q-mb-xs">Upload</div>
			<div class="text-ink-1 text-body3 text-center">
				Click to add here<br />
				to chat
			</div>
		</div>

		<div class="write-warp">
			<div class="write column items-start justify-end">
				<img class="write-bg" :src="getPluginImageUrl('writing-bg.svg')" />
				<img class="q-mb-xs" :src="getPluginImageUrl('write.svg')" />
				<div class="text-grey-10 text-h6">Writing Agent</div>
			</div>

			<div class="agent text-grey-10 text-subtitle2">Auto Agent</div>
		</div>
	</div> -->
</template>

<script lang="ts" setup>
import { useUserStore } from '../../stores/user';
// import { getPluginImageUrl } from '../../stores/bex-url';

const userStore = useUserStore();

function getGreetingTime() {
	var time = new Date().getHours();
	if (time < 12) {
		return 'morning';
	} else if (time < 18) {
		return 'afternoon';
	} else {
		return 'evening';
	}
}
</script>

<style lang="scss" scoped>
.search {
	width: 100%;
	border: 1px solid $separator;
	border-radius: 20px;
	padding: 20px;
	margin-top: 20px;
	.search-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		background: $yellow;
		img {
			width: 18px;
			height: 18px;
		}
	}
}

.upload {
	width: calc(50% - 6px);
	height: 162px;
	border: 1px solid $separator;
	border-radius: 12px;
	padding: 32px 0;
	background: linear-gradient(
		121.71deg,
		#fffef7 4.54%,
		rgba(223, 255, 208, 0.5) 51.95%,
		rgba(210, 255, 188, 0.5) 89.25%
	);
}

.write-warp {
	width: calc(50% - 6px);
	height: 162px;
	display: flex;
	flex-direction: column;
	aligin-items: center;
	justify-content: space-between;

	.write {
		width: 100%;
		height: 112px;
		border: 1px solid $separator;
		border-radius: 12px;
		padding: 18px 20px;
		position: relative;
		.write-bg {
			position: absolute;
			width: 100%;
			height: 100%;
			right: 0;
			top: 0;
		}
	}

	.agent {
		width: 104px;
		height: 36px;
		line-height: 34px;
		text-align: center;
		border: 1px solid $separator;
		border-radius: 20px;
	}
}
</style>
