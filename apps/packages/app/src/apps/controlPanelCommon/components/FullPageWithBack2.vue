<template>
	<div
		class="fixed-full bg-background-1"
		:style="{ left: '0px', zIndex: 1000 }"
	>
		<RouteBackBar
			:title="!visible ? title : ''"
			:avatar="!visible ? avatar : ''"
			:class="[!visible ? 'top-show' : '']"
			style="z-index: 999; position: relative"
		>
			<template #avatar v-if="avatar">
				<MyAvatarImg
					:style="{ opacity: !visible ? 1 : 0 }"
					:src="avatar"
				></MyAvatarImg>
			</template>
			<template #extra>
				<slot name="extra" v-if="!visible"></slot>
			</template>
		</RouteBackBar>

		<bt-scroll-area style="height: calc(100vh - 56px)" @scroll="scrollHander">
			<MyLayoutContainer>
				<div
					class="q-pb-lg"
					:style="{
						opacity: visible ? 1 : 0
					}"
				>
					<RouteBackBar
						:title="title"
						:avatar="avatar"
						isContent
						:titleClickBack="false"
						class="q-py-md q-mt-md"
					>
						<template #extra>
							<slot name="extra"></slot>
						</template>
					</RouteBackBar>
				</div>
				<slot></slot>
			</MyLayoutContainer>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
import RouteBackBar from './RouteBackBar.vue';
import MyLayoutContainer from './MyLayoutContainer.vue';
import { ref, computed } from 'vue';
import MyAvatarImg from './MyAvatarImg.vue';
import { useQuasar } from 'quasar';
const $q = useQuasar();

interface Props {
	title: string;
	avatar?: string;
}
withDefaults(defineProps<Props>(), {});
const visible = ref(false);
const emits = defineEmits(['scroll']);

const scrollHander = (data: any) => {
	if (data.verticalPosition < 48) {
		visible.value = true;
	} else {
		visible.value = false;
	}
	emits('scroll', data);
};

const shadowColor = computed(() =>
	$q.dark.isActive
		? 'rgba(160, 160, 160, 0.1) 0px 2px 4px 0px'
		: 'rgba(0, 0, 0, 0.1) 0px 2px 4px 0px'
);
</script>

<style lang="scss" scoped>
.top-show {
	box-shadow: v-bind(shadowColor);
}
</style>
