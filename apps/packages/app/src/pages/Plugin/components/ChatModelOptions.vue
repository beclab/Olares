<template>
	<q-list class="q-gutter-y-xs">
		<q-item
			v-for="item in options"
			:key="item.value"
			active-class="my-active-class"
			class="my-option-item q-pa-sm"
			clickable
			v-ripple
			:active="modelValue.id === item.id"
			@click="clickHandler(item)"
		>
			<div class="row items-center no-wrap full-width item-wrapper">
				<div class="row items-center no-wrap full-width">
					<div style="flex: 0 0 24px">
						<q-img
							:src="item.img"
							:ratio="1"
							width="24px"
							spinner-size="xs"
							class="q-mr-sm"
						/>
					</div>
					<div class="text-body2 text-grey-10 ellipsis">
						{{ item.name }}
					</div>
					<div style="flex: 0 0 24px">
						<q-icon
							v-show="modelValue.id === item.id"
							name="check"
							color="light-blue-6"
							size="16px"
							class="q-ml-xs"
						/>
					</div>
				</div>
				<div
					class="delete-container row items-center justify-center"
					style="flex: 0 0 24px; height: 24px"
					@click.stop="deleteHanlder(item)"
				>
					<q-icon
						name="delete"
						color="negative"
						size="16px"
						class="q-ml-xs delete-icon"
					/>
				</div>
			</div>
		</q-item>
	</q-list>
</template>

<script setup lang="ts">
interface Props {
	modelValue: any;
	options: any;
}

withDefaults(defineProps<Props>(), {});
const emits = defineEmits(['update:modelValue', 'delete']);

const clickHandler = (data) => {
	emits('update:modelValue', data);
};
const deleteHanlder = (data) => {
	emits('delete', data);
};
</script>

<style lang="scss" scoped>
.delete-container {
	.delete-icon {
		display: none;
	}
}
.delete-container:hover {
	.delete-icon {
		display: block;
	}
}
.my-option-item {
	border-radius: 8px;
}
.my-active-class {
	background: $grey-1;
}
</style>
