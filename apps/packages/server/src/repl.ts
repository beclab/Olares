// import { StorageListOptions } from "@padloc/core/src/storage";
// import { Server } from "@padloc/core/src/server";
// import { Account } from "@padloc/core/src/account";
// import { Session } from "@padloc/core/src/session";
// import { Org, OrgRole } from "@padloc/core/src/org";
// import { Serializable } from "@padloc/core/src/encoding";
// import { Vault } from "@padloc/core/src/vault";
// import { PlanType, Subscription, SubscriptionStatus, UpdateBillingParams } from "@padloc/core/src/billing";
// import { ListLogEventsOptions } from "@padloc/core/src/logging";
// import * as colors from "ansi-colors";
// import { format } from "date-fns";
// import repl from "repl";
// import net from "net";

// function planColor(plan: PlanType | undefined) {
//     switch (plan) {
//         case PlanType.Free:
//             return "gray";
//         case PlanType.Premium:
//             return "red";
//         case PlanType.Family:
//             return "yellow";
//         case PlanType.Team:
//             return "green";
//         case PlanType.Business:
//             return "blue";
//         default:
//             return "white";
//     }
// }

// function subColor(sub: Subscription | undefined | null) {
//     const status = sub && sub.status;
//     switch (status) {
//         case SubscriptionStatus.Trialing:
//             return "yellow";
//         case SubscriptionStatus.Active:
//             return "green";
//         case SubscriptionStatus.Inactive:
//             return "red";
//         case SubscriptionStatus.Canceled:
//             return "dim";
//         default:
//             return "white";
//     }
// }

// function subLabel(sub: Subscription | undefined | null) {
//     const status = sub && sub.status;
//     switch (status) {
//         case SubscriptionStatus.Trialing:
//             const daysLeft = Math.ceil((sub!.trialEnd!.getTime() - Date.now()) / 1000 / 60 / 60 / 24);
//             return `${status} (${daysLeft}d)`;
//         case SubscriptionStatus.Active:
//         case SubscriptionStatus.Inactive:
//         case SubscriptionStatus.Canceled:
//             return status;
//         default:
//             return "N/A";
//     }
// }

// function col(val: any, width = 30) {
//     let str = val.toString();
//     if (str.length > width) {
//         str = str.slice(0, width - 3) + "...";
//     }
//     return str.padEnd(width, " ");
// }

// function displayAccountItem(account: Account) {
//     // const sub = account.billing && account.billing.subscription;
//     // const planName = sub ? sub.plan.name : "none";
//     // const planType = sub ? sub.plan.type : undefined;
//     const planName = "none";
//     const planType = undefined;
//     // const lastActive = account.sessions.length
//     //     ? formatDistanceToNow(new Date(Math.max(...account.sessions.map(s => s.lastUsed.getTime()))))
//     //     : "N/A";
//     return [
//         colors.bold(col(account.email, 30)),
//         col(account.name, 20),
//         col(format(account.created, "yyyy-MM-dd"), 12),
//         col("N/A", 15),
//         col("N/A", 5),
//         colors.bold[planColor(planType)](col(planName, 15)),
//         colors.bold[subColor(sub)](col(subLabel(sub), 20)),
//         col(account.orgs.length.toString(), 5),
//         colors.dim(account.id),
//     ].join(" ");
// }

// function displayOrgItem(org: Org) {
//     const owner = org.members.find((m) => m.role === OrgRole.Owner);
//     const sub = org.billing && org.billing.subscription;
//     const planName = sub ? sub.plan.name : "none";
//     const planType = sub ? sub.plan.type : undefined;
//     return [
//         colors.bold(col(org.name, 20)),
//         col((owner && owner.email) || "N/A", 25),
//         col(format(org.created, "yyyy-MM-dd"), 12),
//         col(org.members.length, 7),
//         col(org.groups.length, 7),
//         col(org.vaults.length, 7),
//         colors.bold[planColor(planType)](col(planName, 15)),
//         colors.bold[subColor(sub)](col(subLabel(sub), 20)),
//         colors.dim(org.id),
//         org.frozen ? colors.red.bold("frozen") : "",
//     ].join(" ");
// }

// type ListAccountsOptions = StorageListOptions<Account> & { name?: string; email?: string };
// type ListOrgsOptions = StorageListOptions<Org> & {
//     name?: string;
//     member?: { id?: string; name?: string; email?: string };
// };

// export class ReplSession {
//     constructor(public server: Server, public socket: net.Socket) {}

//     wrap(fn: (...args: any[]) => Promise<any>) {
//         return (...args: any[]) => {
//             fn.apply(this, args).catch((e: Error) => {
//                 this.error(e);
//             });
//         };
//     }

//     start() {
//         const r = repl.start({
//             prompt: "> ",
//             input: this.socket,
//             output: this.socket,
//             terminal: true,
//             useGlobal: false,
//         });
//         if (process.env.PL_REPL_HISTORY) {
//             r.setupHistory(process.env.PL_REPL_HISTORY, () => {});
//         }
//         Object.assign(r.context, {
//             server: this.server,
//             storage: this.server.storage,
//             accounts: {
//                 list: this.wrap(this.listAccounts),
//                 get: this.wrap(this.showAccount),
//                 delete: this.wrap(this.deleteAccount),
//                 update: this.wrap(this.updateAccount),
//                 syncBilling: this.wrap(this.syncAccountBilling),
//             },
//             orgs: {
//                 list: this.wrap(this.listOrgs),
//                 get: this.wrap(this.showOrg),
//                 delete: this.wrap(this.deleteOrg),
//                 update: this.wrap(this.updateOrg),
//                 syncBilling: this.wrap(this.syncOrgBilling),
//             },
//             logs: {
//                 list: this.wrap(this.listEvents),
//                 get: this.wrap(this.getEvent),
//             },
//             socket: this.socket,
//             Account,
//             Org,
//             Session,
//             Vault,
//         });
//         r.on("exit", () => this.socket.end());
//     }

//     get storage() {
//         return this.server.storage;
//     }

//     print(obj: any) {
//         if (obj instanceof Serializable) {
//             obj = obj.toRaw();
//         }
//         const str = typeof obj === "object" ? JSON.stringify(obj, null, 4) : obj.toString();
//         this.socket.write(str + "\n");
//     }

//     error(err: Error) {
//         this.print(colors.red(err.toString()));
//     }

//     async listAccounts({ name = "", email = "", ...listOpts }: ListAccountsOptions = {}) {
//         const nameRgx = new RegExp(name, "i");
//         const emailRgx = new RegExp(email, "i");
//         const filter = listOpts.filter || (() => true);

//         listOpts.filter = (acc: Account) => nameRgx.test(acc.name) && emailRgx.test(acc.email) && filter(acc);

//         const accounts = await this.storage.list(Account, listOpts);

//         const header =
//             `${colors.bold(accounts.length.toString())} accounts found\n\n` +
//             [
//                 col("Email", 30),
//                 col("Name", 20),
//                 col("Created", 12),
//                 col("Last Active", 15),
//                 col("Sessions", 5),
//                 col("Plan", 15),
//                 col("Status", 20),
//                 col("Orgs", 5),
//                 "ID",
//             ]
//                 .map((c) => colors.bold.underline(c))
//                 .join(" ");

//         const items = accounts.map((acc) => displayAccountItem(acc));

//         this.print([header, ...items].join("\n"));

//         return accounts;
//     }

//     async showAccount(id: string) {
//         const { name, email, quota, billing, usedStorage, created, updated, sessions, orgs } = (
//             await this.storage.get(Account, id)
//         ).toRaw();
//         this.print({ id, name, email, quota, usedStorage, created, updated, sessions, orgs, billing });
//     }

//     async deleteAccount(id: string) {
//         const account = await this.storage.get(Account, id);
//         const ctlr = this.server.makeController({ session: new Session(), account });
//         await ctlr.deleteAccount();
//         this.print(colors.bold(`${colors.green("✓")} account deleted successfully`));
//     }

//     async syncAccountBilling(id: string) {
//         const acc = await this.storage.get(Account, id);
//         await this.syncBilling(acc);
//         this.print(displayAccountItem(await this.storage.get(Account, id)));
//     }

//     async updateAccount(id: string, transform: (acc: Account) => Promise<Account | unknown>) {
//         const acc = await this.storage.get(Account, id);
//         const res = await transform(acc);
//         await this.storage.save(res instanceof Account ? res : acc);
//     }

//     async showOrg(id: string) {
//         const { name, owner, quota, billing, usedStorage, created, updated, members, groups, vaults } = (
//             await this.storage.get(Org, id)
//         ).toRaw();

//         const { email: ownerEmail, name: ownerName } = await this.storage.get(Account, owner);
//         this.print({
//             id,
//             name,
//             owner: {
//                 id: owner,
//                 email: ownerEmail,
//                 name: ownerName,
//             },
//             quota,
//             usedStorage,
//             created,
//             updated,
//             members: members.length,
//             groups: groups.length,
//             vaults: vaults.length,
//             billing,
//         });
//     }

//     async deleteOrg(id: string) {
//         const org = await this.storage.get(Org, id);
//         const account = await this.storage.get(Account, org.owner);
//         const ctlr = this.server.makeController({ session: new Session(), account });
//         await ctlr.deleteOrg(id);
//         this.print(colors.bold(`${colors.green("✓")} org deleted successfully`));
//     }

//     async syncOrgBilling(id: string) {
//         const acc = await this.storage.get(Org, id);
//         await this.syncBilling(acc);
//         this.print(displayOrgItem(await this.storage.get(Org, id)));
//     }

//     async syncBilling(acc: Account | Org) {
//         await this.server.billingProvider!.update(
//             new UpdateBillingParams(acc instanceof Account ? { account: acc.id } : { org: acc.id })
//         );
//         this.print(colors.bold(`${colors.green("✓")} billing synced successfully`));
//     }

//     async listOrgs({ name = "", member = {}, ...listOpts }: ListOrgsOptions = {}) {
//         const nameReg = name && new RegExp(name, "i");
//         const mNameReg = member.name && new RegExp(member.name, "i");
//         const emailReg = member.email && new RegExp(member.email, "i");

//         const filter = listOpts.filter || (() => true);

//         listOpts.filter = (org: Org) =>
//             (!nameReg || nameReg.test(org.name)) &&
//             org.members.some(
//                 (m) =>
//                     (!member.id || m.id === member.id) &&
//                     (!mNameReg || mNameReg.test(m.name)) &&
//                     (!emailReg || (emailReg as RegExp).test(m.email))
//             ) &&
//             filter(org);

//         const orgs = await this.storage.list(Org, listOpts);

//         const header =
//             `${colors.bold(orgs.length.toString())} orgs found\n\n` +
//             [
//                 col("Name", 20),
//                 col("Owner", 25),
//                 col("Created", 12),
//                 col("Members", 7),
//                 col("Groups", 7),
//                 col("Vaults", 7),
//                 col("Plan", 15),
//                 col("Status", 20),
//                 "ID",
//             ]
//                 .map((c) => colors.bold.underline(c))
//                 .join(" ");

//         const items = orgs.map((org) => displayOrgItem(org));

//         this.print([header, ...items].join("\n"));
//     }

//     async updateOrg(id: string, transform: (org: Org) => Promise<Org | unknown>) {
//         const org = await this.storage.get(Org, id);
//         const res = await transform(org);
//         await this.storage.save(res instanceof Org ? res : org);
//     }

//     async listEvents(opts: ListEventsOptions) {
//         const events = await this.server.logger.listEvents(opts);

//         this.print(
//             [
//                 `${colors.bold(events.length.toString())} Events found:`,
//                 [col("Time", 19), col("Type", 20), col("Account", 20), col("ID", 50)]
//                     .map((c) => colors.bold.underline(c))
//                     .join(" "),
//                 ...events.map((e) =>
//                     [
//                         col(format(e.time, "yyyy-MM-dd hh:mm:ss"), 19),
//                         colors.bold(col(e.type, 20)),
//                         col((e.data.account && e.data.account.email) || e.data.email || "N/A", 20),
//                         colors.dim(e.id),
//                     ].join(" ")
//                 ),
//             ].join("\n")
//         );
//     }

//     async getEvent(id: string) {
//         this.print(await this.server.logger.getEvent(id));
//     }
// }

// export class ReplServer {
//     constructor(public server: Server) {}

//     start(port: number) {
//         net.createServer((socket) => new ReplSession(this.server, socket).start()).listen(port);
//     }
// }

// export class ReplClient {
//     constructor() {}

//     connect(port: number) {
//         const socket = net.connect(port);

//         process.stdin.pipe(socket);
//         socket.pipe(process.stdout);

//         socket.on("connect", () => {
//             console.log("connection successful");
//             process.stdin.resume();
//             if (process.stdin.setRawMode) {
//                 process.stdin.setRawMode(true);
//             }
//         });

//         socket.on("close", function done() {
//             console.log("connection closed");
//             if (process.stdin.setRawMode) {
//                 process.stdin.setRawMode(false);
//             }
//             process.stdin.pause();
//             socket.removeListener("close", done);
//         });

//         process.stdin.on("end", () => {
//             console.log("exiting, closing connection...");
//             socket.destroy();
//         });

//         process.stdin.on("data", (b) => {
//             if (b.length === 1 && b[0] === 4) {
//                 console.log("end?");
//                 process.stdin.emit("end");
//             }
//         });
//     }
// }
