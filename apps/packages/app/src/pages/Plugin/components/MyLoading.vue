<template>
	<div>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			class="vital-spinner"
			width="12"
			height="12"
			viewBox="0 0 12 12"
			fill="none"
		>
			<circle cx="6" cy="6" r="4.5" stroke="white" stroke-width="2" />
			<path
				d="M9.18198 2.81802L9.88909 2.11091L9.18198 2.81802ZM2.81802 2.81802L3.52513 3.52513L2.81802 2.81802ZM0.5 6C0.5 6.55228 0.947715 7 1.5 7C2.05228 7 2.5 6.55228 2.5 6H0.5ZM11.5 6C11.5 4.54131 10.9205 3.14236 9.88909 2.11091L8.47487 3.52513C9.13125 4.1815 9.5 5.07174 9.5 6H11.5ZM9.88909 2.11091C8.85764 1.07946 7.45869 0.5 6 0.5V2.5C6.92826 2.5 7.8185 2.86875 8.47487 3.52513L9.88909 2.11091ZM6 0.5C4.54131 0.5 3.14236 1.07946 2.11091 2.11091L3.52513 3.52513C4.1815 2.86875 5.07174 2.5 6 2.5V0.5ZM2.11091 2.11091C1.07946 3.14236 0.5 4.54131 0.5 6H2.5C2.5 5.07174 2.86875 4.1815 3.52513 3.52513L2.11091 2.11091Z"
				fill="url(#paint0_linear_131_9392)"
			/>
			<defs>
				<linearGradient
					id="paint0_linear_131_9392"
					x1="1.5"
					y1="6"
					x2="9.6"
					y2="6"
					gradientUnits="userSpaceOnUse"
				>
					<stop stop-color="$blue-4" />
					<stop offset="1" stop-color="$blue-4" stop-opacity="0" />
				</linearGradient>
			</defs>
		</svg>
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
