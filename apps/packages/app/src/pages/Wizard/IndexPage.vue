<template>
	<div class="boot_dialog">
		<div class="dialog_box_stepper">
			<div class="float-wrap" v-if="!isWelcome">
				<div
					class="float"
					:style="{
						background: `${
							tokenStore.step && tokenStore.step < 6
								? FloatBg[tokenStore.step].color1
								: '#93e1ff'
						}`
					}"
				></div>
				<div
					class="float"
					:style="{
						background: `${
							tokenStore.step && tokenStore.step < 6
								? FloatBg[tokenStore.step].color2
								: '#faffd9'
						}`
					}"
				></div>
			</div>

			<div
				class="welcome-wrap fade-in"
				v-if="isWelcome"
				@click="isWelcome = false"
			>
				<div class="blur"></div>
				<div class="blur"></div>
				<div class="blur"></div>
				<div class="welcome">
					<div class="outer">
						<h1 class="text boujee-text">{{ t('wizard_welcome') }}</h1>
						<div class="inner">
							<h1 class="text boujee-text">{{ t('wizard_welcome') }}</h1>
						</div>
					</div>
					<p class="boujee-desc">{{ t('wizard_continue') }}</p>
				</div>
			</div>
			<div class="fade-in stepper_step_itme" v-else>
				<q-stepper
					v-model="tokenStore.wizard.step"
					ref="stepper"
					color="primary"
					header-class="text_bold_header"
					class="step_height"
					animated
				>
					<q-step :name="1" :done="tokenStore.step > 1" title="">
						<step1-page ref="step1PageRef" />
					</q-step>
					<q-step :name="2" :done="tokenStore.step > 2" title="">
						<step2-page ref="step2PageRef" />
					</q-step>
					<q-step :name="3" :done="tokenStore.step > 3" title="">
						<step3-page ref="step3PageRef" />
					</q-step>
					<q-step :name="4" :done="tokenStore.step > 4" title="">
						<step4-page ref="step4PageRef" />
					</q-step>

					<template v-slot:navigation>
						<q-stepper-navigation class="navigation_Continue">
							<div class="Continue_bottom" v-if="tokenStore.step < 4">
								<div
									class="refush row items-center justify-center"
									v-if="tokenStore.raw_login_loading"
								>
									<q-img
										src="../../assets/wizard-web/progress_activity.svg"
										spinner-color="white"
										style="width: 18px; height: 18px"
									/>
								</div>
								<q-btn
									:disable="tokenStore.step > 5"
									v-if="tokenStore.step > 2"
									:ripple="false"
									round
									icon="sym_r_arrow_left_alt"
									@click="stepperPrevious"
									:class="[
										tokenStore.step > 1
											? 'stepper_btn_auto previous_btn'
											: 'previous_btn'
									]"
									:style="tokenStore.step > 4 ? 'background: #F7F7F7;' : ''"
								></q-btn>
								<q-btn
									:ripple="false"
									@click="stepperNext"
									:disable="
										tokenStore.step == 3
											? tokenStore.frpList.length == 0
											: false
									"
									no-caps
									:label="t('step_continue')"
									:class="[stepperNextCalss]"
								/>
							</div>
						</q-stepper-navigation>
					</template>
				</q-stepper>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue';
import gsap from 'gsap';
import { useI18n } from 'vue-i18n';

import Step1Page from './step/Step1Page.vue';
import Step2Page from './step/Step2Page.vue';
import Step3Page from './step/Step3Page.vue';
import Step4Page from './step/Step4Page.vue';

import { useTokenStore } from '../../stores/wizard-step';
import { FloatBg } from '../../utils/constants';

const { t } = useI18n();
const stepper = ref();
const tokenStore = useTokenStore();

let isWelcome = ref(true);

const step1PageRef = ref();
const step2PageRef = ref();
const step3PageRef = ref();
const step4PageRef = ref();

function stepperNext() {
	if (tokenStore.step === 1) {
		step1PageRef.value.click();
	} else if (tokenStore.step === 2) {
		step2PageRef.value.click();
	} else if (tokenStore.step === 3) {
		step3PageRef.value.click();
	} else if (tokenStore.step === 4) {
		step4PageRef.value.click();
	}
}

function stepperPrevious() {
	stepper.value.previous();
}

const randomX = random(-400, 400);
const randomY = random(-200, 200);
//const randomDelay = random(0, 50);
const randomTime = random(6, 12);
const randomTime2 = random(5, 6);
const randomAngle = random(-30, 150);
const ease = 'elastic';

function rotate(target: any, direction: any) {
	return gsap.to(target, randomTime2(), {
		rotation: randomAngle(direction),
		// delay: randomDelay(),
		duration: 0.1,
		ease: ease,
		onComplete: rotate,
		onCompleteParams: [target, direction * -1]
	});
}

function moveX(target: any, direction: any) {
	return gsap.to(target, randomTime(), {
		x: randomX(direction),
		ease: ease,
		duration: 0.1,
		onComplete: moveX,
		onCompleteParams: [target, direction * -1]
	});
}

function moveY(target: any, direction: any) {
	return gsap.to(target, randomTime(), {
		y: randomY(direction),
		ease: 'none',
		duration: 0.1,
		onComplete: moveY,
		onCompleteParams: [target, direction * -1]
	});
}

function random(min: number, max: number) {
	const delta = max - min;
	return (direction = 1) => (min + delta * Math.random()) * direction;
}

const stepperNextCalss = computed(() => {
	if (tokenStore.raw_login_loading) {
		return 'stepper_btn_disable stepper_btn_left';
	}
	if (tokenStore.step == 1 && tokenStore.wizard.password) {
		return 'stepper_btn stepper_btn_left stepper_btn_auto';
	}

	if (tokenStore.step == 5 || !tokenStore.wizard.password) {
		return 'stepper_btn_disable stepper_btn_left';
	}

	return 'stepper_btn stepper_btn_left';
});

onMounted(() => {
	const blurs: any = gsap.utils.toArray('.blur');
	blurs.forEach((blur: string) => {
		gsap.set(blur, {
			x: randomX(-1),
			y: randomX(1),
			rotation: randomAngle(-1)
		});
		moveX(blur, 1);
		moveY(blur, -1);
		rotate(blur, 1);
	});

	document.addEventListener('keydown', function (event: any) {
		if (isWelcome.value) isWelcome.value = false;
		if (event.keyCode === 13) {
			if (tokenStore.step < 5) {
				stepperNext();
			}
		}
	});
});
</script>
<style lang="scss"></style>
