<template>
	<adaptive-layout>
		<template v-slot:pc>
			<div
				class="header-bar row justify-start items-center"
				v-if="show"
				:style="[{ boxShadow: shadow ? '0px 2px 4px 0px #0000001A' : 'none' }]"
			>
				<div v-if="showBack" class="row items-center" style="padding: 6px">
					<q-icon
						class="cursor-pointer"
						name="sym_r_arrow_back_ios_new"
						size="20px"
						color="ink-1"
						@click="onReturn"
					/>
				</div>
				<transition name="fade">
					<div v-if="showTitle" class="header-title text-subtitle2 text-ink-1">
						{{ title }}
					</div>
				</transition>
			</div>
		</template>

		<template v-slot:mobile>
			<div
				class="header-bar-mobile"
				v-if="show"
				:style="[{ boxShadow: shadow ? '0px 2px 4px 0px #0000001A' : 'none' }]"
			>
				<div v-if="showBack" class="row items-center" style="padding: 6px">
					<q-icon
						class="cursor-pointer"
						name="sym_r_arrow_back_ios_new"
						size="20px"
						color="ink-1"
						@click="onReturn"
					/>
				</div>

				<transition name="fade">
					<div
						v-if="showTitle"
						:style="[{ marginLeft: offset + 'px' }]"
						class="header-title-mobile text-h6 text-ink-1"
					>
						{{ title }}
					</div>
				</transition>

				<div class="header-right">
					<slot name="right" />
				</div>
			</div>
		</template>
	</adaptive-layout>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import AdaptiveLayout from '../settings/AdaptiveLayout.vue';

defineProps({
	show: {
		type: Boolean,
		require: true,
		default: false
	},
	showBack: {
		type: Boolean,
		default: true
	},
	showTitle: {
		type: Boolean,
		default: true
	},
	shadow: {
		type: Boolean,
		default: false
	},
	title: {
		type: String,
		require: true,
		default: ''
	},
	offset: {
		type: Number,
		default: 0
	}
});

const emit = defineEmits(['onReturn']);
const router = useRouter();
const onReturn = () => {
	if (window.history && window.history.state && !window.history.state.back) {
		router.replace('/');
		return;
	}
	emit('onReturn');
};
</script>

<style scoped lang="scss">
.header-bar {
	width: 100%;
	height: 56px;
	background: transparent;
	text-align: center;
	z-index: 9999;

	.header-title {
		margin-left: 6px;
	}
}

.header-bar-mobile {
	width: 100%;
	height: 56px;
	background: transparent;
	text-align: center;
	z-index: 9999;
	display: flex;
	justify-items: center;
	align-items: center;
	padding: 15px 12px;

	.header-left {
		width: 40px;
		display: flex;
		align-items: start;
		justify-content: flex-start;
		flex-shrink: 0;
	}

	.header-title-mobile {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.header-right {
		min-width: 32px;
		display: flex;
		align-items: end;
		justify-content: flex-end;
		flex-shrink: 0;
	}
}
</style>
