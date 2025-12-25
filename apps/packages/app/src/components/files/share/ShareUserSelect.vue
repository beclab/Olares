<template>
	<div class="terminus-edit">
		<div
			class="terminus-edit__bg terminus_background_edit_base"
			:class="editBorderClass"
		>
			<div
				v-for="person in users?.filter((e) => e.selected)"
				:key="person.name"
				class="selected-person text-subtitle2 q-px-xs"
			>
				<span>{{ person.name }}</span>
				<button class="remove-btn" @click="person.selected = false">×</button>
			</div>
			<q-input
				v-model="modelValue"
				class="text-body3 terminus-edit__bg__input"
				bg-color="transparent"
				:placeholder="hintText"
				borderless
				:inputClass="inputClass + ' custom-input-padding'"
				:readonly="isReadOnly"
				@update:model-value="onTextChange"
				dense
				@keyup.enter="submit"
				@focus="onFocus"
				@blur="onBlur"
				@keydown="handleKeydown"
			>
				<template v-slot:append>
					<q-icon size="24px" name="sym_r_add" color="text-ink-2" />
				</template>
			</q-input>
		</div>
		<q-menu :offset="[0, 10]" fit>
			<q-list
				style="max-height: 200px"
				class="q-pa-sm"
				v-if="usersOptions.length > 0"
			>
				<q-item
					v-for="user in usersOptions"
					:key="user.name"
					class="row items-center justify-start text-ink-3"
					style="height: 40px; border-radius: 8px"
					clickable
					dense
					@click="user.selected = !user.selected"
				>
					<terminus-check-box
						v-model="user.selected"
						:label="user.name"
						@update:modelValue="
							(value) => {
								console.log('usersRef ===>', users);
							}
						"
					/>
				</q-item>
			</q-list>
			<div v-else style="height: 40px" class="row items-center justify-center">
				{{ $t('files.lonely') }}
			</div>
		</q-menu>
	</div>
</template>

<script setup lang="ts">
import { computed, inject, ref } from 'vue';
import TerminusCheckBox from '../../common/TerminusCheckBox.vue';

interface UsersProps {
	name: string;
	selected: boolean;
}

const props = defineProps({
	hintText: {
		type: String,
		default: '',
		required: false
	},

	isReadOnly: {
		type: Boolean,
		default: false,
		require: false
	},

	inputClass: {
		type: String,
		default: 'text-ink-1',
		require: false
	},
	users: {
		type: Array<UsersProps>,
		required: false,
		default: []
	}
});

const emit = defineEmits(['onTextChange', 'update:modelValue', 'submit']);

function onTextChange(value: any) {
	if (setBlured) {
		setBlured(false);
	}
}

const submit = () => {
	emit('submit');
};

const setFocused = inject('setFocused') as any;
const setBlured = inject('setBlured') as any;
let focused = ref(false);
const modelValue = ref('');
const menuShow = ref(false);

const onFocus = () => {
	focused.value = true;
	if (setFocused) {
		setFocused(true);
	}
	menuShow.value = true;
};
const onBlur = () => {
	focused.value = false;
	if (setBlured) {
		setBlured(true);
	}
	menuShow.value = false;
};

const editBorderClass = computed(() => {
	if (props.isReadOnly) {
		return 'terminus_input_border_edit_normal';
	}
	if (focused.value) {
		return 'terminus_input_border_editing';
	}
	return 'terminus_input_border_edit_normal';
});

const handleKeydown = (event) => {
	const selectUsers = props.users?.filter((e) => e.selected);
	if (
		event.key === 'Backspace' &&
		modelValue.value.length == 0 &&
		selectUsers &&
		selectUsers?.length > 0
	) {
		selectUsers[selectUsers.length - 1].selected = false;
	}
};

const usersOptions = computed(() => {
	if (!modelValue.value || modelValue.value.length == 0) {
		return props.users;
	}
	return props.users.filter((e) => e.name.includes(modelValue.value));
});
</script>

<style lang="scss" scoped>
.terminus_background_edit_base {
	height: 40px;
	backdrop-filter: blur(6.07811px);
	border-radius: 8px;
	background: transparent;
}

.terminus_input_border_edt_read_only {
	background: linear-gradient(0deg, $background-3, $background-3);
	border: 1px solid $separator;
}

.terminus-edit {
	width: auto;

	.terminus_input_border_edit_normal {
		border: 1px solid $separator;
	}

	.terminus_input_border_editing {
		border: 1px solid $blue-default;
	}

	.terminus_input_border_edit_error {
		border: 1px solid $red;
	}

	&__label {
		color: $ink-3;
	}

	&__bg {
		width: 100%;
		margin-top: 4px;
		position: relative;

		display: flex;
		flex-wrap: wrap;
		align-items: center;
		// border-radius: 12px;
		padding-inline: 12px;
		// min-height: 50px;
		gap: 8px;
		// background: red;

		&__input {
			height: 100%;
			width: calc(100% - 30px);
			flex: 1;
			white-space: wrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.selected-person {
			display: flex;
			align-items: center;
			border-radius: 4px;
			gap: 8px;
			background-color: $background-3;
			height: 28px;
			color: $ink-2;
		}

		.remove-btn {
			background: none;
			border: none;
			color: $ink-3;
			cursor: pointer;
			font-size: 16px;
			line-height: 1;
			padding: 0;
			width: 16px;
			height: 16px;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 50%;
		}

		.remove-btn:hover {
			background: #bbdefb;
		}
	}

	&__input__less_width {
		width: calc(100% - 64px);
	}

	&__error {
		width: 100%;
		margin-top: 4px;
		color: $red;
	}
}
</style>
