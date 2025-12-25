<template>
	<div>
		<q-responsive
			class="circle-first"
			:ratio="1"
			style="max-width: 498px; min-width: 498px"
		>
		</q-responsive>
		<q-responsive
			class="circle-second"
			:ratio="1"
			style="max-width: 306px; min-width: 306px"
		>
		</q-responsive>
	</div>
</template>

<script setup lang="ts"></script>

<style lang="scss" scoped>
.circle-first {
	position: absolute;
	border-radius: 50%;
	background: $light-blue-soft;
	filter: blur(250px);
	right: 214px;
	top: -102px;
}
.circle-second {
	position: absolute;
	border-radius: 50%;
	background: $teal-soft;
	filter: blur(250px);
	right: 40px;
	top: -39px;
}
</style>
