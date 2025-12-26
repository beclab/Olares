<template>
	<q-dialog
		ref="dialogRef"
		@hide="onDialogCancel"
		:maximized="$q.platform.is.mobile"
		transition-show="slide-up"
		transition-hide="slide-down"
		class="d-creatVault"
	>
		<q-card
			class="column root items-center"
			:class="isDesktop ? 'q-dialog-plugin' : 'd-createVault-mobile'"
		>
			<DesktopTermipassUnlockContent
				v-if="isDesktop"
				@unlockSuccess="onDialogOK"
				@cancel="onDialogCancel"
				:detail-text="
					info && info.length > 0
						? info
						: t('unlock.auth_popup_unlock_introduce')
				"
				:logo="
					$q.dark.isActive
						? 'login/larepass_brand_desktop_dark.svg'
						: 'login/larepass_brand_desktop_light.svg'
				"
			/>
			<MobileTermipassUnlockContent
				v-else-if="isMobile"
				@unlockSuccess="onDialogOK"
				@cancel="onDialogCancel"
				:detail-text="
					info && info.length > 0
						? info
						: t('unlock.auth_popup_unlock_introduce')
				"
				:logo="
					$q.dark.isActive
						? 'login/larepass_brand_dark.svg'
						: 'login/larepass_brand.svg'
				"
			/>
		</q-card>
	</q-dialog>
</template>
<script setup lang="ts">
import { useDialogPluginComponent } from 'quasar';
import MobileTermipassUnlockContent from './mobile/TermipassUnlockContent.vue';
import DesktopTermipassUnlockContent from './desktop/TermipassUnlockContent.vue';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();
defineProps({
	info: {
		type: String,
		required: false,
		default: ''
	}
});
const isDesktop = ref(
	process.env.PLATFORM == 'DESKTOP' || process.env.APPLICATION_SUB_IS_BEX
);
const isMobile = ref(process.env.PLATFORM == 'MOBILE');

const { t } = useI18n();
</script>
<style lang="scss" scoped>
.d-creatVault {
	.q-dialog-plugin {
		width: 800px;
		height: 600px;
		border-radius: 12px;
	}
	.d-createVault-mobile {
		width: 100%;
		height: 100%;
	}
}
</style>
