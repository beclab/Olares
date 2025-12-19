<template>
	<div class="page-title-root">
		<div v-if="!showBack" class="title setting-major-text row items-center">
			{{ title }}
		</div>

		<div v-if="showBack" class="row items-center" style="height: 100%">
			<div
				class="back-btn row items-center justify-center"
				@click="backToPrePage"
			>
				<q-icon name="bi-chevron-left" size="16px" />
			</div>
			<div class="title title-margin-left setting-major-text">
				{{ backTitle }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';

withDefaults(
	defineProps<{
		title: string;
		backTitle: string;
		showBack: boolean;
	}>(),
	{
		title: '',
		showBack: false,
		backTitle: ''
	}
);

const router = useRouter();

const backToPrePage = () => {
	router.back();
};
</script>

<style scoped lang="scss">
.page-title-root {
	height: 64px;
	width: 100%;
}

.title {
	font-size: 14px;
	font-weight: 500;
	line-height: 64px;
	// height: 100%;
}

.title-margin-left {
	margin-left: 12px;
}

.back-btn {
	width: 32px;
	height: 32px;
	border-radius: 4px;
	border: 1px solid #e0e2e9;
}

.back-btn:hover {
	background-color: #f5f8fb;
}
</style>
