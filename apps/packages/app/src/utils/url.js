export function removeLastDir(url) {
	var arr = url.split('/');
	if (arr.pop() === '') {
		arr.pop();
	}
	return arr.join('/');
}

// this code borrow from mozilla
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/encodeURIComponent#Examples
export function encodeRFC5987ValueChars(str) {
	return (
		encodeURIComponent(str)
			// Note that although RFC3986 reserves "!", RFC5987 does not,
			// so we do not need to escape it
			.replace(/['()]/g, escape) // i.e., %27 %28 %29
			.replace(/\*/g, '%2A')
			// The following are not required for percent-encoding per RFC5987,
			// so we can allow for a little better readability over the wire: |`^
			.replace(/%(?:7C|60|5E)/g, unescape)
	);
}

export function encodePath(str) {
	return str
		.split('/')
		.map((v) => encodeURIComponent(v))
		.join('/');
}

const getSearchParamsObj = () => {
	const url = new URL(document.location.href);
	return Object.fromEntries(url.searchParams.entries());
};

const getWindowPathname = () => {
	if (window.location.hash && window.location.hash.length > 0) {
		const pathWithQuery = window.location.hash.startsWith('#')
			? window.location.hash.substring(1)
			: window.location.hash;
		const [path] = pathWithQuery.split('?');
		return path;
	}
	return window.location.pathname;
};

const getWindowFullpath = () => {
	if (window.location.hash && window.location.hash.length > 0) {
		const pathWithQuery = window.location.hash.startsWith('#')
			? window.location.hash.substring(1)
			: window.location.hash;
		return pathWithQuery;
	}
	return window.location.pathname + window.location.search;
};

export default {
	encodeRFC5987ValueChars,
	removeLastDir,
	encodePath,
	getSearchParamsObj,
	getWindowPathname,
	getWindowFullpath
};
