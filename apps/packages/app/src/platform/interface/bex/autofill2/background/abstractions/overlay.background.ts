import AutofillPageDetails from '../../models/autofill-page-details';

type OverlayAddNewItemMessage = {
	login?: {
		uri?: string;
		hostname: string;
		username: string;
		password: string;
	};
};

type OverlayBackgroundExtensionMessage = {
	[key: string]: any;
	// type: string;
	tab?: chrome.tabs.Tab;
	sender?: string;
	details?: AutofillPageDetails;
	overlayElement?: string;
	display?: string;
} & OverlayAddNewItemMessage;

export { OverlayBackgroundExtensionMessage, OverlayAddNewItemMessage };
