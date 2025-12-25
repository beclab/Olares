<template>
	<div
		class="studio-notify row items-start justify-between q-px-lg q-py-md"
		:style="{ background: app_status_style[status].notify_bg }"
	>
		<div class="notify-content row items-start justify-start">
			<q-icon
				class="q-mr-sm"
				name="sym_r_error"
				size="20px"
				:style="{ color: app_status_style[status].color }"
			/>
			<div
				ref="messageContent"
				class="text-ink-1 text-body3 content"
				:style="{
					WebkitLineClamp: isExpanded ? '1' : '4',
					display: '-webkit-box',
					overflow: 'hidden',
					textOverflow: 'ellipsis',
					WebkitBoxOrient: 'vertical'
				}"
			>
				{{ message }}
			</div>
		</div>

		<div class="q-ml-lg">
			<span
				v-if="needsExpand"
				class="text-blue-default text-body3 cursor-pointer"
				@click="toggleExpand"
				>{{ isExpanded ? t('header_btn.open') : t('header_btn.fold') }}</span
			>
			<q-icon
				class="q-ml-sm text-ink-2 cursor-pointer"
				name="sym_r_close_small"
				size="20px"
				@click="close"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, PropType, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { APP_STATUS, app_status_style } from '../../types/core';

const props = defineProps({
	status: {
		type: String as PropType<APP_STATUS>,
		required: true
	},
	message: {
		type: String,
		required: false,
		default: ''
	}
});

const emits = defineEmits(['close']);

const { t } = useI18n();

const isExpanded = ref(false);
const messageContent = ref<HTMLElement | null>(null);

const toggleExpand = () => {
	isExpanded.value = !isExpanded.value;
};

const needsExpand = ref(false);

const checkIfNeedsExpand = () => {
	if (!messageContent.value) return;

	const scrollHeight = messageContent.value.scrollHeight;

	if (scrollHeight > 22) {
		needsExpand.value = true;
	} else {
		needsExpand.value = false;
	}
};

const close = () => {
	emits('close');
};

watch(
	() => [isExpanded.value, props.message],
	() => {
		nextTick(() => {
			checkIfNeedsExpand();
		});
	},
	{
		immediate: true
	}
);
</script>

<style lang="scss" scoped>
.studio-notify {
	width: 100%;

	.notify-content {
		width: calc(100% - 100px);

		.content {
			flex: 1;
			line-height: 22px;
			transition: all 0.3s ease;
			word-break: break-all;
		}
	}
}
</style>
