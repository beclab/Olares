<template>
	<q-input
		debounce="500"
		borderless
		color="orange-9"
		placeholder="Search"
		class="bt-search"
		v-model="value"
		v-bind="$attrs"
		input-class="text-color-title"
	>
		<template v-slot:prepend>
			<q-icon class="search-icon" name="search" />
		</template>
	</q-input>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';

export default defineComponent({
	name: 'BtSearch',
	props: {},
	setup() {
		const value = ref('');
		return {
			value
		};
	}
});
</script>

<style scoped lang="scss">
.bt-search {
	text-align: center;
	margin: 32px auto 0;
	background: rgba(255, 255, 255, 0.4);
	box-shadow: 0px 2px 12px 0px rgba(0, 0, 0, 0.2),
		inset 0px 0px 6px 0px rgba(255, 255, 255, 0.7);
	border-radius: 8px;
	backdrop-filter: blur(16px);
	width: 320px;
	padding-left: 100px;
	.search-icon {
		color: $white;
	}

	&:focus-within {
		animation: aniUp 0.2s linear 1 forwards;
	}

	@keyframes aniUp {
		from {
			padding-left: 100px;
		}
		to {
			padding-left: 20px;
		}
	}
}
</style>
