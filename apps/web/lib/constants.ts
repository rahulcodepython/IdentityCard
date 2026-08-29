// API Route Prefixes
export const API_V1_PREFIX = "/api/v1"

// User / Platform Roles
export const USER_ROLE_ADMIN = "admin"
export const USER_ROLE_USER = "user"
export type UserRole = typeof USER_ROLE_ADMIN | typeof USER_ROLE_USER

// Organization Member Roles
export const ROLE_ADMIN = "admin"
export const ROLE_MEMBER = "member"
export type OrgRole = typeof ROLE_ADMIN | typeof ROLE_MEMBER

// Storage Keys
export const STORAGE_KEY_SESSION = "identitycard_session"
export const STORAGE_KEY_DEVICE = "identitycard_device"

// Common Error Messages
export const ERR_MSG_GENERIC = "An unexpected error occurred. Please try again."
export const ERR_MSG_UNAUTHORIZED = "You must be signed in to perform this action."
export const ERR_MSG_SESSION_EXPIRED = "Your session has expired. Please sign in again."
export const ERR_MSG_FAILED_INVITE = "Couldn't send invitation."
export const ERR_MSG_FAILED_ROLE_UPDATE = "Couldn't update role."
export const ERR_MSG_FAILED_REMOVE_MEMBER = "Couldn't remove member."
export const ERR_MSG_FAILED_CANCEL_INVITE = "Couldn't cancel invitation."

// Common Success / Toast Messages
export const MSG_INVITE_SENT = "Invitation sent successfully"
export const MSG_INVITE_CANCELED = "Invitation canceled"
export const MSG_ROLE_UPDATED = "Role updated"
export const MSG_CARD_RESENT = "Card resent"
export const MSG_SETTINGS_SAVED = "Settings saved"

