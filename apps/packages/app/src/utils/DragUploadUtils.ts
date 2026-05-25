/**
 * Drag-and-drop Upload Common Utility Class
 * Features:
 *    1. Bind drag-and-drop area
 *    2. Prevent browser default behaviors
 *    3. File validity filtering
 *    4. Expose upload callbacks
 * Applicable to: Vue/React/Vanilla JS projects, reusable across multiple projects
 */

export type FileFilterErrorReason =
	| 'illegal_name'
	| 'type_not_allowed'
	| 'size_exceeded'
	| 'count_exceeded';

export interface FileFilterErrorInfo {
	file: File;
	reason: FileFilterErrorReason;
	message: string;
}

const ERROR_MESSAGES: Record<
	FileFilterErrorReason,
	(file: File, options?: DragUploadOptions) => string
> = {
	illegal_name: (file) =>
		`File "${file.name}" contains invalid characters (\\ or /) and cannot be uploaded.`,
	type_not_allowed: (file, options) => {
		const allowedTypes = options?.allowedFileTypes?.join(', ') || 'unknown';
		return `File "${file.name}" is of an unsupported type. Only the following types are allowed: ${allowedTypes}.`;
	},
	size_exceeded: (file, options) => {
		const maxSize = options?.maxFileSizeMB || 0;
		const fileSize = (file.size / 1024 / 1024).toFixed(2);
		return `File "${file.name}" is ${fileSize}MB, which exceeds the maximum allowed size of ${maxSize}MB.`;
	},
	count_exceeded: (file, options) => {
		const maxCount = options?.maxFileCount || 0;
		return `File "${file.name}" exceeds the maximum file upload limit. You may upload up to ${maxCount} files.`;
	}
};

export interface ProcessedEventTarget {
	files: FileList;
	element: HTMLInputElement | undefined;
}

export interface ProcessedEvent extends Event {
	target: ProcessedEventTarget & EventTarget;
}

// Define the type of utility class configuration options
export interface DragUploadOptions {
	// DOM selector for the drag-and-drop area (e.g., "#drop-zone") or HTMLElement directly
	dropTarget: string | HTMLElement;
	// Optional: DOM selector for the file input (used to trigger traditional file upload, optional)
	fileInputSelector?: string;
	// Optional: Allowed file types for upload (e.g., ["image/*", ".pdf", ".doc"])
	allowedFileTypes?: string[];
	// Optional: Maximum size for a single file (unit: MB, no limit by default)
	maxFileSizeMB?: number;
	// Optional: Maximum number of files allowed (no limit by default)
	maxFileCount?: number;
	// Optional: Whether to filter file names containing invalid characters (\ or /) (true by default)
	filterIllegalFileName?: boolean;
	// Optional: Callback after file processing is completed (returns valid file list, only triggered when there are valid files)
	onFilesReady?: (files: FileList, originalFiles: FileList) => void;
	// Optional: Callback after target processing is completed (passed to the component's targetRef, only triggered when there are valid files)
	onTargetProcessed?: (
		processedEvent: ProcessedEvent,
		validFiles: FileList
	) => void;
	// Optional: Callback when dragging enters the area (used for visual feedback)
	onDragEnter?: () => void;
	// Optional: Callback when dragging leaves the area (used for visual feedback)
	onDragLeave?: () => void;
	// Optional: Callback when file filtering fails (returns invalid file information, including specific error messages)
	onFileFilterError?: (errorInfo: FileFilterErrorInfo) => void;
	// Optional: Callback when all files are filtered out (no valid files available for upload)
	onAllFilesFiltered?: (errors: FileFilterErrorInfo[]) => void;
}

export class DragUploadUtil {
	private dropTarget: HTMLElement | null;
	private fileInput?: HTMLInputElement;
	private options: DragUploadOptions;
	private dragCounter = 0; // Drag counter (solves the problem of repeated triggering by child elements)
	private isExternal = false; // Whether to disable drag-and-drop (for external pages)

	// Save references to event handlers (for correctly removing event listeners)
	private handleDragOver: (e: DragEvent) => void;
	private handleDragEnter: (e: DragEvent) => void;
	private handleDragLeave: (e: DragEvent) => void;
	private handleDrop: (e: DragEvent) => void;
	private handleFileInputChange: (e: Event) => void;

	constructor(options: DragUploadOptions) {
		if (!options.dropTarget) {
			throw new Error(
				'DragUploadUtil: Required configuration "dropTarget" cannot be empty'
			);
		}

		if (!options.onFilesReady && !options.onTargetProcessed) {
			throw new Error(
				'DragUploadUtil: Required configurations "onFilesReady" and "onTargetProcessed" cannot both be empty'
			);
		}

		this.options = {
			allowedFileTypes: [],
			maxFileSizeMB: Infinity,
			maxFileCount: Infinity,
			filterIllegalFileName: true,
			...options
		};

		// Resolve the drag-and-drop area DOM element
		this.dropTarget = this.resolveElement(this.options.dropTarget);
		if (!this.dropTarget) {
			throw new Error('DragUploadUtil: Drag-and-drop area DOM does not exist');
		}

		// Initialize event handlers (bind the this context)
		this.handleDragOver = this.onDragOver.bind(this);
		this.handleDragEnter = this.onDragEnter.bind(this);
		this.handleDragLeave = this.onDragLeave.bind(this);
		this.handleDrop = this.onDrop.bind(this);
		this.handleFileInputChange = this.onFileInputChange.bind(this);

		// Resolve the file input element (optional)
		if (this.options.fileInputSelector) {
			this.fileInput = document.querySelector(
				this.options.fileInputSelector
			) as HTMLInputElement;
			this.bindFileInputChange();
		}

		// Bind drag-and-drop related events
		this.bindDragEvents();
	}

	// Private method: Resolve DOM element (supports selector / direct HTMLElement)
	private resolveElement(target: string | HTMLElement): HTMLElement | null {
		if (typeof target === 'string') {
			return document.querySelector(target) as HTMLElement;
		}
		return target instanceof HTMLElement ? target : null;
	}

	// Private method: Prevent browser default drag-and-drop behaviors
	private preventDefault(e: Event): void {
		e.preventDefault();
		e.stopPropagation();
	}

	// Private method: Generate error information
	private createErrorInfo(
		file: File,
		reason: FileFilterErrorReason
	): FileFilterErrorInfo {
		const message = ERROR_MESSAGES[reason](file, this.options);
		return { file, reason, message };
	}

	// Event handler: dragover
	private onDragOver(e: DragEvent): void {
		this.preventDefault(e);
	}

	// Event handler: dragenter
	private onDragEnter(e: DragEvent): void {
		this.preventDefault(e);
		if (this.isExternal) return;

		this.dragCounter++;
		if (this.dragCounter === 1) {
			this.options.onDragEnter?.();
		}
	}

	// Event handler: dragleave
	private onDragLeave(e: DragEvent): void {
		this.preventDefault(e);
		if (this.isExternal) return;

		this.dragCounter--;
		if (this.dragCounter === 0) {
			this.options.onDragLeave?.();
		}
	}

	// Event handler: drop
	private onDrop(e: DragEvent): void {
		this.preventDefault(e);
		if (this.isExternal) return;

		// Reset counter and visual feedback
		this.dragCounter = 0;
		this.options.onDragLeave?.();

		// Get the dragged file list
		const originalFiles = e.dataTransfer?.files || new DataTransfer().files;
		if (originalFiles.length <= 0) return;

		// Filter valid files
		const { validFiles, errors } = this.filterFiles(originalFiles);

		// Do not trigger upload callbacks if there are no valid files
		if (validFiles.length === 0) {
			// Trigger callback when all files are filtered out
			this.options.onAllFilesFiltered?.(errors);
			return;
		}

		const processedEvent = new Event('change', { bubbles: true });
		// Override the read-only target property using Object.defineProperty
		Object.defineProperty(processedEvent, 'target', {
			value: {
				files: validFiles,
				element: this.fileInput // Saved file input element in the utility class
			},
			writable: false,
			configurable: true
		});

		// Call the additional callback to pass the processed object (for component to assign to targetRef)
		this.options.onTargetProcessed?.(
			processedEvent as ProcessedEvent,
			validFiles
		);

		// Optional chaining call: Avoid errors when onFilesReady is not configured
		this.options.onFilesReady?.(validFiles, originalFiles);
	}

	// Event handler: fileInput change
	private onFileInputChange(e: Event): void {
		const target = e.target as HTMLInputElement;
		const files = target.files || new DataTransfer().files;
		if (files.length > 0) {
			const { validFiles, errors } = this.filterFiles(files);

			// Do not trigger upload callbacks if there are no valid files
			if (validFiles.length === 0) {
				// Trigger callback when all files are filtered out
				this.options.onAllFilesFiltered?.(errors);
				// Reset the input to allow reselecting the same file
				target.value = '';
				return;
			}

			// Optional chaining call: Avoid errors when onFilesReady is not configured
			this.options.onFilesReady?.(validFiles, files);
		}
	}

	// Private method: Bind file input change event
	private bindFileInputChange(): void {
		if (!this.fileInput) return;
		this.fileInput.addEventListener('change', this.handleFileInputChange);
	}

	// Private method: Bind drag-and-drop related events
	private bindDragEvents(): void {
		if (!this.dropTarget) return;

		this.dropTarget.addEventListener('dragover', this.handleDragOver);
		this.dropTarget.addEventListener('dragenter', this.handleDragEnter);
		this.dropTarget.addEventListener('dragleave', this.handleDragLeave);
		this.dropTarget.addEventListener('drop', this.handleDrop);
	}

	// Private method: Get file extension (correctly handle files without extensions)
	private getFileExtension(fileName: string): string | null {
		const lastDotIndex = fileName.lastIndexOf('.');
		// Return null if there is no dot, or the dot is in the first position (e.g., .gitignore)
		if (lastDotIndex <= 0) {
			return null;
		}
		return `.${fileName.slice(lastDotIndex + 1).toLowerCase()}`;
	}

	// Private method: Check if the file type is allowed
	private isFileTypeAllowed(file: File): boolean {
		const allowedTypes = this.options.allowedFileTypes;
		if (!allowedTypes || allowedTypes.length === 0) {
			return true; // No type restrictions configured, allow all types by default
		}

		const fileExtension = this.getFileExtension(file.name);
		const fileMimeType = file.type;

		return allowedTypes.some((type) => {
			// Support wildcards (e.g., image/*)
			if (type.endsWith('/*')) {
				const mimePrefix = type.replace('/*', '/');
				return fileMimeType.startsWith(mimePrefix);
			}
			// Support file extensions (e.g., .pdf)
			if (type.startsWith('.')) {
				return fileExtension === type.toLowerCase();
			}
			// Support complete MIME types (e.g., application/pdf)
			return fileMimeType === type;
		});
	}

	// Core private method: File validity filtering (supports 4 filtering rules)
	private filterFiles(originalFiles: FileList): {
		validFiles: FileList;
		errors: FileFilterErrorInfo[];
	} {
		const dataTransfer = new DataTransfer();
		const maxFileSizeBytes = this.options.maxFileSizeMB! * 1024 * 1024;
		const maxFileCount = this.options.maxFileCount!;
		let validCount = 0;
		const errors: FileFilterErrorInfo[] = [];

		// Traverse all files and filter them one by one
		Array.from(originalFiles).forEach((file) => {
			let isLegal = true;
			let errorReason: FileFilterErrorReason | undefined;

			// Rule 1: Check if the file count exceeds the limit
			if (validCount >= maxFileCount) {
				isLegal = false;
				errorReason = 'count_exceeded';
			}

			// Rule 2: Filter file names containing invalid characters (\ or /)
			if (isLegal && this.options.filterIllegalFileName) {
				if (file.name.includes('\\') || file.name.includes('/')) {
					isLegal = false;
					errorReason = 'illegal_name';
				}
			}

			// Rule 3: Filter unsupported file types
			if (isLegal && !this.isFileTypeAllowed(file)) {
				isLegal = false;
				errorReason = 'type_not_allowed';
			}

			// Rule 4: Filter files that exceed the maximum size
			if (isLegal && file.size > maxFileSizeBytes) {
				isLegal = false;
				errorReason = 'size_exceeded';
			}

			// Valid file: Add to the return list
			if (isLegal) {
				dataTransfer.items.add(file);
				validCount++;
			} else {
				// Invalid file: Generate error information and trigger the callback
				const errorInfo = this.createErrorInfo(file, errorReason!);
				errors.push(errorInfo);
				this.options.onFileFilterError?.(errorInfo);
			}
		});

		return {
			validFiles: dataTransfer.files,
			errors
		};
	}

	/**
	 * Public method: Disable drag-and-drop upload functionality
	 */
	public disableUpload(): void {
		this.isExternal = true;
	}

	/**
	 * Public method: Enable drag-and-drop upload functionality
	 */
	public enableUpload(): void {
		this.isExternal = false;
	}

	/**
	 * Public method: Destroy the utility class instance (clean up event listeners and references)
	 */
	public destroy(): void {
		if (this.dropTarget) {
			this.dropTarget.removeEventListener('dragover', this.handleDragOver);
			this.dropTarget.removeEventListener('dragenter', this.handleDragEnter);
			this.dropTarget.removeEventListener('dragleave', this.handleDragLeave);
			this.dropTarget.removeEventListener('drop', this.handleDrop);
		}

		if (this.fileInput) {
			this.fileInput.removeEventListener('change', this.handleFileInputChange);
		}

		this.dropTarget = null;
		this.fileInput = undefined;
	}
}

/**
 * Factory function: Create an instance of DragUploadUtil
 * @param options Configuration options for DragUploadUtil
 * @returns Instance of DragUploadUtil
 */
export const createDragUpload = (options: DragUploadOptions) => {
	return new DragUploadUtil(options);
};
