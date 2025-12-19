<template>
	<div
		class="absolute-full column items-center justify-center"
		style="z-index: 9999"
	>
		<svg
			class="vital-spinner"
			xmlns="http://www.w3.org/2000/svg"
			width="80"
			height="80"
			viewBox="0 0 80 80"
			fill="none"
		>
			<rect width="80" height="80" />
			<circle cx="40" cy="40" r="27" stroke="#EBF1FF" stroke-width="12" />
			<path
				d="M59.0919 21.4939L54.7005 25.5824L59.0919 21.4939ZM20.9081 21.4939L25.2995 25.5824L20.9081 21.4939ZM7 42C7 45.3137 9.68629 48 13 48C16.3137 48 19 45.3137 19 42H7ZM73 42C73 32.8503 69.6199 23.9967 63.4832 17.4054L54.7005 25.5824C58.6908 29.8683 61 35.7672 61 42H73ZM63.4832 17.4054C57.3331 10.7997 48.8961 7 40 7V19C45.4256 19 50.7237 21.311 54.7005 25.5824L63.4832 17.4054ZM40 7C31.1039 7 22.6669 10.7996 16.5168 17.4054L25.2995 25.5824C29.2763 21.311 34.5744 19 40 19V7ZM16.5168 17.4054C10.3801 23.9967 7 32.8503 7 42H19C19 35.7672 21.3092 29.8683 25.2995 25.5824L16.5168 17.4054Z"
				fill="url(#paint0_linear_2510_133377)"
			/>
			<defs>
				<linearGradient
					id="paint0_linear_2510_133377"
					x1="13"
					y1="42"
					x2="61.6"
					y2="42"
					gradientUnits="userSpaceOnUse"
				>
					<stop stop-color="#3377FF" />
					<stop offset="1" stop-color="#3377FF" stop-opacity="0" />
				</linearGradient>
			</defs>
		</svg>
		<div class="q-mt-lg text-body3 text-grey-8">
			{{ $t('PROGRESS_WAIT') }}
		</div>
	</div>
</template>

<script setup lang="ts"></script>

<style lang="scss" scoped>
.vital-spinner {
	animation: animate 1.3s linear infinite;
	-webkit-animation: animate 1.3s linear infinite;
	-moz-animation: animate 1.3s linear infinite;
}

@keyframes animate {
	0% {
		transform: rotate(360deg);
		-webkit-transform: rotate(360deg);
	}
	100% {
		transform: rotate(0deg);
		-webkit-transform: rotate(0deg);
	}
}
</style>
