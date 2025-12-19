class NativeInputBlurMonitor {
	previousHeight: number;
	blur?: () => void;
	isFocus: boolean;

	constructor() {
		this.isFocus = false;
	}

	onResize = () => {
		if (window.visualViewport) {
			const currentHeight = window.visualViewport.height;
			if (Math.abs(currentHeight - this.previousHeight) < 1 && this.isFocus) {
				setTimeout(() => {
					if (this.blur) {
						this.blur();
						this.isFocus = false;
					}
				}, 100);
			}
		}
	};

	onStart = () => {
		if (window.visualViewport) {
			window.visualViewport.addEventListener('resize', this.onResize);
			this.previousHeight = window.visualViewport.height;
		}
	};

	onEnd = () => {
		if (window.visualViewport) {
			window.visualViewport.removeEventListener('resize', this.onResize);
		}
	};
}

export default NativeInputBlurMonitor;
