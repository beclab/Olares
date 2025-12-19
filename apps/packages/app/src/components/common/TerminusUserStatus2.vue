<template>
	<div
		class="terminus-user-status row items-center"
		:class="`${
			termipassStore.totalStatus?.isError == UserStatusActive.error
				? 'text-red-1'
				: termipassStore.totalStatus?.isError == UserStatusActive.normal
				? 'text-background-1'
				: 'text-light-green-1'
		}`"
		@click.stop="itemClick"
	>
		<q-icon
			:name="`sym_r_${termipassStore.totalStatus?.icon}`"
			:color="
				configIconClass(
					termipassStore.totalStatus?.isError || UserStatusActive.normal
				)
			"
			size="12px"
			v-if="
				termipassStore.totalStatus?.status != TermiPassStatus.VPNConnecting &&
				termipassStore.totalStatus?.status != TermiPassStatus.VPNDisconnecting
			"
		/>
		<div v-else class="row items-center justify-center q-mr-xs">
			<svg
				class="vital-spinner"
				xmlns="http://www.w3.org/2000/svg"
				width="12"
				height="12"
				viewBox="0 0 200 200"
				fill="none"
				color="#3f51b5"
			>
				<defs>
					<linearGradient id="spinner-secondHalf">
						<stop offset="0%" stop-opacity="0" stop-color="#B8B8B8" />
						<stop offset="100%" stop-opacity="0.45" stop-color="#00CDFF" />
					</linearGradient>
					<linearGradient id="spinner-firstHalf">
						<stop offset="0%" stop-opacity="1" stop-color="#1080BF" />
						<stop offset="100%" stop-opacity="0.5" stop-color="#33D1F9" />
					</linearGradient>
				</defs>

				<g stroke-width="24">
					<path
						stroke="url(#spinner-secondHalf)"
						d="M 12 100 A 88 88 0 0 1 188 100"
					/>
					<path
						stroke="url(#spinner-firstHalf)"
						d="M 188 100 A 88 88 0 0 1 12 100"
					/>
					<path
						stroke="#1080BF"
						stroke-linecap="round"
						d="M 12 100 A 88 88 0 0 1 12 98"
					/>
				</g>
			</svg>
		</div>
		<div
			class="text-subtitle3 status"
			:class="
				configTitleClass(
					termipassStore.totalStatus?.isError || UserStatusActive.normal
				)
			"
		>
			{{ termipassStore.totalStatus?.title }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { useTerminusUserStatus } from 'src/composables/mobile/useTerminusUserStatus';
import { TermiPassStatus } from '../../utils/termipassState';
const emit = defineEmits(['superAction']);
const {
	itemClick,
	configIconClass,
	configTitleClass,
	termipassStore,
	UserStatusActive
} = useTerminusUserStatus(emit);
</script>

<style scoped lang="scss">
.terminus-user-status {
	border-radius: 12px;
	.status-active {
		color: $green;
	}

	.status-normal {
		color: $grey;
	}

	.status-error {
		color: $red;
	}

	.vital-spinner {
		animation: animate 1.3s linear infinite;
		-webkit-animation: animate 1.3s linear infinite;
		-moz-animation: animate 1.3s linear infinite;
	}

	// Spinner 2
	@keyframes animate {
		0% {
			transform: rotate(0deg);
			-webkit-transform: rotate(0deg);
		}

		100% {
			transform: rotate(360deg);
			-webkit-transform: rotate(360deg);
		}
	}

	@-webkit-keyframes animate {
		0% {
			transform: rotate(0deg);
			-webkit-transform: rotate(0deg);
		}

		100% {
			transform: rotate(360deg);
			-webkit-transform: rotate(360deg);
		}
	}

	.status {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		margin-left: 3px;
		// width: calc(100% - 17px);
		// background-color: red;
		// font-size: 9px;
	}
}
</style>
