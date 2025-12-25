<template>
	<div class="monthpicker">
		<div class="monthpicker-header">
			<q-btn
				@click="changeYear(false)"
				dense
				:disable="
					minClean && minClean.getFullYear() === selectedMonth.getFullYear()
				"
				flat
				icon="navigate_before"
				round
			/>
			{{ selectedMonth.getFullYear() }}
			<q-btn
				@click="changeYear(true)"
				dense
				:disable="
					maxClean && maxClean.getFullYear() === selectedMonth.getFullYear()
				"
				flat
				icon="navigate_next"
				round
			/>
		</div>
		<div class="monthpicker-months">
			<q-btn
				:class="{ 'monthpicker-current': isCurrentMonth(month) }"
				@click="selectMonth(month, 'hide')"
				:color="isSelectedMonth(month) ? color : ''"
				:disable="isDisabled(month)"
				:flat="!isSelectedMonth(month)"
				:key="month.getTime()"
				:label="
					month.toLocaleDateString(localeArray, {
						month: 'short'
					})
				"
				no-caps
				no-outline
				no-ripple
				rounded
				:text-color="
					isCurrentMonth(month) && !isSelectedMonth(month) ? color : ''
				"
				v-for="month in displayedMonths"
			/>
		</div>
	</div>
</template>

<script>
export default {
	name: 'MonthPicker',
	computed: {
		displayedMonths() {
			let months = [];

			for (let i = 0; i < 12; i++) {
				months.push(new Date(this.selectedMonth.getFullYear(), i));
			}

			return months;
		},
		localeArray() {
			return this.locale ? [this.locale] : [];
		},
		maxClean() {
			return this.max ? this.cleanDate(this.max) : null;
		},
		minClean() {
			return this.min ? this.cleanDate(this.min) : null;
		}
	},
	data() {
		return {
			interval: null,
			selectedMonth: null
		};
	},
	unmounted() {
		clearInterval(this.interval);
	},
	methods: {
		changeYear(up) {
			if (up) {
				let d = new Date(this.selectedMonth);
				d.setMonth(d.getMonth() + 12);

				if (this.maxClean && d > this.maxClean) {
					d = this.maxClean;
				}

				this.selectMonth(d);
			} else {
				let d = new Date(this.selectedMonth);
				d.setMonth(d.getMonth() - 12);

				if (this.minClean && d < this.minClean) {
					d = this.minClean;
				}

				this.selectMonth(d);
			}
		},
		cleanDate(date) {
			let d = date ? new Date(date) : new Date();
			d.setDate(1);
			d.setHours(0, 0, 0, 0);
			return d;
		},
		currentMonth() {
			return this.cleanDate();
		},
		isCurrentMonth(month) {
			return this.currentMonth().getTime() === month.getTime();
		},
		isDisabled(month) {
			let disabled = false;

			if (
				(this.minClean && month < this.minClean) ||
				(this.maxClean && month > this.maxClean)
			) {
				disabled = true;
			}

			return disabled;
		},
		isSelectedMonth(month) {
			return this.selectedMonth
				? this.selectedMonth.getTime() === month.getTime()
				: false;
		},
		selectMonth(month, hide) {
			this.$emit('input', month, hide);
		}
	},
	props: ['color', 'locale', 'max', 'min', 'value'],
	watch: {
		value: {
			handler(val) {
				this.selectedMonth = val ? this.cleanDate(val) : this.currentMonth();
			},
			immediate: true
		}
	}
};
</script>

<style scoped lang="scss">
.monthpicker {
	width: 250px;
}

.monthpicker-current {
	font-weight: bold;
}

.monthpicker-header {
	align-items: center;
	display: flex;
	justify-content: space-between;
	margin-bottom: 8px;
}

.monthpicker-months {
	display: flex;
	justify-content: space-between;
	flex-wrap: wrap;
}

.monthpicker-months .q-btn {
	box-shadow: none;
	margin: 4px 0;
	width: 30%;
}
</style>
