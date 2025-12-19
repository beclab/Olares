<template>
	<div class="bind-vc-success-root">
		<div class="float-wrap">
			<div class="float" style="background: #93e1ff"></div>
			<div class="float" style="background: #faffd9"></div>
		</div>
		<terminus-wizard-view :btn-title="t('next')" @on-confirm="enterActive">
			<template v-slot:content>
				<div class="column items-center justify-center" style="height: 100%">
					<div class="bind-vc-success-root__image">
						<TerminusAvatar :info="userStore.terminusInfo()" :size="124" />
					</div>

					<div class="bind-vc-success-root__name text-h5 q-mt-xs">
						{{ userStore.current_user?.local_name }}
					</div>
					<div class="bind-vc-success-root__info text-body3 q-mt-xs">
						@{{ userStore.current_user?.domain_name }}
					</div>
					<div class="bind-vc-success-root__info text-body3 q-mt-xs">
						{{ t('congratulations_on_successfully_owning_a_olares_id') }}
					</div>
				</div>
			</template>
		</terminus-wizard-view>
	</div>
</template>

<script lang="ts" setup>
import TerminusWizardView from '../../../../components/common/TerminusWizardView.vue';
import { useI18n } from 'vue-i18n';
import { useUserStore } from '../../../../stores/user';
import { useRouter } from 'vue-router';
const userStore = useUserStore();
const { t } = useI18n();
const router = useRouter();

const enterActive = () => {
	router.replace({ path: '/connectLoading' });
};
</script>

<style lang="scss" scoped>
.bind-vc-success-root {
	width: 100%;
	height: 100%;
	position: relative;

	&__image {
		width: 124px;
		height: 124px;
		border-radius: 62px;
		overflow: hidden;
	}

	&__name {
		text-align: center;
		color: $ink-1;
	}

	&__info {
		text-align: center;
		color: $ink-2;
	}
}
</style>
