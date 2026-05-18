const styleMapping = {
	y: {
		t: 'top',
		m: 'marginTop',
		b: 'bottom'
	},
	x: {
		l: 'left',
		m: 'marginLeft',
		r: 'right'
	}
};

function addEvents(events) {
	events.forEach((cb, eventName) => {
		if (typeof document !== 'undefined') {
			document.documentElement.addEventListener(eventName, cb);
		}
	});
}

function removeEvents(events) {
	events.forEach((cb, eventName) => {
		if (typeof document !== 'undefined') {
			document.documentElement.removeEventListener(eventName, cb);
		}
	});
}

export default {
	name: 'vue-drag-resize',

	emits: [
		'clicked',
		'dragging',
		'dragstop',
		'resizing',
		'resizestop',
		'activated',
		'deactivated'
	],

	props: {
		stickSize: {
			type: Number,
			default: 8
		},
		parentScaleX: {
			type: Number,
			default: 1
		},
		parentScaleY: {
			type: Number,
			default: 1
		},
		isActive: {
			type: Boolean,
			default: false
		},
		preventActiveBehavior: {
			type: Boolean,
			default: false
		},
		isDraggable: {
			type: Boolean,
			default: true
		},
		isResizable: {
			type: Boolean,
			default: true
		},
		aspectRatio: {
			type: Boolean,
			default: false
		},
		parentLimitation: {
			type: Boolean,
			default: false
		},
		snapToGrid: {
			type: Boolean,
			default: false
		},
		gridX: {
			type: Number,
			default: 50,
			validator(val) {
				return val >= 0;
			}
		},
		gridY: {
			type: Number,
			default: 50,
			validator(val) {
				return val >= 0;
			}
		},
		parentW: {
			type: Number,
			default: 0,
			validator(val) {
				return val >= 0;
			}
		},
		parentH: {
			type: Number,
			default: 0,
			validator(val) {
				return val >= 0;
			}
		},
		w: {
			type: [String, Number],
			default: 200,
			validator(val) {
				return typeof val === 'string' ? val === 'auto' : val >= 0;
			}
		},
		h: {
			type: [String, Number],
			default: 200,
			validator(val) {
				return typeof val === 'string' ? val === 'auto' : val >= 0;
			}
		},
		minw: {
			type: Number,
			default: 400,
			validator(val) {
				return val >= 0;
			}
		},
		minh: {
			type: Number,
			default: 200,
			validator(val) {
				return val >= 0;
			}
		},
		x: {
			type: Number,
			default: 0,
			validator(val) {
				return typeof val === 'number';
			}
		},
		y: {
			type: Number,
			default: 0,
			validator(val) {
				return typeof val === 'number';
			}
		},
		z: {
			type: [String, Number],
			default: 'auto',
			validator(val) {
				return typeof val === 'string' ? val === 'auto' : val >= 0;
			}
		},
		dragHandle: {
			type: String,
			default: null
		},
		dragCancel: {
			type: String,
			default: null
		},
		sticks: {
			type: Array,
			default() {
				return ['tl', 'tm', 'tr', 'mr', 'br', 'bm', 'bl', 'ml'];
			}
		},
		axis: {
			type: String,
			default: 'both',
			validator(val) {
				return ['x', 'y', 'both', 'none'].indexOf(val) !== -1;
			}
		},
		contentClass: {
			type: String,
			required: false,
			default: ''
		}
	},

	data() {
		return {
			fixAspectRatio: null,
			active: null,
			zIndex: null,
			parentWidth: null,
			parentHeight: null,
			left: null,
			top: null,
			right: null,
			bottom: null,
			minHeight: null,
			_uid: null
		};
	},

	beforeCreate() {
		this.stickDrag = false;
		this.bodyDrag = false;
		this.dimensionsBeforeMove = {
			pointerX: 0,
			pointerY: 0,
			x: 0,
			y: 0,
			w: 0,
			h: 0
		};
		this.limits = {
			left: { min: null, max: null },
			right: { min: null, max: null },
			top: { min: null, max: null },
			bottom: { min: null, max: null }
		};

		this.currentStick = null;
	},

	async mounted() {
		this._uid = this.generateRandomId(8);
		this.parentElement = this.$el.parentNode;
		this.parentWidth = this.parentW
			? this.parentW
			: this.parentElement.clientWidth;
		this.parentHeight = this.parentH
			? this.parentH
			: this.parentElement.clientHeight;

		this.left = this.x;
		this.top = this.y;
		this.right =
			this.parentWidth -
			(this.w === 'auto' ? this.$refs.container.scrollWidth : this.w) -
			this.left;
		this.bottom =
			this.parentHeight -
			(this.h === 'auto' ? this.$refs.container.scrollHeight : this.h) -
			this.top;

		this.domEvents = new Map([
			['mousemove', this.move],
			['mouseup', this.up],
			['mouseleave', this.up],
			['mousedown', this.deselect],
			['touchmove', this.move],
			['touchend', this.up],
			['touchcancel', this.up],
			['touchstart', this.up]
		]);

		addEvents(this.domEvents);

		if (this.dragHandle) {
			[...this.$el.querySelectorAll(this.dragHandle)].forEach((dragHandle) => {
				dragHandle.setAttribute('data-drag-handle', this._uid);
			});
		}

		if (this.dragCancel) {
			[...this.$el.querySelectorAll(this.dragCancel)].forEach(
				(cancelHandle) => {
					cancelHandle.setAttribute('data-drag-cancel', this._uid);
				}
			);
		}
	},

	beforeUnmount() {
		removeEvents(this.domEvents);
	},

	methods: {
		generateRandomId(length) {
			let result = '';
			let characters =
				'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
			let charactersLength = characters.length;
			for (let i = 0; i < length; i++) {
				result += characters.charAt(
					Math.floor(Math.random() * charactersLength)
				);
			}
			return result;
		},

		deselect() {
			if (this.preventActiveBehavior) {
				return;
			}
			this.active = false;
		},

		move(ev) {
			if (!this.stickDrag && !this.bodyDrag) {
				return;
			}

			ev.stopPropagation();

			const pageX =
				typeof ev.pageX !== 'undefined' ? ev.pageX : ev.touches[0].pageX;
			const pageY =
				typeof ev.pageY !== 'undefined' ? ev.pageY : ev.touches[0].pageY;

			const { dimensionsBeforeMove } = this;

			const delta = {
				x: (dimensionsBeforeMove.pointerX - pageX) / this.parentScaleX,
				y: (dimensionsBeforeMove.pointerY - pageY) / this.parentScaleY
			};

			if (this.stickDrag) {
				this.stickMove(delta);
			}

			if (this.bodyDrag) {
				if (this.axis === 'x') {
					delta.y = 0;
				} else if (this.axis === 'y') {
					delta.x = 0;
				} else if (this.axis === 'none') {
					return;
				}
				this.bodyMove(delta);
			}
		},

		up(ev) {
			if (this.stickDrag) {
				this.stickUp(ev);
			} else if (this.bodyDrag) {
				this.bodyUp(ev);
			}
		},

		bodyDown(ev) {
			const { target, button } = ev;

			if (!this.preventActiveBehavior) {
				this.active = true;
			}

			if (button && button !== 0) {
				return;
			}

			this.$emit('clicked', ev);

			if (!this.active) {
				return;
			}

			if (
				this.dragHandle &&
				target.getAttribute('data-drag-handle') !== this._uid.toString()
			) {
				return;
			}

			if (
				this.dragCancel &&
				target.getAttribute('data-drag-cancel') === this._uid.toString()
			) {
				return;
			}

			if (typeof ev.stopPropagation !== 'undefined') {
				ev.stopPropagation();
			}

			if (typeof ev.preventDefault !== 'undefined') {
				ev.preventDefault();
			}

			if (this.isDraggable) {
				this.bodyDrag = true;
			}

			const pointerX =
				typeof ev.pageX !== 'undefined' ? ev.pageX : ev.touches[0].pageX;
			const pointerY =
				typeof ev.pageY !== 'undefined' ? ev.pageY : ev.touches[0].pageY;

			this.saveDimensionsBeforeMove({ pointerX, pointerY });

			if (this.parentLimitation) {
				this.limits = this.calcDragLimitation();
			}
		},

		bodyMove(delta) {
			const {
				dimensionsBeforeMove,
				parentWidth,
				parentHeight,
				gridX,
				gridY,
				width,
				height
			} = this;

			let newTop = dimensionsBeforeMove.top - delta.y;
			let newBottom = dimensionsBeforeMove.bottom + delta.y;
			let newLeft = dimensionsBeforeMove.left - delta.x;
			let newRight = dimensionsBeforeMove.right + delta.x;

			if (this.snapToGrid) {
				let alignTop = true;
				let alignLeft = true;

				let diffT = newTop - Math.floor(newTop / gridY) * gridY;
				let diffB =
					parentHeight -
					newBottom -
					Math.floor((parentHeight - newBottom) / gridY) * gridY;
				let diffL = newLeft - Math.floor(newLeft / gridX) * gridX;
				let diffR =
					parentWidth -
					newRight -
					Math.floor((parentWidth - newRight) / gridX) * gridX;

				if (diffT > gridY / 2) {
					diffT -= gridY;
				}
				if (diffB > gridY / 2) {
					diffB -= gridY;
				}
				if (diffL > gridX / 2) {
					diffL -= gridX;
				}
				if (diffR > gridX / 2) {
					diffR -= gridX;
				}

				if (Math.abs(diffB) < Math.abs(diffT)) {
					alignTop = false;
				}
				if (Math.abs(diffR) < Math.abs(diffL)) {
					alignLeft = false;
				}

				newTop -= alignTop ? diffT : diffB;
				newBottom = parentHeight - height - newTop;
				newLeft -= alignLeft ? diffL : diffR;
				newRight = parentWidth - width - newLeft;
			}

			({
				newLeft: this.left,
				newRight: this.right,
				newTop: this.top,
				newBottom: this.bottom
			} = this.rectCorrectionByLimit({
				newLeft,
				newRight,
				newTop,
				newBottom
			}));

			this.$emit('dragging', this.rect);
		},

		bodyUp() {
			this.bodyDrag = false;
			this.$emit('dragging', this.rect);
			this.$emit('dragstop', this.rect);

			this.dimensionsBeforeMove = {
				pointerX: 0,
				pointerY: 0,
				x: 0,
				y: 0,
				w: 0,
				h: 0
			};

			this.limits = {
				left: { min: null, max: null },
				right: { min: null, max: null },
				top: { min: null, max: null },
				bottom: { min: null, max: null }
			};
		},

		stickDown(stick, ev, force = false) {
			if ((!this.isResizable || !this.active) && !force) {
				return;
			}

			this.stickDrag = true;

			const pointerX =
				typeof ev.pageX !== 'undefined' ? ev.pageX : ev.touches[0].pageX;
			const pointerY =
				typeof ev.pageY !== 'undefined' ? ev.pageY : ev.touches[0].pageY;

			this.saveDimensionsBeforeMove({ pointerX, pointerY });

			this.currentStick = stick;

			this.limits = this.calcResizeLimits();
		},

		saveDimensionsBeforeMove({ pointerX, pointerY }) {
			this.dimensionsBeforeMove.pointerX = pointerX;
			this.dimensionsBeforeMove.pointerY = pointerY;

			this.dimensionsBeforeMove.left = this.left;
			this.dimensionsBeforeMove.right = this.right;
			this.dimensionsBeforeMove.top = this.top;
			this.dimensionsBeforeMove.bottom = this.bottom;

			this.dimensionsBeforeMove.width = this.width;
			this.dimensionsBeforeMove.height = this.height;

			this.aspectFactor = this.width / this.height;
		},

		stickMove(delta) {
			const {
				currentStick,
				dimensionsBeforeMove,
				gridY,
				gridX,
				snapToGrid,
				parentHeight,
				parentWidth
			} = this;

			let newTop = dimensionsBeforeMove.top;
			let newBottom = dimensionsBeforeMove.bottom;
			let newLeft = dimensionsBeforeMove.left;
			let newRight = dimensionsBeforeMove.right;

			switch (currentStick[0]) {
				case 'b':
					newBottom = dimensionsBeforeMove.bottom + delta.y;

					if (snapToGrid) {
						newBottom =
							parentHeight -
							Math.round((parentHeight - newBottom) / gridY) * gridY;
					}

					break;

				case 't':
					newTop = dimensionsBeforeMove.top - delta.y;

					if (snapToGrid) {
						newTop = Math.round(newTop / gridY) * gridY;
					}

					break;
				default:
					break;
			}

			switch (currentStick[1]) {
				case 'r':
					newRight = dimensionsBeforeMove.right + delta.x;

					if (snapToGrid) {
						newRight =
							parentWidth -
							Math.round((parentWidth - newRight) / gridX) * gridX;
					}

					break;

				case 'l':
					newLeft = dimensionsBeforeMove.left - delta.x;

					if (snapToGrid) {
						newLeft = Math.round(newLeft / gridX) * gridX;
					}

					break;
				default:
					break;
			}

			({ newLeft, newRight, newTop, newBottom } = this.rectCorrectionByLimit({
				newLeft,
				newRight,
				newTop,
				newBottom
			}));

			if (this.aspectRatio) {
				({ newLeft, newRight, newTop, newBottom } =
					this.rectCorrectionByAspectRatio({
						newLeft,
						newRight,
						newTop,
						newBottom
					}));
			}

			this.left = newLeft;
			this.right = newRight;
			this.top = newTop;
			this.bottom = newBottom;

			this.$emit('resizing', this.rect);
		},

		stickUp() {
			this.stickDrag = false;
			this.dimensionsBeforeMove = {
				pointerX: 0,
				pointerY: 0,
				x: 0,
				y: 0,
				w: 0,
				h: 0
			};
			this.limits = {
				left: { min: null, max: null },
				right: { min: null, max: null },
				top: { min: null, max: null },
				bottom: { min: null, max: null }
			};

			this.$emit('resizing', this.rect);
			this.$emit('resizestop', this.rect);
		},

		calcDragLimitation() {
			const { parentWidth, parentHeight } = this;

			return {
				left: { min: 0, max: parentWidth - this.width },
				right: { min: 0, max: parentWidth - this.width },
				top: { min: 0, max: parentHeight - this.height },
				bottom: { min: 0, max: parentHeight - this.height }
			};
		},

		calcResizeLimits() {
			const { aspectFactor, width, height, bottom, top, left, right } = this;
			let { minh: minHeight, minw: minWidth } = this;

			const parentLim = this.parentLimitation ? 0 : null;

			if (this.aspectRatio) {
				if (minWidth / minHeight > aspectFactor) {
					minHeight = minWidth / aspectFactor;
				} else {
					minWidth = aspectFactor * minHeight;
				}
			}

			const limits = {
				left: { min: parentLim, max: left + (width - minWidth) },
				right: { min: parentLim, max: right + (width - minWidth) },
				top: { min: parentLim, max: top + (height - minHeight) },
				bottom: { min: parentLim, max: bottom + (height - minHeight) }
			};

			if (this.aspectRatio) {
				const aspectLimits = {
					left: {
						min: left - Math.min(top, bottom) * aspectFactor * 2,
						max: left + ((height - minHeight) / 2) * aspectFactor * 2
					},
					right: {
						min: right - Math.min(top, bottom) * aspectFactor * 2,
						max: right + ((height - minHeight) / 2) * aspectFactor * 2
					},
					top: {
						min: top - (Math.min(left, right) / aspectFactor) * 2,
						max: top + ((width - minWidth) / 2 / aspectFactor) * 2
					},
					bottom: {
						min: bottom - (Math.min(left, right) / aspectFactor) * 2,
						max: bottom + ((width - minWidth) / 2 / aspectFactor) * 2
					}
				};

				if (this.currentStick[0] === 'm') {
					limits.left = {
						min: Math.max(limits.left.min, aspectLimits.left.min),
						max: Math.min(limits.left.max, aspectLimits.left.max)
					};
					limits.right = {
						min: Math.max(limits.right.min, aspectLimits.right.min),
						max: Math.min(limits.right.max, aspectLimits.right.max)
					};
				} else if (this.currentStick[1] === 'm') {
					limits.top = {
						min: Math.max(limits.top.min, aspectLimits.top.min),
						max: Math.min(limits.top.max, aspectLimits.top.max)
					};
					limits.bottom = {
						min: Math.max(limits.bottom.min, aspectLimits.bottom.min),
						max: Math.min(limits.bottom.max, aspectLimits.bottom.max)
					};
				}
			}

			return limits;
		},

		sideCorrectionByLimit(limit, current) {
			let value = current;

			if (limit.min !== null && current < limit.min) {
				value = limit.min;
			} else if (limit.max !== null && limit.max < current) {
				value = limit.max;
			}

			return value;
		},

		rectCorrectionByLimit(rect) {
			const { limits } = this;
			let { newRight, newLeft, newBottom, newTop } = rect;

			newLeft = this.sideCorrectionByLimit(limits.left, newLeft);
			newRight = this.sideCorrectionByLimit(limits.right, newRight);
			newTop = this.sideCorrectionByLimit(limits.top, newTop);
			newBottom = this.sideCorrectionByLimit(limits.bottom, newBottom);

			return {
				newLeft,
				newRight,
				newTop,
				newBottom
			};
		},

		rectCorrectionByAspectRatio(rect) {
			let { newLeft, newRight, newTop, newBottom } = rect;
			const {
				parentWidth,
				parentHeight,
				currentStick,
				aspectFactor,
				dimensionsBeforeMove
			} = this;

			let newWidth = parentWidth - newLeft - newRight;
			let newHeight = parentHeight - newTop - newBottom;

			if (currentStick[1] === 'm') {
				const deltaHeight = newHeight - dimensionsBeforeMove.height;

				newLeft -= (deltaHeight * aspectFactor) / 2;
				newRight -= (deltaHeight * aspectFactor) / 2;
			} else if (currentStick[0] === 'm') {
				const deltaWidth = newWidth - dimensionsBeforeMove.width;

				newTop -= deltaWidth / aspectFactor / 2;
				newBottom -= deltaWidth / aspectFactor / 2;
			} else if (newWidth / newHeight > aspectFactor) {
				newWidth = aspectFactor * newHeight;

				if (currentStick[1] === 'l') {
					newLeft = parentWidth - newRight - newWidth;
				} else {
					newRight = parentWidth - newLeft - newWidth;
				}
			} else {
				newHeight = newWidth / aspectFactor;

				if (currentStick[0] === 't') {
					newTop = parentHeight - newBottom - newHeight;
				} else {
					newBottom = parentHeight - newTop - newHeight;
				}
			}

			return { newLeft, newRight, newTop, newBottom };
		}
	},

	computed: {
		positionStyle() {
			return {
				top: this.top + 'px',
				left: this.left + 'px',
				zIndex: this.zIndex
			};
		},

		sizeStyle() {
			return {
				width: this.width + 'px',
				height: this.height + 'px'
			};
		},

		vdrStick() {
			return (stick) => {
				const stickStyle = {
					width: `${this.stickSize / this.parentScaleX}px`,
					height: `${this.stickSize / this.parentScaleY}px`
				};
				stickStyle[styleMapping.y[stick[0]]] = `${
					this.stickSize / this.parentScaleX / -2
				}px`;
				stickStyle[styleMapping.x[stick[1]]] = `${
					this.stickSize / this.parentScaleX / -2
				}px`;
				return stickStyle;
			};
		},

		width() {
			return this.parentWidth - this.left - this.right;
		},

		height() {
			return this.parentHeight - this.top - this.bottom;
		},

		rect() {
			return {
				left: Math.round(this.left),
				top: Math.round(this.top),
				width: Math.round(this.width),
				height: Math.round(this.height)
			};
		}
	},

	watch: {
		active(isActive) {
			if (isActive) {
				this.$emit('activated');
			} else {
				this.$emit('deactivated');
			}
		},

		isActive: {
			immediate: true,
			handler(val) {
				this.active = val;
			}
		},

		z: {
			immediate: true,
			handler(val) {
				if (val >= 0 || val === 'auto') {
					this.zIndex = val;
				}
			}
		},

		x: {
			handler(newVal, oldVal) {
				if (this.stickDrag || this.bodyDrag || newVal === this.left) {
					return;
				}

				const delta = oldVal - newVal;

				this.bodyDown({ pageX: this.left, pageY: this.top });
				this.bodyMove({ x: delta, y: 0 });

				this.$nextTick(() => {
					this.bodyUp();
				});
			}
		},

		y: {
			handler(newVal, oldVal) {
				if (this.stickDrag || this.bodyDrag || newVal === this.top) {
					return;
				}

				const delta = oldVal - newVal;

				this.bodyDown({ pageX: this.left, pageY: this.top });
				this.bodyMove({ x: 0, y: delta });

				this.$nextTick(() => {
					this.bodyUp();
				});
			}
		},

		w: {
			handler(newVal, oldVal) {
				if (this.stickDrag || this.bodyDrag || newVal === this.width) {
					return;
				}

				const stick = 'mr';
				const delta = oldVal - newVal;

				this.stickDown(
					stick,
					{ pageX: this.right, pageY: this.top + this.height / 2 },
					true
				);
				this.stickMove({ x: delta, y: 0 });

				this.$nextTick(() => {
					this.stickUp();
				});
			}
		},

		h: {
			handler(newVal, oldVal) {
				if (this.stickDrag || this.bodyDrag || newVal === this.height) {
					return;
				}

				const stick = 'bm';
				const delta = oldVal - newVal;

				this.stickDown(
					stick,
					{ pageX: this.left + this.width / 2, pageY: this.bottom },
					true
				);
				this.stickMove({ x: 0, y: delta });

				this.$nextTick(() => {
					this.stickUp();
				});
			}
		},

		parentW(val) {
			this.right = val - this.width - this.left;
			this.parentWidth = val;
		},

		parentH(val) {
			this.bottom = val - this.height - this.top;
			this.parentHeight = val;
		}
	}
};
