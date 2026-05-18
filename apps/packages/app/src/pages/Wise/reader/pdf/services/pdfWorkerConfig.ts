import * as pdfjsLib from 'pdfjs-dist';
import PdfWorker from 'pdfjs-dist/build/pdf.worker.entry';

let workerConfigured = false;

export function configurePdfWorker() {
	if (workerConfigured) {
		return;
	}

	pdfjsLib.GlobalWorkerOptions.workerSrc = PdfWorker;

	workerConfigured = true;

	console.log(`PDF.js Worker config:`);
	console.log(`  - PDF.js version: ${pdfjsLib.version}`);
	console.log(`  - Worker: local file`);
}
