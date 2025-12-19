<template>
	<div
		class="row items-center justify-between my-routebackbar-container"
		:style="{ height: !isContent ? '56px' : 'auto' }"
		:class="[!isContent ? 'q-px-md' : '']"
	>
		<div class="row items-center" @click="routeBack">
			<div
				class="row justify-center items-center my-icon"
				:class="[size === 'md' ? 'md-icon-container' : 'lg-icon-container']"
				v-if="!isContent"
			>
				<q-icon
					name="arrow_back_ios_new"
					:size="size === 'md' ? '20px' : '24px'"
					class="text-ink-1"
				/>
			</div>
			<div v-if="$slots.avatar" class="q-mr-sm">
				<slot name="avatar"></slot>
			</div>
			<MyAvatarImg
				v-else-if="avatar"
				:src="avatar"
				:style="{
					width: !isContent ? '32px' : '48px',
					height: !isContent ? '32px' : '48px'
				}"
				:class="!isContent ? 'q-mr-sm' : 'q-mr-lg'"
			></MyAvatarImg>
			<span
				class="text-ink-1"
				:class="[
					size === 'md' ? 'text-h6' : !isContent ? 'text-h5' : 'text-h3'
				]"
			>
				{{ title }}
			</span>
		</div>
		<div class="row">
			<slot name="extra"></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import { toRefs } from 'vue';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import MyAvatarImg from './MyAvatarImg.vue';

interface Props {
	title?: string;
	subTitle?: string;
	avatar?: string;
	isContent?: boolean;
	size?: 'md' | 'lg';
	titleClickBack?: boolean;
}
const props = withDefaults(defineProps<Props>(), {
	size: 'lg',
	titleClickBack: true
});
const { title, subTitle } = toRefs(props);

const router = useRouter();
const route = useRoute();

const routeBack = () => {
	if (props.titleClickBack) {
		goBack();
	}
};

const goBack = () => {
	if (route.meta.parentRouteName) {
		router.replace({
			name: route.meta.parentRouteName as string
		});
	} else {
		router.go(-1);
	}
};
</script>

<style lang="scss" scoped>
.my-routebackbar-container {
	background: $background-1;
	cursor: default;
	.my-icon {
		border-radius: 8px;
		&:hover {
			background-color: $btn-bg-hover;
		}
	}
	.md-icon-container {
		width: 32px;
		height: 32px;
	}
	.lg-icon-container {
		width: 32px;
		height: 32px;
	}
}
.title-img {
	box-shadow: 0px 2px 4px 0px rgba(0, 0, 0, 0.1);
}
</style>
