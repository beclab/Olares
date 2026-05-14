import { reactive, InjectionKey } from 'vue';

export interface ColWidths {
	modified: number;
	type: number;
	size: number;
}

export type HandleKey = 'name' | 'size' | 'type';

export interface ColResizeContext {
	colWidths: ColWidths;
	cssVarsStyle: () => Record<string, string>;
	startResize: (handle: HandleKey, event: MouseEvent) => void;
}

export const COL_RESIZE_KEY: InjectionKey<ColResizeContext> =
	Symbol('colResize');

const STORAGE_KEY = 'files_listing_col_widths';
const DEFAULTS: ColWidths = { modified: 144, type: 80, size: 72 };
const MIN_COL_WIDTHS: ColWidths = {
	modified: 72,
	type: 64,
	size: 72
};
const NAME_MIN_WIDTH = 120;
const GRID_COLUMN_GAPS = 3 * 8;

function loadWidths(): ColWidths {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored) return { ...DEFAULTS, ...JSON.parse(stored) };
	} catch (e) {
		console.error(e);
	}
	return { ...DEFAULTS };
}

function saveWidths(widths: ColWidths) {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(widths));
	} catch (e) {
		console.error(e);
	}
}

// Per handle configuration:
//   shrinkOnRightDrag: when dragging right, shrink these columns in order
//   growOnRightDrag:   when dragging right, this fixed column grows
//                      (null means the name/1fr column grows automatically)
//   shrinkOnLeftDrag:  when dragging left, shrink these columns in order
//   growOnLeftDrag:    when dragging left, this column grows
//
// Visual column order is: name | size | type | modified
const HANDLE_CFG: Record<
	HandleKey,
	{
		growOnRightDrag: keyof ColWidths | null;
		shrinkOnRightDrag: (keyof ColWidths)[];
		growOnLeftDrag: keyof ColWidths;
		shrinkOnLeftDrag: (keyof ColWidths)[];
	}
> = {
	// name divider: right drag => name grows, then size/type/modified shrink in order
	//               left drag  => size grows, while name (1fr) shrinks automatically
	name: {
		growOnRightDrag: null,
		shrinkOnRightDrag: ['size', 'type', 'modified'],
		growOnLeftDrag: 'size',
		shrinkOnLeftDrag: []
	},
	// size divider: right drag => size grows, then type/modified shrink
	//               left drag  => type grows, then size shrinks
	//                             (remaining space comes from name/1fr)
	size: {
		growOnRightDrag: 'size',
		shrinkOnRightDrag: ['type', 'modified'],
		growOnLeftDrag: 'type',
		shrinkOnLeftDrag: ['size']
	},
	// type divider: right drag => type grows, modified shrinks
	//               left drag  => modified grows, then type/size shrink
	//                             (remaining space comes from name/1fr)
	type: {
		growOnRightDrag: 'type',
		shrinkOnRightDrag: ['modified'],
		growOnLeftDrag: 'modified',
		shrinkOnLeftDrag: ['type', 'size']
	}
};

export function useColumnResize(): ColResizeContext {
	const colWidths = reactive<ColWidths>(loadWidths());

	function cssVarsStyle(): Record<string, string> {
		return {
			'--col-width-modified': `${colWidths.modified}px`,
			'--col-width-type': `${colWidths.type}px`,
			'--col-width-size': `${colWidths.size}px`
		};
	}

	function startResize(handle: HandleKey, event: MouseEvent) {
		event.preventDefault();
		const startX = event.clientX;
		const startWidths = { ...colWidths };
		const target = event.target as HTMLElement | null;
		const headerItem = target?.closest('.item.header') as HTMLElement | null;
		const nameCell = headerItem?.querySelector('.name') as HTMLElement | null;
		const contentRow = headerItem?.querySelector(
			'div:last-of-type'
		) as HTMLElement | null;
		const startNameWidth =
			nameCell?.getBoundingClientRect().width || NAME_MIN_WIDTH;
		const contentWidth = contentRow?.getBoundingClientRect().width || 0;
		// Strict cap: name must keep at least NAME_MIN_WIDTH; the rest is for fixed cols + gaps.
		const maxFixedTotal = Math.max(
			0,
			contentWidth - NAME_MIN_WIDTH - GRID_COLUMN_GAPS
		);
		const startFixedTotal =
			startWidths.modified + startWidths.type + startWidths.size;
		const {
			growOnRightDrag,
			shrinkOnRightDrag,
			growOnLeftDrag,
			shrinkOnLeftDrag
		} = HANDLE_CFG[handle];

		const onMove = (e: MouseEvent) => {
			const delta = e.clientX - startX; // right = positive

			if (delta > 0) {
				// Dragging right: grow configured column, shrink right-side columns in cascade.
				// Do NOT cap with maxFixedTotal - startFixedTotal here: that cap is for *increasing*
				// the sum of fixed columns (left drag). Right drag *redistributes* or *shrinks*
				// fixed cols toward name/1fr, so when startFixedTotal === maxFixedTotal we must
				// still allow shrinking fixed cols (otherwise drag is stuck after pushing left).
				const totalAvail = shrinkOnRightDrag.reduce(
					(s, col) => s + Math.max(0, startWidths[col] - MIN_COL_WIDTHS[col]),
					0
				);
				const actualGrow =
					shrinkOnRightDrag.length > 0
						? Math.min(delta, totalAvail)
						: Math.min(delta, Math.max(0, maxFixedTotal - startFixedTotal));

				if (growOnRightDrag) {
					colWidths[growOnRightDrag] =
						startWidths[growOnRightDrag] + actualGrow;
				}
				// else: name/1fr grows automatically

				// Cascade shrink on configured right-side columns.
				let remaining = actualGrow;
				for (const col of shrinkOnRightDrag) {
					if (remaining <= 0) break;
					const avail = Math.max(0, startWidths[col] - MIN_COL_WIDTHS[col]);
					const take = Math.min(remaining, avail);
					colWidths[col] = startWidths[col] - take;
					remaining -= take;
				}
			} else if (delta < 0) {
				// Dragging left: grow configured right-side column, shrink left-side columns.
				// If left fixed columns are insufficient, name/1fr can provide remaining space
				// until it reaches NAME_MIN_WIDTH. Growth is capped to avoid total-width overflow.
				const absDelta = -delta;
				const nameAvail = Math.max(0, startNameWidth - NAME_MIN_WIDTH);
				const byTotalCap = Math.max(0, maxFixedTotal - startFixedTotal);

				let fixedAvail = 0;
				for (const col of shrinkOnLeftDrag) {
					fixedAvail += Math.max(0, startWidths[col] - MIN_COL_WIDTHS[col]);
				}
				// Net fixed-total growth = max(0, actualGrow - fixedAvail), capped by byTotalCap;
				// also capped by nameAvail (name/1fr cannot shrink below NAME_MIN_WIDTH).
				const headroom = Math.min(nameAvail, byTotalCap);
				const actualGrow = Math.min(absDelta, fixedAvail + headroom);

				let remaining = actualGrow;
				for (const col of shrinkOnLeftDrag) {
					if (remaining <= 0) break;
					const avail = Math.max(0, startWidths[col] - MIN_COL_WIDTHS[col]);
					const take = Math.min(remaining, avail);
					colWidths[col] = startWidths[col] - take;
					remaining -= take;
				}

				// Name/1fr implicitly absorbs any remaining part (if any), but never below NAME_MIN_WIDTH.
				colWidths[growOnLeftDrag] = startWidths[growOnLeftDrag] + actualGrow;
			}
		};

		const onUp = () => {
			saveWidths({ ...colWidths });
			window.removeEventListener('mousemove', onMove);
			window.removeEventListener('mouseup', onUp);
		};

		window.addEventListener('mousemove', onMove);
		window.addEventListener('mouseup', onUp);
	}

	return { colWidths, cssVarsStyle, startResize };
}
