// https://developer.android.com/reference/android/content/Intent#developer-guides
// https://developer.android.com/guide/components/intents-filters?hl=zh-cn
import { Action } from './action.interface';
import { Category } from './category.interface';
import { Uri, RouterID } from './base.interface';

/*
{
	app_id: 'desktop',
	entrance_id: '0',
	router_id : 'desktop',
	actions: [
		{
			action: 'view', 
			data: {type: 'video/*'},
			uri : '/video/:path'
		},
		{
			action: 'editor', 
			data: {type: 'txt'},
			uri : '/video/:path'
		},
		
		
	],	
}
*/
export class IntentFilter {
	id?: string;
	app_id: string;
	entrance_id: string;
	router_id: RouterID;

	actions: Action[] = [];
	categories: Category[] = [];
	data: Record<string, string> = {};

	constructor(props?: Partial<IntentFilter>) {
		props && Object.assign(this, props);
		this.actions = this.actions.sort();
		this.categories = this.categories.sort();
	}

	containAction(action: Action): boolean {
		if (this.actions.length === 0) {
			return false;
		}
		if (this.actions.findIndex((a) => a === action) >= 0) {
			return true;
		}
		return false;
	}

	equalActions(a: Action[]): boolean {
		const temp_actions = a.sort();
		if (temp_actions.length !== this.actions.length) {
			return false;
		}

		for (let i = 0; i < this.actions.length; i++) {
			if (temp_actions[i] !== this.actions[i]) {
				return false;
			}
		}

		return true;
	}

	containCategory(category: Category): boolean {
		if (this.categories.length === 0) {
			return false;
		}
		if (this.categories.findIndex((c) => c === category) >= 0) {
			return true;
		}
		return false;
	}

	equalCategories(a: Category[]): boolean {
		const temp_categories = a.sort();
		if (temp_categories.length !== this.categories.length) {
			return false;
		}

		for (let i = 0; i < this.categories.length; i++) {
			if (temp_categories[i] !== this.categories[i]) {
				return false;
			}
		}

		return true;
	}

	equalData(a: Record<string, string>): boolean {
		const key1 = Object.keys(this.data) as Array<string>;
		const key2 = Object.keys(a) as Array<string>;
		if (key1.length !== key2.length) {
			return false;
		}
		for (const key in key1) {
			if (this.data[key] !== a[key]) {
				return false;
			}
		}

		return true;
	}

	equal(b: IntentFilter): boolean {
		if (
			this.router_id !== b.router_id ||
			!this.equalActions(b.actions) ||
			!this.equalCategories(b.categories)
			//|| !this.equalData(b.data)
		) {
			return false;
		}

		return true;
	}

	// equalIntent(b: Intent): boolean {
	//   if (
	//     this.router_id !== b.router_id ||
	//     !this.equalActions([b.action]) ||
	//     !this.equalCategories([b.category]) ||
	//     !this.equalData(b.data)
	//   ) {
	//     return false;
	//   }

	//   return true;
	// }

	match(intent: Intent): boolean {
		if (intent.isExplicit()) {
			if (this.router_id !== intent.router_id) {
				return false;
			}
		}

		if (
			!this.containAction(intent.action) ||
			!this.containCategory(intent.category)
			//|| !this.equalData(intent.data)
		) {
			return false;
		}

		return true;
	}
}

export class Intent {
	router_id?: RouterID;
	uri?: Uri;
	action: Action;
	category: Category;
	data: Record<string, string> = {};
	extras: Record<string, string> = {};

	constructor(props?: Partial<Intent>) {
		props && Object.assign(this, props);
	}

	// setType(extra: string, data: string) {
	//   this.bundle[extra] = data;
	// }

	// putBundle(extra: string, data: string) {
	//   this.bundle[extra] = data;
	// }

	isExplicit(): boolean {
		if (this.router_id) {
			return true;
		}
		return false;
	}

	isImplicit(): boolean {
		return !this.isExplicit();
	}
}

export enum SendIntentResult {
	SENT = 'sent',
	SENT_WITH_DEFAULT_CHOICE = 'sent_with_default_choices',
	RETURN_FILTERS = 'return_filters',
	NO_EXACTLY_ONE_FILTER = 'error_no_exactly_one_filter',
	NO_MATCH_FILTER = 'error_no_match_filter'
}

export class SendIntentResponse {
	result: SendIntentResult;
	filters: IntentFilter[];
	constructor(props?: Partial<SendIntentResponse>) {
		props && Object.assign(this, props);
	}
}

export class CreateDefaultIntentChoiceRequest {
	intent: Intent;
	router_id: RouterID;

	constructor(props?: Partial<CreateDefaultIntentChoiceRequest>) {
		props && Object.assign(this, props);
	}
}

export class ReplaceDefaultIntentChoiceRequest {
	choice_id: string;
	router_id: RouterID;

	constructor(props?: Partial<ReplaceDefaultIntentChoiceRequest>) {
		props && Object.assign(this, props);
	}
}
