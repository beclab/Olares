<template>
	<div class="wrapper">
		<div class="content" :class="{ collapsed: isCollapsed }">
			<slot></slot>
		</div>
		<div class="header" @click="isCollapsed = !isCollapsed">
			<div>
				<q-icon
					color="primary"
					:name="
						isCollapsed
							? 'keyboard_double_arrow_up'
							: 'keyboard_double_arrow_down'
					"
				></q-icon>
			</div>
		</div>
	</div>
</template>

<script>
import { defineComponent, ref } from 'vue';

export default defineComponent({
	name: 'CollapseWrapper',
	props: {
		title: {
			type: String,
			default: ''
		}
	},
	setup(props) {
		const isCollapsed = ref(false);
		return { isCollapsed };
	}
});
</script>

<style scoped lang="scss">
.wrapper {
	border-radius: 4px;
	.header {
		display: flex;
		justify-content: center;
		font-size: 18px;
		cursor: pointer;
	}
	.content {
		max-height: 64px;
		overflow: hidden;
		transition: max-height 0.6s ease-out;
		&.collapsed {
			max-height: 1000px;
		}
	}
}
</style>
