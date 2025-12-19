<template>
	<div class="uploadHeader row items-center justify-between">
		<div class="row items-center justify-center" v-if="processingCount">
			<q-icon
				class="text-ink-1 q-mr-sm"
				name="sym_r_deployed_code_history"
				size="20px"
			></q-icon>

			<span class="text-ink-1 text-subtitle3">
				{{
					processingCount > 1
						? t('files.panel_tasks_operating', {
								count: processingCount
						  })
						: t('files.panel_task_operating', {
								count: processingCount
						  })
				}}
			</span>
		</div>

		<div class="row items-center justify-center" v-else>
			<img
				class="uploadStatus q-ml-xs q-mr-sm"
				src="../../../assets/images/uploaded.png"
				alt=""
			/>
			<span class="text-ink-1 text-subtitle3">{{
				t('files.panel_operated')
			}}</span>
		</div>

		<span>
			<q-icon
				class="q-mr-md cursor-pointer text-ink-2"
				rounded
				:name="
					showUpload ? 'sym_r_keyboard_arrow_down' : 'sym_r_keyboard_arrow_up'
				"
				@click="toggle"
				style="font-size: 20px"
			></q-icon>
			<q-icon
				class="cursor-pointer text-ink-2"
				rounded
				name="sym_r_close"
				@click="closeUploadModal"
				style="font-size: 20px"
			></q-icon>
		</span>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

const props = defineProps({
	processingCount: {
		type: Number,
		required: false,
		default: 0
	},

	totalCount: {
		type: Number,
		required: false,
		default: 0
	},

	showUpload: {
		type: Boolean,
		required: false,
		default: true
	}
});

const emits = defineEmits(['closePanel', 'togglePanel']);

const { t } = useI18n();

const toggle = () => {
	emits('togglePanel');
};

const closeUploadModal = () => {
	emits('closePanel');
};
</script>

<style scoped lang="scss">
.uploadModal {
	width: 350px;
	position: fixed;
	right: 20px;
	bottom: 20px;
	border-radius: 12px;
	overflow: hidden;
	box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2);
	.uploadHeader {
		width: 100%;
		height: 48px;
		padding: 0 20px;

		div {
			color: $ink-1;
			font-weight: 700;

			.uploadStatus {
				width: 16px;
			}
		}
	}

	.uploadContent {
		transition: height 0.3s;
		box-sizing: border-box;
	}

	.heightFull {
		height: 180px;
		border-top: 1px solid $separator;
	}

	.heightZero {
		height: 0;
	}
}
</style>
