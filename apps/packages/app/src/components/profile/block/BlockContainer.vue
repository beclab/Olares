<template>
	<div class="block-container row justify-between items-center">
		<q-icon color="ink-3" size="20px" name="sym_r_drag_indicator" />
		<div class="container-right column">
			<div class="container-title row justify-between items-center">
				<span class="container-type">{{ titleRef }}</span>
				<q-icon
					color="ink-3"
					size="20px"
					name="sym_r_delete"
					@click="deleteLink"
				/>
			</div>
			<div class="container-slot">
				<slot></slot>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, PropType, ref } from 'vue';

import { BLOCK_TYPE } from '@apps/profile/src/types/User';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { useI18n } from 'vue-i18n';
const userStore = useUserStore();
const props = defineProps({
	type: Object as PropType<BLOCK_TYPE>,
	index: {
		type: Number,
		required: true
	}
});

const { t } = useI18n();
const titleRef = ref();

onMounted(() => {
	switch (props.type) {
		case BLOCK_TYPE.link:
			titleRef.value = t('blocks.link');
			break;
		case BLOCK_TYPE.text:
			titleRef.value = t('blocks.text');
			break;
		case BLOCK_TYPE.image:
			titleRef.value = t('blocks.image');
			break;
	}
});

const deleteLink = () => {
	// const data = {
	//   ...userData.value,
	//   links:
	// };
	//updateUserData(data as User);
	if (!userStore.user) {
		return;
	}

	userStore.user.links = userStore.user.links.filter(
		(link, index) => index !== props.index
	);
};
</script>

<style scoped lang="scss">
.block-container {
	width: 400px;
	height: auto;
	border-radius: 12px;
	border: 1px solid $separator-2;
	background: $background-1;
	padding: 12px;

	.container-right {
		width: calc(100% - 20px);
		height: 100%;
		padding-left: 12px;

		.container-title {
			height: 20px;
			width: 100%;

			.container-type {
				text-transform: capitalize;
				font-family: Roboto;
				font-size: 12px;
				font-weight: 500;
				line-height: 16px;
				letter-spacing: 0em;
				text-align: left;
				color: $ink-1;
			}
		}

		.container-slot {
			width: 100%;
			height: calc(100% - 32px);
			margin-top: 12px;
		}
	}
}
</style>
