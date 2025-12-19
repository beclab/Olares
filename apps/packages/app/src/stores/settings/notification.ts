import { defineStore } from 'pinia';
import axios from 'axios';

type NotifyPolicy = any;
type Sender = any;
type Template = any;
type TemplateContent = any;
type NotifyRule = any;
type Recipients = any;
type RecipientAddress = any;
type Job = any;
type Message = any;

export type NotificationState = {
	url: string;
	get_resourced: boolean;
	senders: Sender[];
	recipients: Recipients[];
	recipientAddress: RecipientAddress[];
	notifyPolicy: NotifyPolicy[];
	//notifyRule: NotifyRule[];
	templates: Template[];
	templateContent: TemplateContent[];
	jobs: Job[];
};

export const useNotificationStore = defineStore('notification', {
	state: () => {
		return {
			url: '',
			get_resourced: false,
			senders: [],
			recipients: [],
			recipientAddress: [],
			templates: [],
			templateContent: [],
			notifyPolicy: [],
			jobs: []
		} as NotificationState;
	},
	getters: {},
	actions: {
		async init() {
			await this.refreshSenders();

			await this.refreshRecipients();

			await this.refreshRecipientAddress();

			await this.refreshNotifyPolicy();

			//await this.refreshNotifyRule();

			await this.refreshTemplate();

			await this.refreshTemplateContent();

			await this.refreshJobs();

			//await this.refreshMessage();
		},

		setUrl(new_url: string) {
			this.url = new_url;
		},

		async refreshSenders() {
			this.senders = await axios.get(this.url + '/notification/sender', {});
		},
		async createSender(sender: Sender, credential: Record<string, string>) {
			await axios.post(this.url + '/sender', {
				sender,
				credential
			});
		},

		async refreshRecipients() {
			this.recipients = await axios.get(
				this.url + '/notification/recipients',
				{}
			);
		},
		async createRecipients(recipients: Recipients) {
			await axios.post(this.url + '/notification/recipients', {
				recipients
			});
		},

		async refreshRecipientAddress() {
			this.recipientAddress = await axios.get(
				this.url + '/notification/recipientAddress',
				{}
			);
		},
		async createRecipientsAddress(
			id: number,
			name: string,
			data: Record<string, string>
		) {
			await axios.post(this.url + '/notification/recipientAddress', {
				id,
				name,
				data
			});
		},
		async refreshTemplate() {
			this.templates = await axios.get(this.url + '/notification/template', {});
		},
		async refreshTemplateContent() {
			this.templateContent = await axios.get(
				this.url + '/notification/templateContent',
				{}
			);
		},
		async refreshNotifyPolicy() {
			this.notifyPolicy = await axios.get(
				this.url + '/notification/notifyPolicy',
				{}
			);
		},
		async createNotifyPolicy(policy: NotifyPolicy) {
			await axios.post(this.url + '/notification/notifyPolicy', {
				policy
			});
		},
		async getNotifyRuleByPolicyId(id: number): Promise<NotifyRule[]> {
			return await axios.get(this.url + '/notification/notifyRule/' + id);
		},
		async createNotifyRule(rule: NotifyRule) {
			await axios.post(this.url + '/notification/notifyRule', { rule });
		},
		async deleteNotifyRule(id: number) {
			await axios.delete(this.url + '/notification/notifyRule/' + id);
		},
		async refreshJobs() {
			this.jobs = await axios.get(this.url + '/notification/job', {});
		},
		async createJob(job: Job) {
			await axios.post(this.url + '/notification/job', {
				job
			});
		},
		async getMessageByJob(jobId: number): Promise<Message[]> {
			return await axios.get(
				this.url + '/notification/job/message/' + jobId,
				{}
			);
		}
	}
});
