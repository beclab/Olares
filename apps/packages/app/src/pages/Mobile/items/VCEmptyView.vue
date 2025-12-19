<template>
	<div class="vc-empty column justify-center items-center q-pb-md">
		<q-img
			class="bind-terminus-vc-page__image"
			:src="getRequireImage('login/create_terminus_name.svg')"
		/>
		<span class="bind-terminus-vc-page__desc login-sub-title">{{
			t(
				'Bind a verification credential (VC) via your social accounts to create an Olares ID'
			)
		}}</span>
		<TerminusEnterBtn
			:title="t('bind_vc')"
			icon="sym_r_link"
			@sureAction="goBind"
			class="q-mt-md"
		/>
	</div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { getRequireImage } from '../../../utils/imageUtils';
import TerminusEnterBtn from '../../../components/common/TerminusEnterBtn.vue';

const { t } = useI18n();
const router = useRouter();

const goBind = () => {
	router.push({
		path: '/addVC'
	});
};
</script>

<style scoped lang="scss">
.vc-empty {
	width: 100%;
	height: 100%;

	.bind-terminus-vc-page {
		width: 100%;
		height: 100%;
		padding-left: 20px;
		padding-right: 20px;

		&__image {
			width: 160px;
			height: 160px;
		}

		&__desc {
			width: calc(100% - 40px);
			margin-top: 20px;
		}
	}
}
</style>
