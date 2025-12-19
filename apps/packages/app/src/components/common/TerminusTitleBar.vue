<template>
	<div
		class="header-bar row justify-between items-center"
		v-if="show"
		:style="{
			boxShadow: 'none',
			background: translate ? '' : 'translate'
		}"
		:class="isDark ? 'bg-shadow-color' : ''"
	>
		<div class="row justify-start items-center">
			<q-btn
				class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
				icon="sym_r_chevron_left"
				text-color="ink-2"
				@click="onReturn"
			>
			</q-btn>

			<div
				class="icon-container column justify-center items-center"
				@click="onLeftSecondClick"
			>
				<q-icon
					v-if="leftSecondIcon"
					:name="leftSecondIcon"
					size="24px"
					color="ink-2"
				/>
			</div>
		</div>

		<div
			class="header-title text-body1 text-color-title"
			:class="isDark ? 'text-white' : 'text-ink-1'"
		>
			{{ title }}
		</div>

		<div class="row justify-end items-center">
			<div
				class="icon-container column justify-center items-center"
				@click="onRightSecondClick"
			>
				<q-icon
					v-if="rightSecondIcon"
					:name="rightSecondIcon"
					size="24px"
					:color="isDark ? 'white' : 'text-ink-2'"
				/>
			</div>

			<div
				class="icon-container column justify-center items-center"
				@click="onRightClick"
			>
				<q-icon
					v-if="rightIcon && showRightIcon"
					:name="rightIcon"
					size="24px"
					color="ink-2"
				/>
			</div>
			<div
				v-if="rightText"
				class="right-text text-body2 column justify-center items-center"
				@click.stop="onRightTextClick"
			>
				{{ rightText }}
			</div>
			<div class="column justify-center items-center" v-if="$slots.right">
				<slot name="right"></slot>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';

import { useQuasar } from 'quasar';
import { getNativeAppPlatform } from 'src/application/platform';

const $q = useQuasar();

const router = useRouter();
const props = defineProps({
	show: {
		type: Boolean,
		require: false,
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
	rightIcon: {
		type: String,
		require: false,
		default: ''
	},
	translate: {
		type: Boolean,
		default: false
	},
	rightText: {
		type: String,
		require: false,
		default: ''
	},
	showRightIcon: {
		type: Boolean,
		require: false,
		default: true
	},
	rightSecondIcon: {
		type: String,
		require: false,
		default: ''
	},
	leftSecondIcon: {
		type: String,
		require: false,
		default: ''
	},
	isDark: {
		type: Boolean,
		require: false,
		default: false
	},
	hookBackAction: {
		type: Boolean,
		require: false,
		default: false
	}
});

const emit = defineEmits([
	'onRightClick',
	'onRightSecondClick',
	'onLeftSecondClick',
	'onRightTextClick',
	'onReturnAction'
]);
const onReturn = () => {
	if (props.hookBackAction) {
		emit('onReturnAction');
		return;
	}

	if ($q.platform.is.nativeMobile) {
		getNativeAppPlatform().hookBackAction();
		return;
	}

	if (process.env.PLATFORM !== 'BEX') {
		if (window.history.length <= 1) {
			return;
		}
	}
	router.go(-1);
};
const onRightClick = () => {
	emit('onRightClick');
};

const onRightTextClick = () => {
	emit('onRightTextClick');
};

const onRightSecondClick = () => {
	emit('onRightSecondClick');
};

const onLeftSecondClick = () => {
	emit('onLeftSecondClick');
};
</script>

<style scoped lang="scss">
.header-bar {
	width: 100%;
	height: 56px;
	text-align: center;
	padding: 0 20px;
	position: relative;

	.icon-container {
		width: 32px;
		height: 32px;
	}

	.right-text {
		height: 56px;

		text-align: right;
		color: $blue-4;
		padding-left: 20px;
	}

	.header-title {
		position: absolute;
		top: 16px;
		left: 84px;
		right: 84px;
		width: calc(100% - 84px - 84px);
		word-wrap: break-word;
		word-break: break-all;
		white-space: nowrap;
		overflow: hidden;
	}
}
</style>
