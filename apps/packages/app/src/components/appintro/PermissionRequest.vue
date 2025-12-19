<template>
	<div class="permission-request-root column justify-center">
		<div class="row justify-start items-center">
			<q-icon size="20px" color="ink-2" :name="name" />
			<div class="text-ink-2 text-subtitle2 q-ml-xs">{{ title }}</div>
		</div>
		<div
			class="row justify-start q-ml-xs"
			style="width: calc(100% - 20px)"
			v-for="(item, num) in nodes"
			:key="num"
		>
			<div class="column justify-start q-ml-lg">
				<div class="text-overline text-ink-3">{{ item.label }}</div>
				<template v-for="(child, index) in item.children" :key="index">
					<div class="row justify-start items-center">
						<div class="permission-circle-parent">
							<div class="permission-circle bg-ink-3" />
						</div>
						<div
							class="text-overline text-ink-3"
							style="width: calc(100% - 12px)"
						>
							{{ child.label }}
						</div>
					</div>
					<div class="q-ml-md">
						<template v-for="(child2, index2) in child.children" :key="index2">
							<div class="row justify-start items-center">
								<div class="permission-circle-parent">
									<div class="permission-circle bg-ink-3" />
								</div>
								<div
									class="text-overline text-ink-3"
									style="width: calc(100% - 12px)"
								>
									{{ child2.label }}
								</div>
							</div>
						</template>
					</div>
				</template>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { PermissionNode } from '../../constant/constants';

defineProps({
	name: String,
	title: String,
	nodes: Object as PropType<PermissionNode[]>
});
</script>

<style scoped lang="scss">
.permission-request-root {
	width: 100%;
	height: auto;

	.permission-circle-parent {
		width: 12px;
		height: 12px;
		padding: 5px;

		.permission-circle {
			width: 2px;
			height: 2px;
			border-radius: 50%;
		}
	}
}
</style>
