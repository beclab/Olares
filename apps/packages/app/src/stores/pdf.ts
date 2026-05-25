import { defineStore } from 'pinia';
import { bus } from '../utils/bus';

export interface PdfOutlineItem {
	title: string;
	page: number;
	level: number;
	children?: PdfOutlineItem[];
}

interface PdfState {
	numPages: number;
	pageNum: number;
	rotate: number;
	scale: number;
	source: any;
	topicLoad: boolean;
	pdfOutline: PdfOutlineItem[];
	pdfDocument: any; // PDF.js document reference for thumbnail generation
}

export const usePDfStore = defineStore('pdf', {
	state: () => {
		return {
			source: null,
			pageNum: 1,
			scale: 1,
			numPages: 0,
			rotate: 0,
			topicLoad: false,
			pdfOutline: [],
			pdfDocument: null
		} as PdfState;
	},

	actions: {
		init() {
			this.source = null;
			this.pageNum = 1;
			this.scale = 1;
			this.numPages = 0;
			this.rotate = 0;
			this.pdfOutline = [];
			this.pdfDocument = null;
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
		},

		setOutline(outline: PdfOutlineItem[]) {
			this.pdfOutline = outline;
		},

		clearOutline() {
			this.pdfOutline = [];
		},

		async extractOutline(pdf: any): Promise<void> {
			try {
				const outline = await pdf.getOutline();
				if (!outline || outline.length === 0) {
					this.pdfOutline = [];
					console.log('PDF Store: None extracted');
					return;
				}

				const processItems = async (
					items: any[],
					level = 0
				): Promise<PdfOutlineItem[]> => {
					const result: PdfOutlineItem[] = [];

					for (const item of items) {
						let pageNum = 1;

						if (item.dest) {
							try {
								let dest = item.dest;
								if (typeof dest === 'string') {
									dest = await pdf.getDestination(dest);
								}
								if (dest && dest[0]) {
									const pageIndex = await pdf.getPageIndex(dest[0]);
									pageNum = pageIndex + 1;
								}
							} catch (e) {
								console.warn('PDF Store: parse dest failed', item.title, e);
							}
						}

						const outlineItem: PdfOutlineItem = {
							title: item.title || 'Untitled',
							page: pageNum,
							level
						};

						if (item.items && item.items.length > 0) {
							outlineItem.children = await processItems(item.items, level + 1);
						}

						result.push(outlineItem);
					}

					return result;
				};

				this.pdfOutline = await processItems(outline);
				console.log(`PDF Store: process length ${this.pdfOutline.length}`);
			} catch (e) {
				console.warn('PDF Store: process topic failed', e);
				this.pdfOutline = [];
			}
		}
	}
});
