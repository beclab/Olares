class MonitorKeyboard {
	type: number | undefined;
	originalHeight: number;
	show: any;
	hidden: any;

	constructor() {
		this.type = this.IsIA();
		this.originalHeight = window.innerHeight;
	}

	IsIA = () => {
		const userAgent =
			typeof window === 'object' ? window.navigator.userAgent : '';
		if (/android/i.test(userAgent)) {
			return 1;
		} else if (/iPhone|iPod|iPad/i.test(userAgent)) {
			return 2;
		}
	};

	onResize = () => {
		const resizeHeight = window.innerHeight;
		if (this.originalHeight - resizeHeight > 50) {
			this.show('Android system show keyboard');
		} else {
			this.hidden('Android system close keyboard');
		}
	};

	// iOS获取焦点
	onFocusin = () => {
		this.show('iOS system show keyboard');
	};

	// iOS失去焦点
	onFocusout = () => {
		this.hidden('iOS system hidden keyboard');
	};

	onStart = () => {
		if (this.type == 1) {
			window.addEventListener('resize', this.onResize);
		}
		if (this.type == 2) {
			window.addEventListener('focusin', this.onFocusin);
			window.addEventListener('focusout', this.onFocusout);
		}
	};

	onEnd = () => {
		if (this.type == 1) {
			window.removeEventListener('resize', this.onResize);
		}
		if (this.type == 2) {
			window.removeEventListener('focusin', this.onFocusin);
			window.removeEventListener('focusout', this.onFocusout);
		}
	};

	onShow = (fn: () => void) => {
		this.show = fn;
	};

	onHidden = (fn: () => void) => {
		this.hidden = fn;
	};
}

export default MonitorKeyboard;
