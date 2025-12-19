<template>
	<div class="my-timeline-container q-gutter-y-md">
		<MyExpand2 title="Highlights">
			<template #extra>
				<q-icon name="sym_r_content_copy" size="18px" color="grey-8" />
			</template>
			<bt-scroll-area style="height: 300px">
				<div style="margin-top: -12px">
					<q-timeline color="grey-2" layout="dense">
						<q-timeline-entry
							v-for="item in 5"
							:key="item"
							style="margin-bottom: -4px"
						>
							<div class="avatar-wrapper">20:32</div>
							<div class="timeline-content">
								<div class="text-body2 text-grey-10">
									Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed
									do eiusmod tempor incididunt ut labore et dolore magna aliqua.
									Ut enim ad minim veniam, quis nostrud exercitation ullamco
									laboris nisi ut aliquip ex ea commodo consequat.
								</div>
								<MyExpand>
									<div class="detail-wrapper text-body3 text-grey-10 q-mt-sm">
										<ul class="ul-wrapper">
											<li>test</li>
											<li>test</li>
											<li>test</li>
											<li>test</li>
										</ul>
									</div>
								</MyExpand>
							</div>
						</q-timeline-entry>
					</q-timeline>
				</div>
			</bt-scroll-area>
		</MyExpand2>
	</div>
</template>

<script setup lang="ts">
import MyExpand from './MyExpand.vue';
import MyExpand2 from './MyExpand2.vue';
</script>

<style lang="scss" scoped>
.my-timeline-container {
	.timeline-content {
		padding-top: 27px;
	}
	.avatar-wrapper {
		position: absolute;
		top: -4px;
		left: -1px;
		border-radius: 6px;
		background: $yellow-6;
		padding: 4px 8px;
		line-height: 12px;
	}
	.detail-wrapper {
		border-left: 1px solid $separator;
		.ul-wrapper {
			padding-inline-start: 20px;
			list-style-type: disc;
			color: $grey-10;
		}
	}
	::v-deep(.q-timeline__dot:before) {
		display: none;
	}
	::v-deep(.q-timeline__content) {
		padding-bottom: 0px;
	}
	::v-deep(.q-timeline--dense--right .q-timeline__entry) {
		padding-left: 24px;
	}
}
</style>
