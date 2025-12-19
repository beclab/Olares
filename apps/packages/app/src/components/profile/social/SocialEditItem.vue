<template>
	<div
		:style="{ '--border': isEdit ? border : '' }"
		class="social-edit column justify-start"
	>
		<div class="social-edit-row row justify-between items-center">
			<div class="social-edit-board row items-center">
				<q-icon color="ink-2" size="20px" name="sym_r_drag_indicator" />
				<social-svg
					:platform="platform"
					width="24px"
					height="24px"
					class="q-ml-md"
				/>
				<div class="edit-area row justify-start items-center q-px-sm">
					<span class="text-body1 text-ink-1">@</span>
					<q-input
						borderless
						type="text"
						class="username-input"
						@blur="isEdit = false"
						@focus="isEdit = true"
						input-class="platform-name"
						:model-value="userName"
						@keyup.enter="onUserNameChanged(userName)"
						@update:model-value="onUserNameChanged"
					/>
				</div>
			</div>
			<q-icon
				color="ink-3 q-ma-sm cursor-pointer"
				size="24px"
				name="sym_r_delete"
				@click="onDeleteClick"
			/>
		</div>
		<div
			class="display-url q-mt-xs q-mr-xl text-ink-3 text-body3 cursor-pointer"
			@click="socialOpen"
		>
			{{ displayUrl }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import SocialSvg from '@apps/profile/src/components/profile/social/SocialSvg.vue';

const props = defineProps({
	platform: {
		type: String,
		required: true
	},
	userName: {
		type: String,
		required: true
	},
	url: {
		type: String,
		required: true
	}
});
const { color: border } = useColor('input-stroke');
const emit = defineEmits(['update:userName']);
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { useColor } from '@bytetrade/ui';

const userStore = useUserStore();
const isEdit = ref(false);
const displayUrl = ref();

const onDeleteClick = () => {
	if (userStore.user) {
		userStore.user.social.data = userStore.user.social.data.filter(
			(item) => item.platform !== props.platform
		);
	}
};

const onUserNameChanged = (value: string) => {
	formatUrl(value);
	emit('update:userName', value);
};

const formatUrl = (value: string) => {
	if (props.url) {
		try {
			if (value) {
				const itemUrl = props.url;
				displayUrl.value = itemUrl.replace('${username}', value);
			} else {
				displayUrl.value = props.url.replace('${username}', '');
			}
		} catch (e) {
			displayUrl.value = props.url;
		}
	} else {
		displayUrl.value = value;
	}
	console.log(displayUrl.value);
};

const socialOpen = () => {
	if (displayUrl.value) {
		window.open(displayUrl.value);
	}
};

onMounted(() => {
	formatUrl(props.userName);
});
</script>

<style scoped lang="scss">
.social-edit {
	width: 100%;
	height: auto;

	.social-edit-row {
		width: 100%;
		height: 40px;

		.social-edit-board {
			width: calc(100% - 40px);
			height: 40px;
			border-radius: 8px;
			border: 1px solid $input-stroke;
			background: $background-1;
			padding: 0 8px;

			.edit-area {
				height: 100%;
				width: calc(100% - 56px);

				.username-input {
					height: 100%;
					margin-left: 3px;
					width: calc(100% - 20px);
				}
			}
		}
	}

	.display-url {
		text-decoration: none;
	}
}
</style>
