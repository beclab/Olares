interface PromiseConstructor {
	allSettled<T extends readonly unknown[] | []>(
		promises: T
	): Promise<
		{
			[P in keyof T]: T[P] extends Promise<infer U>
				? { status: 'fulfilled'; value: U }
				: { status: 'fulfilled'; value: T[P] };
		}[number][]
	>;
}
