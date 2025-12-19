<template>
	<div
		class="page-title-root row items-center"
		:style="{
			'--root-padding': showBack || deviceStore.isMobile ? '12px' : '20px'
		}"
	>
		<AdaptiveLayout>
			<template v-slot:pc>
				<div class="row justify-between items-center" style="height: 100%">
					<div class="row items-center" style="height: 100%">
						<div
							v-if="showBack"
							class="back-btn row items-center justify-center"
							@click="backToPrePage"
						>
							<q-icon
								name="sym_r_arrow_back_ios_new"
								size="20px"
								class="text-ink-2"
							/>
						</div>
						<div class="text-h5 text-ink-1">
							{{ title }}
						</div>
					</div>
					<slot name="end" />
				</div>
			</template>
			<template v-slot:mobile>
				<div class="row items-center justify-between" style="height: 100%">
					<div
						class="back-btn row items-center justify-center"
						@click="backToPrePage"
					>
						<q-icon name="sym_r_chevron_left" size="32px" />
					</div>
					<div
						class="mobile-title row items-center justify-center text-h6-m text-ink-1"
					>
						{{ title }}
					</div>
					<slot name="end" />
				</div>
			</template>
		</AdaptiveLayout>
	</div>
</template>

<script setup lang="ts">
import AdaptiveLayout from './AdaptiveLayout.vue';
import { useRouter } from 'vue-router';
import { useDeviceStore } from 'src/stores/settings/device';

const props = defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	showBack: {
		type: Boolean,
		required: false,
		default: false
	},
	customBack: {
		type: Boolean,
		required: false,
		default: false
	}
});

const emit = defineEmits(['onBackClick']);
const router = useRouter();
const deviceStore = useDeviceStore();

const backToPrePage = () => {
	if (props.customBack) {
		emit('onBackClick');
	} else {
		// router.back();
		// window.history.length
		console.log('window.history.length ===>', window.history.length);

		if (deviceStore.isMobile && window.history.length <= 1) {
			router.replace('/');
		} else {
			router.back();
		}
	}
};
</script>

<style scoped lang="scss">
.page-title-root {
	height: 56px;
	padding-left: var(--root-padding);
	padding-right: 16px;
	width: 100%;
	position: relative;

	.back-btn {
		width: 32px;
		height: 32px;
		border-radius: 8px;
		margin-right: 4px;

		&:hover {
			background-color: $background-3;
		}
	}
	.mobile-title {
		position: absolute;
		top: 16px;
		left: 50px;
		right: 50px;
		width: calc(100% - 100px);
		word-wrap: break-word;
		word-break: break-all;
		white-space: nowrap;
		overflow: hidden;
	}
}
</style>
