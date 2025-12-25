<template>
	<div class="mobile-user-select">
		<div
			class="mobile-user-select__bg terminus_background_edit_base"
			:class="editBorderClass"
		>
			<q-input
				v-model="modelValue"
				class="text-body3 mobile-user-select__bg__input"
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
			>
				<template v-slot:prepend>
					<q-icon class="search_icon" name="search" size="20px" color="ink-3" />
				</template>
			</q-input>
		</div>
		<div
			class="row items-center q-gutter-x-sm q-gutter-y-sm q-mt-md"
			v-if="selectUsers.length > 0"
		>
			<template v-for="user in selectUsers" :key="user.name">
				<slot name="select-avatar" :user="user" />
			</template>
		</div>

		<q-separator spaced v-if="selectUsers.length > 0" class="q-mt-md" />

		<div class="text-caption-m text-light-blue-default q-mt-lg">
			{{ t('files.Selected') + ': ' + selectUsers.length }}
		</div>

		<q-list class="q-pa-none q-mt-md" v-if="usersOptions.length > 0">
			<template v-for="user in usersOptions" :key="user.name">
				<q-item
					class="row items-center justify-start text-ink-3 q-pa-none"
					style="height: 56px; border-radius: 8px"
					clickable
					dense
					@click="user.selected = !user.selected"
				>
					<terminus-check-box
						v-model="user.selected"
						activeImage="./img/checkbox/check_box_yellow.svg"
					/>
					<div class="row items-center q-ml-lg">
						<slot name="list-avatar" :user="user" />
						<div class="text-body1 text-ink-2 q-ml-sm">
							{{ user.name }}
						</div>
					</div>
				</q-item>
				<div class="q-pl-lg">
					<q-separator inset />
				</div>
			</template>
		</q-list>
		<div v-else style="height: 40px" class="row items-center justify-center">
			{{ $t('files.lonely') }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, inject, ref } from 'vue';
import TerminusCheckBox from '../../common/TerminusCheckBox.vue';
import { useI18n } from 'vue-i18n';

interface UsersProps {
	name: string;
	selected: boolean;
	olaresId?: string;
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

const { t } = useI18n();

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

const usersOptions = computed(() => {
	if (!modelValue.value || modelValue.value.length == 0) {
		return props.users;
	}
	return props.users.filter((e) => e.name.includes(modelValue.value));
});

const selectUsers = computed(() => {
	return props.users.filter((e) => e.selected) || [];
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

.mobile-user-select {
	// width: auto;
	height: calc(75vh - 80px);

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
		padding-inline: 12px;
		gap: 8px;

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
