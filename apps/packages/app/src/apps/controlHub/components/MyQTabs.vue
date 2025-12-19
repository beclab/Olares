<template>
	<q-splitter v-model="splitterModel" style="height: 100%">
		<template v-slot:before>
			<div style="padding-top: 12px">
				<q-tabs
					v-model="tab"
					vertical
					align="left"
					class="text-grey"
					active-color="primary"
					indicator-color="primary"
				>
					<q-tab
						v-for="item in list"
						:key="item.name"
						:name="item.name"
						style="min-height: 44px"
						no-caps
					>
						<span class="my-tab-wrapper">
							{{ item.label }}
						</span>
					</q-tab>
				</q-tabs>
			</div>
		</template>
		<template v-slot:after>
			<q-tab-panels
				v-model="tab"
				animated
				style="padding: 0; background: #f8f8f8; margin: -8px; min-height: 100%"
			>
				<slot></slot>
			</q-tab-panels>
		</template>
	</q-splitter>
</template>

<script setup lang="ts">
import { ref } from 'vue';
interface Props {
	defaultValue: string;
	list: Array<{
		name: string;
		label: string;
	}>;
}
const splitterModel = ref(12);
const props = withDefaults(defineProps<Props>(), {});
const tab = ref(props.defaultValue);
</script>

<style></style>
