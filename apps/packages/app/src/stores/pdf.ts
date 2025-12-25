import { defineStore } from 'pinia';
import { bus } from '../utils/bus';

interface PdfState {
	source: string;
	pageNum: number;
	scale: number;
	numPages: number;
	rotate: number;
	topicLoad: boolean;
}

export const usePDfStore = defineStore('pdf', {
	state: () => {
		return {
			source: '',
			pageNum: 1,
			scale: 1,
			numPages: 0,
			rotate: 0,
			topicLoad: false
		} as PdfState;
	},

	actions: {
		init() {
			this.source = '';
			this.pageNum = 1;
			this.scale = 1;
			this.numPages = 0;
			this.rotate = 0;
		},
		lastPage() {
			if (this.pageNum > 1) {
				this.pageNum -= 1;
				bus.emit('scrollIntoPos');
			}
		},

		skipPage(index: number) {
			if (index <= 0) {
				this.pageNum = 1;
			} else if (index >= this.numPages) {
				this.pageNum = this.numPages;
			} else {
				this.pageNum = index;
			}
			bus.emit('scrollIntoPos');
		},

		handleItemClick(number) {
			this.pageNum = number;
			bus.emit('scrollIntoPos');
		},

		nextPage() {
			if (this.pageNum < this.numPages) {
				this.pageNum += 1;
				bus.emit('scrollIntoPos');
			}
		},

		pageZoomIn() {
			if (this.scale < 2) {
				this.scale += 0.1;
			}
		},

		pageZoomOut() {
			if (this.scale > 1) {
				this.scale -= 0.1;
			}
		},

		pageRotate() {
			if (this.rotate === 360) {
				this.rotate = 0;
			} else {
				this.rotate += 90;
			}
		},

		pageCounterRotate() {
			if (this.rotate === 0) {
				this.rotate = 270;
			} else {
				this.rotate -= 90;
			}
		}
	}
});
