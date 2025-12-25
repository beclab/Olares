export interface TwitterLoginPlugin {
	login(options: { oauthUrl: string }): Promise<{
		status: boolean;
		message: string;
	}>;
}
