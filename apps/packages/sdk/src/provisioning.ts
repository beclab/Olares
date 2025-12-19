// import {
//     //OrgProvisioning,
//     ProvisioningStatus,
//     AccountProvisioning,
//     Feature,
//     OrgFeature,
//     RichContent,
// } from "@didvault/core/src/provisioning";
//import { html } from "lit";
//import { alert } from "./dialog";
//import "../elements/rich-content";
//import { openExternalUrl } from "@didvault/core/src/platform";

// export function checkFeatureDisabled(feature: Feature): boolean;
// export function checkFeatureDisabled(feature: OrgFeature, isOwner: boolean): boolean;
// export function checkFeatureDisabled(feature: Feature | OrgFeature, isOwner?: boolean) {
//     if (feature.disabled) {
//         alertDisabledFeature(feature, isOwner as boolean);
//     }
//     return feature.disabled;
// }

// async function alertMessage(message: string | RichContent, action?: { label: string; url: string }, title?: string) {
//     if (typeof message === "string") {
//         message = {
//             type: "plain",
//             content: message,
//         } as RichContent;
//     }

//     const choice = await alert(
//         html`<pl-rich-content .content=${message.content} .type=${message.type}></pl-rich-content>`,
//         {
//             icon: null,
//             width: "auto",
//             maxWidth: message?.type === "html" ? "100%" : "",
//             options: action ? [action.label!, $l("Dismiss")] : [$l("Dismiss")],
//             type: !action ? "choice" : "info",
//             title,
//             hideOnDocumentVisibilityChange: true,
//             zIndex: 20,
//         }
//     );
//     if (action && choice === 0) {
//         openExternalUrl(action.url);
//     }
// }

//export function alertDisabledFeature(feature: Feature): Promise<void>;
//export function alertDisabledFeature(feature: OrgFeature, isOwner: boolean): Promise<void>;
// export async function alertDisabledFeature(feature: Feature | OrgFeature, isOwner?: boolean) {
//     const message =
//         feature instanceof OrgFeature && isOwner ? feature.messageOwner || feature.message : feature.message;
//     alertMessage(
//         message || "",
//         feature.actionLabel && feature.actionUrl ? { label: feature.actionLabel, url: feature.actionUrl } : undefined
//     );
// }

// export function getDefaultStatusMessage(status: ProvisioningStatus) {
//     switch (status) {
//         case ProvisioningStatus.Frozen:
//             return $l(
//                 "Your account has been frozen, meaning you can still access your existing data, but you won't be able to create new vault items or edit existing ones."
//             );
//         case ProvisioningStatus.Suspended:
//             return $l(
//                 "Your account has been suspended, meaning you can no longer use this service. If you believe your account has been suspended in error, please contact your service administrator or customer support."
//             );
//         case ProvisioningStatus.Unprovisioned:
//             return $l(
//                 "You don't currently have permission to use this service. Please contact the service adminstrator to request access."
//             );
//         default:
//             return "";
//     }
// }

// export function getDefaultStatusLabel(status: ProvisioningStatus) {
//     switch (status) {
//         case ProvisioningStatus.Active:
//             return $l("Active");
//         case ProvisioningStatus.Deleted:
//             return $l("Account Deleted");
//         case ProvisioningStatus.Frozen:
//             return $l("Account Frozen");
//         case ProvisioningStatus.Suspended:
//             return $l("Account Suspended");
//         case ProvisioningStatus.Unprovisioned:
//             return $l("Access Denied");
//     }
// }

// export async function displayProvisioning({
//     status,
//     statusMessage,
//     statusLabel,
//     actionUrl,
//     actionLabel,
// }: AccountProvisioning | OrgProvisioning) {
//     // alertMessage(
//     //     statusMessage || getDefaultStatusMessage(status),
//     //     actionUrl && actionLabel ? { label: actionLabel, url: actionUrl } : undefined,
//     //     statusLabel || getDefaultStatusLabel(status)
//     // );
// }
