<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Access via browser')"
		:ok="t('i_got_it')"
		:cancel="false"
		:size="'small'"
		:noRouteDismiss="true"
		:platform="'mobile'"
		@onSubmit="onDialogOK"
	>
		<UserStatusCommonContent
			@on-dialog-o-k="onDialogOK"
			:btn-title="t('i_got_it')"
			:btnRedefined="true"
			:messageCenter="true"
			:message="
				t(
					'You can use Olares by accessing the following URL through a computer browser.'
				)
			"
		>
			<template v-slot:more>
				<div class="column items-center">
					<div
						class="row items-center justify-center q-mt-lg bg-background-3 url-bg text-body2 text-ink-2"
						style="width: 100%"
					>
						{{ url }}
					</div>
					<div
						class="copy-bg q-mt-md row justify-center items-center"
						:class="'text-ink-2'"
						@click="copyFunc"
					>
						<q-icon size="16px" name="sym_r_content_copy" />

						<span class="paste text-body3 q-ml-xs">
							{{ t('Copy URL') }}
						</span>
					</div>
				</div>
			</template>
		</UserStatusCommonContent>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import UserStatusCommonContent from '../userStatusDialog/UserStatusCommonContent.vue';
import { useI18n } from 'vue-i18n';
import { getPlatform } from '@didvault/sdk/src/core';
import { useUserStore } from '../../stores/user';
import { ref } from 'vue';
import { notifyFailed, notifySuccess } from '../../utils/notifyRedefinedUtil';

const { t } = useI18n();

const userStore = useUserStore();

const url = ref(
	userStore.getModuleSever('desktop', undefined, undefined, false)
);

const copyFunc = async () => {
	const platform = getPlatform();
	platform
		.setClipboard(url.value)
		.then(() => {
			notifySuccess(t('copy_success'));
			onDialogOK();
		})
		.catch(() => {
			notifyFailed(t('copy_fail'));
		});
};

const CustomRef = ref();

const onDialogOK = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style lang="scss" scoped>
.url-bg {
	width: 100%;
	// height: 36px;
	border-radius: 8px;
	padding: 8px 8px;
	// overflow-wrap: break-word;
	word-break: break-all;
}
.copy-bg {
	// border: 1px solid $separator;
	padding: 0px 8px;
	height: 24px;
	width: auto;
	border-radius: 4px;
	border: 1px solid $separator;
	display: inline-block;
	text-align: center;

	.paste {
		line-height: 24px;
	}
}
</style>
