export class QueryParamsBuilder {
	offset = 0;
	limit = 20;
	params: Record<string, string | number | boolean> = {};

	constructor(props?: Partial<QueryParamsBuilder>) {
		if (props) {
			if (props.params) {
				this.params = { ...this.params, ...props.params };
			}
			Object.assign(this, props);
		}
	}

	public build(): string {
		const queryParts: string[] = [];

		if (this.offset !== undefined) {
			queryParts.push(`offset=${this.offset}`);
		}
		if (this.limit !== undefined) {
			queryParts.push(`limit=${this.limit}`);
		}

		for (const key in this.params) {
			if (this.params[key] !== undefined) {
				queryParts.push(`${key}=${encodeURIComponent(this.params[key])}`);
			}
		}

		return queryParts.length > 0 ? `?${queryParts.join('&')}` : '';
	}
}
