<template>
	<QSectionStyle>
		<q-select
			v-bind="$attrs"
			popup-content-class="q-pa-sm my-select-wrapper"
			:emit-value="emitValue"
			class="custom-selection-wrapper"
			:disable="!$attrs?.options?.length"
		>
			<template v-slot:option="scope">
				<q-item
					class="row justify-between items-center cursor-pointer select-option-item"
					:class="[scope.selected ? 'selected-active' : '']"
					v-bind="scope.itemProps"
				>
					<q-item-section class="text-body3">
						<slot
							v-if="$slots['option-label']"
							name="option-label"
							v-bind="scope"
						></slot>
						<template v-else>
							{{
								$attrs['option-label']
									? $t(scope.opt[$attrs['option-label']])
									: scope.opt['label']
							}}
						</template>
					</q-item-section>

					<q-item-section avatar class="my-avatar-wrapper">
						<q-icon
							v-if="
								isEqual(
									emitValue
										? scope.opt[$attrs['option-value'] || 'value']
										: scope.opt,
									$attrs.modelValue
								)
							"
							color="primary"
							size="16px"
							name="sym_r_check_circle"
						/>
					</q-item-section>
				</q-item>
			</template>
			<template
				v-for="(slot, name) in $slots"
				v-slot:[name]="scope"
				:key="name"
			>
				<slot :name="name" v-if="scope" v-bind="scope"></slot>
				<slot :name="name" v-else></slot>
			</template>
		</q-select>
	</QSectionStyle>
</template>

<script lang="ts" setup>
import QSectionStyle from './QSectionStyle.vue';
import { isEqual } from 'lodash';
import { useSlots, computed } from 'vue';
interface Props {
	emitValue?: boolean;
}
defineProps<Props>();
</script>

<style lang="scss" scoped>
.my-select-wrapper {
	.select-option-item {
		gap: 4px;
		padding: 8px;
		&.selected-active {
			border-radius: 8px;
		}

		.my-avatar-wrapper {
			min-width: 20px;
			padding-left: 4px;
		}
	}

	::v-deep(.q-menu) {
		border-radius: 8px;
		background: $background-2;
		box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2);
	}
	::v-deep(.q-menu .q-virtual-scroll__content) {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
}
::v-deep(
		.custom-selection-wrapper.q-field--highlighted .q-field__control:after
	) {
	border-width: 1px;
}
</style>
<style lang="scss">
.my-select-wrapper .q-manual-focusable--focused > .q-focus-helper {
	background: $background-hover !important;
	opacity: 1 !important;
}
</style>
