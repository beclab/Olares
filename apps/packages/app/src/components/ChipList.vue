<template>
	<div class="chip-list-container q-mt-md">
		<q-chip
			v-for="(item, index) in modelValue"
			:key="item.name"
			:color="[isExpired(item.expires) ? 'red-alpha' : 'background-hover']"
			text-color="ink-1"
			square
			:class="itemIndex === index ? 'selected-item' : 'normal-item'"
			class="text-ink-1 chip-wrapper q-px-md q-py-sm"
			size="12px"
		>
			<q-tooltip>
				<q-markup-table
					dense
					class="chip-list-table"
					separator="none"
					dark
					flat
				>
					<tbody>
						<tr v-for="(value, key) in item" :key="key" class="text-white">
							<td class="text-left">{{ key }}</td>
							<td class="text-left">
								<div style="max-width: 150px" class="ellipsis">
									{{ formatExpires(value) }}
								</div>
							</td>
						</tr>
					</tbody>
				</q-markup-table>
			</q-tooltip>
			<q-icon
				v-if="isExpired(item.expires)"
				class="q-mr-xs"
				name="sym_r_error"
				color="negative"
			/>
			<div
				class="text-body3 text-ink-2"
				@click="() => clickHandler(item, index)"
			>
				{{ item.name }}
			</div>
			<q-icon
				class="q-ml-xs"
				name="sym_r_close"
				color="ink-3 delete-icon"
				@click="deleteHandler(index)"
			/>
		</q-chip>
	</div>
	<div class="q-mt-md" v-show="visible">
		<q-input
			v-model="input"
			type="textarea"
			class="input-wrapper"
			input-style="height : 60px"
			borderless
			autogrow
		/>
		<div class="row justify-end items-center text-right">
			<q-btn
				dense
				flat
				class="cancel-btn q-px-md q-mt-md q-mr-md"
				:label="t('cookie_cancel')"
				@click="cancelHandler"
			/>
			<q-btn
				dense
				flat
				class="confirm-btn q-px-md q-mt-md"
				:label="t('cookie_save')"
				@click="saveHandler"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { date } from 'quasar';

interface Props {
	modelValue: { name: string; value: string; expires: number }[];
}

const props = withDefaults(defineProps<Props>(), {});
const emits = defineEmits(['update:modelValue']);
const itemIndex = ref(-1);
const visible = ref(false);
const { t } = useI18n();
const input = ref();
const deleteHandler = (index) => {
	const data = [...props.modelValue];
	data.splice(index, 1);
	if (itemIndex.value === index) {
		visible.value = false;
		itemIndex.value = -1;
	}
	emits('update:modelValue', data);
};

function formatExpires(value: any) {
	if (typeof value !== 'number' || isNaN(value)) {
		return value;
	}

	const integerValue = Math.floor(value);
	const timestamp = integerValue.toString().length <= 10 ? value * 1000 : value;

	const now = Date.now();
	const maxFuture = now + 100 * 365 * 24 * 60 * 60 * 1000;
	if (timestamp <= 0 || timestamp > maxFuture) {
		return value;
	}

	return date.formatDate(timestamp, 'YYYY-MM-DD HH:mm:ss');
}

function isExpired(expires: any): boolean {
	if (expires == undefined || expires == 0 || expires == null) {
		return false;
	}
	const currentTime = Math.floor(Date.now() / 1000);
	return currentTime > expires;
}

const clickHandler = (item, index) => {
	console.log('clickHandler');
	itemIndex.value = index;
	input.value = item.value;
	visible.value = true;
};

const saveHandler = () => {
	const data = [...props.modelValue];
	const obj = { ...props.modelValue[itemIndex.value], value: input.value };
	console.log(input.value);
	console.log(obj);
	data.splice(itemIndex.value, 1, obj);
	console.log(data);
	emits('update:modelValue', data);
	visible.value = false;
	itemIndex.value = -1;
};

const cancelHandler = () => {
	visible.value = false;
	itemIndex.value = -1;
};
</script>

<style lang="scss" scoped>
.chip-list-container {
	.chip-wrapper {
		cursor: pointer;
	}

	.selected-item {
		border: 1px solid $separator;
	}

	.normal-item {
		border: 1px solid transparent;
	}
}
.chip-list-table {
	background: rgba(0, 0, 0, 0);
	color: #fff;
}
.delete-icon {
	cursor: pointer;
	&:hover {
		background: $background-1;
		border-radius: 50%;
	}
}

.input-wrapper {
	padding: 12px;
	border-radius: 8px;
	border: solid 1px $input-stroke;

	::v-deep(textarea) {
		resize: none;
		padding-top: 0;
	}
}
::v-deep(.input-wrapper.q-field--outlined .q-field__control:before) {
	border-color: $input-stroke-hover !important;
}
</style>
