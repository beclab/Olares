export class User {
	public id = '';
	public name = '';
	public createdAt = '';
	public constructor(init?: Partial<User>) {
		Object.assign(this, init);
	}
}
