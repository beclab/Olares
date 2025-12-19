<template>
	<router-view v-slot="{ Component }">
		<transition
			enter-active-class="animated fadeInRight"
			leave-active-class="animated fadeOutRight"
		>
			<component :is="Component" />
		</transition>
	</router-view>
</template>

<script setup lang="ts"></script>
