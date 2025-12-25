<template>
	<transition name="fade">
		<div
			v-if="showHeaderBar"
			class="row justify-between items-center application-details-bar q-px-md"
			:style="absolute ? 'position: absolute;z-index: 99999' : ''"
		>
			<div
				class="row justify-start no-wrap"
				style="max-width: calc(100% - 108px); width: calc(100% - 108px)"
			>
				<div class="row items-center" @click="clickReturn" style="padding: 6px">
					<q-icon
						class="cursor-pointer"
						name="sym_r_arrow_back_ios_new"
						size="20px"
						color="ink-1"
					/>
				</div>
				<q-img v-if="showIcon" class="application_bar_img" :src="appIcon">
					<template v-slot:loading>
						<q-skeleton width="32px" height="32px" />
					</template>
				</q-img>
				<div
					class="row justify-start items-end no-wrap"
					style="max-width: calc(100% - 70px)"
				>
					<div
						class="q-ml-sm text-ink-1 ellipsis"
						:class="deviceStore.isMobile ? 'text-h6' : 'text-h5'"
					>
						{{ appTitle }}
					</div>
					<div v-if="!deviceStore.isMobile" class="q-ml-sm text-h6 text-ink-1">
						{{ appVersion }}
					</div>
				</div>
			</div>
			<slot />
		</div>
	</transition>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import { useDeviceStore } from '../../stores/settings/device';

defineProps({
	appTitle: {
		type: String,
		required: true
	},
	appVersion: {
		type: String,
		required: true
	},
	appIcon: {
		type: String,
		required: true
	},
	showHeaderBar: {
		type: Boolean,
		default: false,
		required: false
	},
	showIcon: {
		type: Boolean,
		default: true
	},
	absolute: {
		type: Boolean,
		default: false
	}
});

const router = useRouter();
const deviceStore = useDeviceStore();
const clickReturn = () => {
	if (window.history && window.history.state && !window.history.state.back) {
		router.replace('/');
		return;
	}
	router.back();
};
</script>

<style scoped lang="scss">
.application-details-bar {
	height: 56px;
	width: 100%;
	background-color: $background-2;
	box-shadow: 0 2px 4px 0 #0000001a;

	.application_bar_img {
		width: 30px;
		height: 30px;
		border-radius: 5.3px;
		box-shadow: 0 2px 4px 0 #0000001a;
	}
}
</style>
