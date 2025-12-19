<template>
	<div
		:class="[
			'listRow',
			'q-py-md',
			'accountInfo',
			{ isMouse: mouseFlag == field.type && !editing }
		]"
		@mouseover="handleMouseover(field.type)"
		@mouseout="handleMouseout()"
	>
		<div class="row items-center justify-between info">
			<div class="col-10 q-pl-lg">
				<div>
					<div class="row items-center text-light-blue-default">
						<q-icon :name="field.icon" size="20px" />
						<span class="text-li-title text-body3">{{
							translate(field.name)
						}}</span>
					</div>
				</div>
				<div class="q-mt-sm" v-if="editing" style="height: 32px">
					<q-input
						v-if="field.type == 'date'"
						borderless
						dense
						v-model="fieldValue"
						mask="date"
						:rules="['date']"
						@update:model-value="onUpdate"
						:placeholder="placeholder[field.type]"
						input-class="text-body3 text-ink-2"
						class="filedInput"
						input-style="width: 100%; text-indent: 10px; height: 32px;"
					>
						<template v-slot:append>
							<q-icon
								name="sym_r_edit_calendar"
								size="16px"
								class="cursor-pointer filedIcon q-mr-sm"
							>
								<q-popup-proxy
									ref="qDateProxy"
									cover
									transition-show="scale"
									transition-hide="scale"
								>
									<q-date
										v-model="fieldValue"
										minimal
										@update:model-value="onUpdate"
									>
										<div class="row items-center justify-end">
											<q-btn v-close-popup label="Close" color="primary" flat />
										</div>
									</q-date>
								</q-popup-proxy>
							</q-icon>
						</template>
					</q-input>

					<q-input
						v-else-if="field.type == 'month'"
						dense
						v-model="fieldValue"
						borderless
						mask="####/##"
						@update:model-value="onUpdate"
						:placeholder="placeholder[field.type]"
						input-class="text-body3 text-ink-2"
						class="filedInput"
						input-style="width: 100%; text-indent: 10px;"
					>
						<template v-slot:append>
							<q-icon
								name="sym_r_edit_calendar"
								size="16px"
								class="cursor-pointer filedIcon q-mr-sm"
							>
								<q-menu class="filedMenu" v-model="showAddField">
									<q-list>
										<q-item>
											<BtMonthPicker
												color="primary"
												@input="focusInput"
												:value="fieldValue"
											/>
										</q-item>
									</q-list>
								</q-menu>
							</q-icon>
						</template>
					</q-input>

					<q-input
						v-else
						borderless
						dense
						v-model="fieldValue"
						@update:model-value="onUpdate"
						:placeholder="placeholder[field.type]"
						input-style="width: 100%; text-indent: 10px;"
						input-class="text-body3 text-body3"
						class="filedInput"
					>
						<template v-slot:append>
							<q-icon
								class="cursor-pointer q-mr-sm"
								name="sym_r_qr_code"
								size="16px"
								v-if="field.type === 'totp' && isMobile"
								@click="startScan"
							/>
						</template>
					</q-input>
				</div>
				<div class="text-ink-1 userInfo" v-else>
					<div
						v-if="
							['password', 'pin', 'apiSecret', 'mnemonic'].includes(
								fieldType
							) && displayVaule !== 'empty'
						"
					>
						<q-input
							input-class="text-ink-1 text-body3"
							v-model="displayVaule"
							dense
							borderless
							readonly
							:type="
								mouseFlag && app.settings.unmaskFieldsOnHover
									? 'text'
									: 'password'
							"
						/>
					</div>
					<div v-else-if="fieldType == 'totp'">
						<Totp :secret="displayVaule" ref="myTotp" />
					</div>
					<div v-else-if="fieldType == 'note'">
						<q-input
							input-class="text-ink-1 text-body3"
							v-model="displayVaule"
							dense
							borderless
							type="textarea"
							readonly
						/>
					</div>
					<div
						v-else
						:class="displayVaule !== 'empty' ? 'hasValue' : 'text-ink-2'"
					>
						<q-input
							input-class="text-ink-1 text-body3"
							v-model="displayVaule"
							dense
							borderless
							readonly
						/>
					</div>
				</div>
			</div>

			<div class="col-2 row items-center justify-center" v-if="editing">
				<span class="trashCan text-grey-7" @click="onRemove">
					<q-btn
						class="btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_delete"
						text-color="ink-2"
					>
						<q-tooltip>{{ t('Delete') }}</q-tooltip>
					</q-btn>
				</span>
				<div class="q-ml-xs">
					<span class="q-mb-xs row items-center" @click="onMoveUp">
						<q-btn
							class="btn-size-xs btn-no-text btn-no-border"
							icon="sym_r_arrow_upward"
							text-color="ink-2"
						>
						</q-btn>
					</span>
					<span class="row items-center" @click="onMoveDown">
						<q-btn
							class="btn-size-xs btn-no-text btn-no-border"
							icon="sym_r_arrow_downward"
							text-color="ink-2"
						>
						</q-btn>
					</span>
				</div>
			</div>

			<div class="row item-center justify-end infoOperateWrap">
				<div
					class="row items-center justify-center q-px-sm q-py-xs q-mr-md bg-grey-11-hover infoOperate text-ink-1"
					@click="openUrl(fieldValue)"
					v-if="field.type == 'url' && fieldValue.startsWith('http')"
				>
					<q-icon name="sym_r_language" size="14px" />
					<span class="text-body3 q-ml-xs">{{ t('base.open') }}</span>
				</div>
				<div
					class="row items-center justify-center q-px-sm q-py-xs q-mr-md bg-grey-11-hover infoOperate text-ink-1"
					@click="copyFile(fieldValue)"
				>
					<q-icon name="sym_r_content_copy" size="14px" />
					<span class="text-body3 q-ml-xs">{{ t('copy') }}</span>
				</div>
				<div
					v-if="isEditable"
					class="row items-center justify-center q-px-sm q-py-xs bg-grey-11-hover infoOperate text-ink-1"
					@click="onEdit"
				>
					<q-icon name="sym_r_edit_note" size="14px" />
					<span class="text-body3 q-ml-xs">{{ t('buttons.edit') }}</span>
				</div>
				<div
					v-if="!isEditable"
					class="row items-center justify-center q-px-sm q-py-xs bg-grey-11-hover infoOperate text-ink-1"
					disabled
				>
					<q-icon name="sym_r_edit_note" size="14px" />
					<span class="text-body3 q-ml-xs">{{ t('buttons.edit') }}</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { computed, ref, watch, reactive } from 'vue';
import { Field, FIELD_DEFS, getPlatform } from '@didvault/sdk/src/core';
import BtMonthPicker from '../../components/base/BtMonthPicker.vue';
import Totp from './totp.vue';
import { app } from '../../globals';
import { notifyFailed, notifySuccess } from '../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import { translate } from '@didvault/sdk/src/util';
import { date } from 'quasar';

const props = defineProps({
	index: {
		type: Number,
		required: true
	},
	field: {
		type: Field,
		required: true
	},
	editing: {
		type: Boolean,
		required: true
	},
	canMoveUp: {
		type: Boolean,
		required: true
	},
	canMoveDown: {
		type: Boolean,
		required: true
	},
	isEditable: {
		type: Boolean,
		required: false
	}
});

const $q = useQuasar();
const _masked = false;
const fieldValue = ref(props.field.value);
const mouseFlag = ref();
const showAddField = ref(false);
const placeholderOption = {
	username: 'Enter value Here',
	password: 'Enter passWord',
	url: 'Enter value Here'
};
const placeholder = reactive(placeholderOption);
const myTotp = ref();

const _fieldDef = ref(FIELD_DEFS[props.field.type] || FIELD_DEFS.text);

const fieldType = ref(props.field.type);
const isMobile = ref($q.platform.is.mobile);

const format = ref(_fieldDef.value.format || ((value: string) => value));

const { t } = useI18n();

watch(
	() => props.field.value,
	(newValue, oldValue) => {
		if (newValue == oldValue) {
			return;
		}

		fieldValue.value = props.field.value;
		_fieldDef.value = FIELD_DEFS[props.field.type] || FIELD_DEFS.text;
		fieldType.value = props.field.type;
		format.value = _fieldDef.value.format || ((value: string) => value);
	}
);

const displayVaule = computed(function () {
	if (!props.field.value) {
		return t('empty');
	}

	switch (props.field.type) {
		case 'password':
		case 'pin':
			return format.value(props.field.value, _masked);
		case 'credit':
			return format.value(props.field.value, true);
		case 'totp':
			return props.field.value;
		case 'note':
			return props.field.value;
		default:
			return format.value(props.field.value, _masked);
	}
});

const emit = defineEmits([
	'fieldUpdate',
	'remove',
	'moveup',
	'movedown',
	'onEdit',
	'startScan'
]);

function onUpdate(val: any) {
	const dateRegExp = /^(\d{4})\/(\d{2})\/(\d{2})$/;
	if (props.field.type == 'date') {
		if (!dateRegExp.test(val)) {
			return;
		}
	}
	emit('fieldUpdate', { index: props.index, value: val });
}

function onRemove() {
	emit('remove');
}

function onMoveUp() {
	if (!props.canMoveUp) return;
	emit('moveup');
}

function onMoveDown() {
	if (!props.canMoveDown) return;
	emit('movedown');
}

function handleMouseover(type: string) {
	mouseFlag.value = type;
}

function handleMouseout() {
	mouseFlag.value = null;
}

function onEdit() {
	emit('onEdit');
}

const focusInput = (month, hide) => {
	const val = `${month.getFullYear()}/${month.getMonth() + 1}`;
	fieldValue.value = val;
	if (hide) {
		showAddField.value = false;
	}
	onUpdate(fieldValue);
};

const copyFile = (fieldValue: string) => {
	let copyTxt = fieldValue;
	if (fieldType.value === 'totp') {
		copyTxt = myTotp.value.token;
	}

	getPlatform()
		.setClipboard(copyTxt)
		.then(() => {
			notifySuccess(t('copy_success'));
		})
		.catch(() => {
			notifyFailed(t('copy_fail'));
		});
};

const openUrl = (url: string) => {
	if (!url.startsWith('http')) {
		return;
	}
	getPlatform().openExternalUrl(url);
};

const startScan = async () => {
	if (isMobile.value) {
		emit('startScan', props.index);
	}
};
</script>

<style lang="scss" scoped>
.listRow {
	.userInfo {
		font-size: map-get($map: $body2, $key: size);
		padding-left: 2px;

		.hasValue {
			width: 100%;
			height: 32px;
			line-height: 32px;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}
	}
}

.filedInput {
	width: 100%;
	height: 32px;
	font-size: map-get($map: $body1, $key: size);
	border: 1px solid $input-stroke;
	border-radius: 8px;
	::v-deep(.q-field__inner) {
		height: 32px;
	}
}
::v-deep(.q-field--dense .q-field__control) {
	height: 32px;
}

::v-deep(.q-field--dense .q-field__marginal) {
	height: 32px;
}

.accountInfo {
	overflow: hidden;

	.info {
		position: relative;

		.text-li-title {
			margin-left: 5px;
		}

		.infoOperateWrap {
			position: absolute;
			right: 20px;
			bottom: -30px;
			z-index: 1;

			.infoOperate {
				&:hover {
					border-radius: 4px;
					cursor: pointer;
					background-color: $background-hover;
				}
			}
		}
	}

	transition: all 0.5s;
}

.isMouse {
	padding-bottom: 40px;
}

.trashCan {
	cursor: pointer;
}
</style>
