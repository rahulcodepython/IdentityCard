// =====================================================================
// API Route Prefixes & Public Unscoped Routes
// =====================================================================
export const API_V1_PREFIX = "/api/v1";

export const DEFAULT_API_BASE_URL = "http://localhost:8000";
export const DEFAULT_APP_BASE_URL = "http://localhost:3000";

// Paths that should not be automatically scoped under /organization/:orgId
export const PUBLIC_OR_UNSCOPED_PREFIXES = [
    "/organization/",
    "/organizations",
    "/plans",
    "/public/",
    "/forms/public/",
    "/devices/pair",
    "/scanner/",
] as const;

// =====================================================================
// Application Routes
// =====================================================================
export const ROUTE_LOGIN = "/auth/login";
export const ROUTE_REGISTER = "/auth/register";
export const ROUTE_DASHBOARD = "/dashboard";
export const ROUTE_ACCEPT_INVITATION = "/accept-invitation";
export const ROUTE_PAIR = "/pair";
export const ROUTE_SCANNER = "/scanner";

// =====================================================================
// User & Platform Roles
// =====================================================================
export const USER_ROLE_ADMIN = "admin";
export const USER_ROLE_USER = "user";
export type UserRole = typeof USER_ROLE_ADMIN | typeof USER_ROLE_USER;

// =====================================================================
// Organization Member Roles (Better-Auth defaults: owner, admin, member)
// =====================================================================
export const ROLE_OWNER = "owner";
export const ROLE_ADMIN = "admin";
export const ROLE_MEMBER = "member";
export type OrgRole = typeof ROLE_OWNER | typeof ROLE_ADMIN | typeof ROLE_MEMBER;

// =====================================================================
// Storage Keys
// =====================================================================
export const STORAGE_KEY_SESSION = "identitycard_session";
export const STORAGE_KEY_DEVICE = "identitycard_device";
export const STORAGE_KEY_DEVICE_KEY = "identitycard_device_key";

// =====================================================================
// File Sizes & Network Timing Limits
// =====================================================================
export const MAX_LOGO_SIZE_BYTES = 2 * 1024 * 1024; // 2 MB
export const MAX_BANNER_SIZE_BYTES = 5 * 1024 * 1024; // 5 MB
export const TOKEN_EXPIRY_BUFFER_MS = 15_000; // 15 seconds

// =====================================================================
// Common Error Messages
// =====================================================================
export const ERR_MSG_GENERIC = "An unexpected error occurred. Please try again.";
export const ERR_MSG_UNAUTHORIZED = "You must be signed in to perform this action.";
export const ERR_MSG_SESSION_EXPIRED = "Your session has expired. Please sign in again.";
export const ERR_MSG_FAILED_INVITE = "Couldn't send invitation.";
export const ERR_MSG_FAILED_ROLE_UPDATE = "Couldn't update role.";
export const ERR_MSG_FAILED_REMOVE_MEMBER = "Couldn't remove member.";
export const ERR_MSG_FAILED_CANCEL_INVITE = "Couldn't cancel invitation.";
export const ERR_MSG_LOGO_TOO_LARGE = "Logo image file size must be less than 2MB.";
export const ERR_MSG_ORG_NAME_REQUIRED = "Organization name cannot be empty.";
export const ERR_MSG_FAILED_DELETE_LOGO = "Failed to delete logo.";
export const ERR_MSG_FAILED_UPDATE_NAME = "Failed to update name.";
export const ERR_MSG_FAILED_UPLOAD_LOGO = "Failed to upload logo.";
export const ERR_MSG_REQUEST_FAILED = "Request failed";
export const ERR_MSG_PAIRING_FAILED = "Invalid code or pairing failed. Please try again.";

// =====================================================================
// Common Success / Toast Messages
// =====================================================================
export const MSG_INVITE_SENT = "Invitation sent successfully";
export const MSG_INVITE_CANCELED = "Invitation canceled";
export const MSG_ROLE_UPDATED = "Role updated";
export const MSG_CARD_RESENT = "Card resent";
export const MSG_SETTINGS_SAVED = "Settings saved";
export const MSG_LOGO_REMOVED = "Logo removed.";
export const MSG_LOGO_UPLOADED = "Logo uploaded";
export const MSG_ORG_METADATA_UPDATED = "Organization metadata updated successfully!";
export const MSG_DEVICE_PAIRED = "Device paired successfully!";
export const MSG_SIGNED_IN = "Signed in successfully!";
export const MSG_ACCOUNT_CREATED = "Account created successfully!";
export const MSG_EVENT_CREATED = "Event created successfully";
export const MSG_EVENT_UPDATED = "Event updated successfully";
export const MSG_EVENT_PUBLISHED = "Event published successfully";
export const MSG_EVENT_DELETED = "Draft event deleted";
export const MSG_SCHEDULE_DAYS_IMPORTED = "Schedule days imported";
export const MSG_BANNER_UPLOADED = "Event banner uploaded";
export const MSG_SIGNATURE_UPLOADED = "Signature uploaded";
export const MSG_CARD_EMAIL_RESENT = "ID card email resent";
export const MSG_DEVICE_CREATED = "Device created — enter OTP on device";
export const MSG_DEVICE_REVOKED = "Device access revoked";
export const MSG_ORG_SETTINGS_UPDATED = "Organization settings updated";
export const MSG_ORG_DELETED = "Organization deleted";
export const MSG_FORM_LINK_CREATED = "Registration link created";
export const MSG_FORM_SETTINGS_UPDATED = "Form settings updated";
export const MSG_FORM_DELETED = "Form deleted";
export const MSG_REGISTRATION_SUBMITTED = "Registration submitted successfully!";
export const MSG_ATTENDEE_ADDED = "Attendee added";
export const MSG_ATTENDEE_UPDATED = "Attendee updated";
export const MSG_ATTENDEE_REMOVED = "Attendee removed";
export const MSG_ATTENDEES_CSV_IMPORTED = "Attendees CSV imported";
export const MSG_CREDITS_PURCHASED = "Credits purchased successfully!";
export const MSG_ANNUAL_RENEWED = "Annual subscription renewed successfully!";
export const MSG_SUBEVENT_CREATED = "Sub-event created";
export const MSG_SUBEVENT_UPDATED = "Sub-event updated";
export const MSG_SUBEVENT_DELETED = "Sub-event deleted";
