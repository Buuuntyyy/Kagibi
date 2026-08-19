# Changelog

## v2.31.0 — 2026-08-19

### New Features

- **Migration from OneDrive and Dropbox**: in addition to Google Drive, you can now import your files directly from OneDrive and Dropbox into Kagibi (client-side end-to-end encryption before upload), on both the web and the desktop app.
- **Background import transfers**: importing files no longer blocks the migration window — the import continues in the background while you keep using Kagibi, tracked via a floating widget.

### Improvements

- **Batch folder creation during import**: tree reconstruction (Google Drive/OneDrive/Dropbox) now sends folders in grouped batches instead of one request per folder, on both web and desktop — a significant time saving on large imports.
- **Folder conflicts resolved automatically**: a folder that already exists during an import no longer triggers a merge/skip prompt; it's simply reused silently.
- **Encryption key for imported folders (web)**: folders created by a cloud import from the web now receive a dedicated encryption key at creation time, as the desktop app already did.

### Fixes

- **502 error on large imports**: bulk folder creation was processed sequentially on the server, which could exceed the proxy timeout on large imports; processing is now parallelized server-side, and the desktop client timeout has been increased accordingly.
- **Disconnection during long migrations**: the client-side session timer was never reset during a long-running import or transfer, causing an unexpected logout despite continuous activity.

---

## v2.30.0 — 2026-08-13

### New Features

- **File requests by email**: the creator of a drop link can now send it directly by email to the recipient, with a choice of language (French/English), following the same model as the P2P transfer invitation.

### Fixes

- **Decryption of dropped files**: a folder owner could not preview or download files received via a file request link ("This file cannot be decrypted: missing key"). Server-side key retrieval now also covers the case of an anonymous drop (key wrapped with a key derived from the link token), in addition to the already-handled case of a friend dropping files into a shared folder.
- **Key persistence after retrieval**: as soon as a file dropped by a third party is first opened, its key is re-wrapped with the owner's master key and stored permanently, so it remains decryptable even after the drop link or the original share is revoked.
- **Revocation warning**: the confirmation dialog for revoking a file request link now warns that dropped files that were never opened or previewed will become permanently unreadable.

---

## v2.29.0 — 2026-08-10

### New Features

- **Personal trash**: files and folders deleted from the personal drive are now kept in a trash bin before permanent deletion, with restore and empty options, following the same model as the organization trash.
- **File requests**: creation of drop links allowing third parties to send encrypted files into a folder without an account, write-only.
- **Recovery kit**: verification of the recovery code backup and regeneration of a new code.
- **Securing recovery code regeneration**: rotation now systematically requires MFA validation (if enabled) and a confirmation code sent by email.
- **P2P transfer**: comparison table and improved consent flow.

### Fixes

- **Share links**: fixed the URL prefix returned by the server for drop-only links (`/r/` instead of `/s/`).

---

## v2.28.0 — 2026-08-06

### Improvements

- **Server-enforced MFA**: two-factor verification is now enforced by the backend at login and for every sensitive action.
- **TOTP secrets encrypted at rest** and aal2 authentication level required for organization MFA.
- **Freshness window and anti-replay**: per-action MFA validations are time-limited and protected against TOTP code replay.
- **MFA challenge bound to the verification step**.

---

## v2.27.0 — 2026-08-04

### New Features

- **Subscription page**: new upgrade page with dynamic plan and pricing retrieval.
- **What's new modal**: version changes can now be viewed directly from within the app.

### Improvements

- **Kagibi branding**: completed the rename from SaferCloud to Kagibi.
- **File table**: fixed column widths, truncation, and harmonized date/unit formatting.
- **Shares**: collapsible expiration and password options.
- **Event notifications** enriched for files and folders.
- **Security**: input validation, storage quotas on upload, hardened docker-compose/nginx, timing-attack prevention on recovery code verification.
- **Miscellaneous**: image/logo optimization; removed P2P exchange limits.

---

## v2.26.0 — 2026-07-02

### New Features

- **Access requests**: members can now request access to an organization folder. Admins receive these requests and can accept or deny them from the admin panel.
- **Effective access view**: new summary view showing a user's actual permissions on a folder, taking into account both direct rights and the groups they belong to.
- **Group encryption management**: organization admins can manage the per-group encrypted key — granting encrypted access to group members and securely revoking it.
- **Folder permission inheritance**: permissions defined for a group on a folder now correctly propagate to all subfolders.

---

## v2.25.0 — 2026-07-01

### New Features

- **Weak password modal**: during sign-up, password security criteria are now displayed in real time to guide the user.
- **Improved trash management**: per-user restore permissions, optimized listing and restore queries.
- **Context menu in organizations**: actions on files and folders (rename, move, delete, etc.) accessible via right-click in an organization's detail view.
- **Favorites system**: files and folders can be favorited from the personal drive, synced across sessions.

---

## v2.24.0 — 2026-06-29

### Improvements

- **Enhanced logging**: account and sharing operations now log the user agent and IP address for more complete auditing.
- **Multipart upload documentation**: detailed documentation of the upload process and log compliance.
- **Privacy policy and terms of service**: updated revision dates and clarified log handling.
- **P2P transfer**: updated descriptions to clarify there's no size limit and that a TURN relay is used.

---

## v2.23.0 — 2026-06-28

### New Features

- **FAQ page**: new frequently asked questions page accessible from the main navigation and the HelpDialog.
- **File compression**: compression support during upload and download to reduce storage usage.
- **Structured logging**: migrated to `slog` with a Loki handler integration for centralized, structured log management.

---

## v2.22.0 — 2026-06-23

### New Features

- **File versioning panel**: file version history accessible from the admin interface.
- **LDAP configuration**: new admin panel for LDAP/Active Directory integration.
- **Desktop sync**: new `synced` column on folders to track synchronization with the desktop app.

### Improvements

- **Authentication security**: strengthened password derivation with migration for existing accounts.
- **MIME type handling**: MIME type detection and storage on upload, with a fallback file-manager display on error.

---

## v2.21.0 — 2026-06-02

### New Features

- **"Values" page**: new page presenting Kagibi's commitments and values, accessible from the public navigation.
- **Improved Google Drive import**: revamped user experience with better handling of folders, conflicts, and import errors.
- **Comments and notifications**: comment system on files with real-time notifications for replies.
- **Welcome tour**: interactive onboarding guide for new users, fully localized in FR/EN.

---

## v2.20.0 — 2026-05-26

### New Features

- **Google Drive import (zero-knowledge)**: import your files from Google Drive directly into Kagibi. Files are downloaded and encrypted in your browser before any upload — Kagibi never has access to their content.
- **Organization ownership transfer**: owners can transfer ownership of their organization to another member from the settings.
- **Mandatory MFA for organizations**: admins can now require MFA on all organization file and folder operations.
- **Monitoring metrics**: new statistics on active organizations, storage used, and members.
- **Subscriptions banner**: added a "Coming soon" banner for upcoming subscription features.

---

## v2.19.0 — 2026-06-30

### New Features

- **Personal favorites**: any file or folder in the personal drive can be favorited. A star button (★) is available directly in the file table and via the context menu (right-click). Favorites are persisted in the database (`user_favorites`) and synced across sessions.
- **"Favorite files" section on the home page**: a new accordion section appears above the "Recent" section, displaying favorited items in a 5-column grid. It only appears if at least one favorite exists. Clicking an item opens it directly; the golden star visible on hover lets you remove it instantly.
- **Sort by favorites**: the file browser now offers a new sort criterion (★ column) that brings favorited items to the top of the list.

### API

- `GET /api/v1/users/favorites` — lists the current user's favorites (files and folders).
- `POST /api/v1/users/favorites` — adds an item to favorites (`{ id, type: "file"|"folder" }`).
- `DELETE /api/v1/users/favorites/:type/:id` — removes an item from favorites.

---

## v2.18.0 — 2026-05-21

### New Features

- **Per-folder access control**: organization admins (and group admins) can define per-folder permissions for individual users or entire groups. Available levels: `manage`, `write`, `read`, `none`. Permissions are cumulative: the effective level is the highest one granted either directly or via a group.
- **Overview and Administration tabs**: an organization's detail view now has dedicated tabs offering a quick summary and centralized access to admin actions.
- **Total folder size**: each folder's recursive size is now calculated and displayed in the file browser.

### Improvements

- **Audit log**: encrypted fields (file names, paths) are now decrypted client-side before display. One-click export of the full log.
- **Per-organization MFA requirement**: owners and admins can require all members to have MFA enabled; non-compliant members see a blocking screen.
- **Improved deletion**: bulk deletion now works correctly with multi-select in the file browser and organization views.
- **Search within organizations**: encrypted names and paths are decrypted before matching, making search functional even with name encryption enabled.
- Refactoring: unified UI notifications via the dedicated store, cleaned up code structure.

---

## v2.17.0 — 2026-05-19

### New Features

- **Organization public sharing**: generate public links for any file or folder in an organization, with client-side decryption for recipients. Options: password protection, one-time use (link revoked after first access). List and revoke active shares.
- **Sorting and filtering**: the organization file browser lets you sort by name, size, or date (ascending/descending) and filter by type category (images, documents, videos, audio, archives) or by tag.
- **Multi-select**: checkboxes and shift-click to select multiple items, with bulk actions: download, move, delete.
- **Drag and drop**: move files or folders to another folder or breadcrumb segment; drag from the OS to upload directly.
- **File preview**: in-browser preview of images, PDFs, audio, and video files, without downloading.
- **Tag management**: organization-wide tags (with color) applicable to any file or folder; filter by tag in the browser.
- **ZIP download**: download an entire folder or a selection as a ZIP archive.
- **Favorites (pins)**: pin frequently accessed files and folders; they appear in a quick-access strip at the top of the browser.
- **Trash**: soft delete with individual restore, permanent deletion, and full trash emptying by admins.

### Improvements

- Optimized upload and decryption pipeline (performance and memory management).

---

## v2.16.0 — 2026-05-15

### New Features

- **Organizations module**: end-to-end encrypted collaborative spaces. Create, list, and manage organizations from a new dedicated dashboard.
- **Organization E2E encryption**: each organization has an OrgKey (AES-256) generated by the owner and individually re-encrypted for each member via RSA-OAEP 4096. The server never holds the key in plaintext.
- **Key provisioning**: admins can provision the organization key for members who joined via an invitation link, individually or in bulk (provision-all).
- **Groups**: create and manage subgroups within an organization, with roles (group admin / group member) and per-group permission assignment.
- **Setup wizard**: step-by-step wizard guiding the owner through creating their first organization (naming, encryption model, first invitation link).
- **Dashboard**: KPIs (members, files, folders, 7-day activity, active links), alert for members without a provisioned key.
- **Audit log**: records all organization actions with a summary and cleanup of old entries.
- **Premium features**: built-in upgrade prompts for quota limits and advanced features.
- **Admin CLI**: command-line tool to create, list, adjust the quota of, and delete organizations server-side.

---

## v2.15.0 — 2026-05-10

### New Features

- **@unhead/vue integration**: HTML `<head>` management via `@unhead/vue` for better control over page metadata.
- **P2P transfer — manual disconnect**: the sender or recipient can now close the connection at any time via a dedicated button, without waiting for the transfer to finish or fail.
- **Improved P2P buffer management**: optimized throughput and stability during high-speed WebRTC transfers.

### Fixes

- Full reset of `P2PInviteDialog` state on close (avoids residual state between sessions).
- Reset of the transfer wizard at the end of a transfer.
- Mobile UI/UX fixes: bottom navigation, layout, and touch interactions.

---

## v2.14.0 — 2026-05-04

### New Features

- **P2P transfer — status indicators**: real-time display of the WebRTC connection state (connecting, transferring, error, reconnecting) with clear messages for each phase.
- **P2P error handling and reconnection**: detection and reporting of network errors, with automatic reconnection attempts and the ability to manually retry the transfer on failure.

---

## v2.13.0 — 2026-05-04

### New Features

- **Password protection for public links**: share links can now be protected with a password (bcrypt-hashed server-side). Visitors must enter the password before accessing the content.
- **One-time links**: option for a public share link to be automatically revoked after its first successful access.
- **Per-item restrictions in folder shares**: a side panel in the management dialog lets you configure rights per sub-item (folder: full access / read-only / hidden; file: download and delete individually configurable). Navigate the tree via a clickable breadcrumb, with bulk controls per level.
- **ZIP download of shared folders**: recursive retrieval of a shared folder's files with ZIP archive generation.
- **Improved shares overview**: deduplicated shares, one-click link copy, direct navigation to the shared folder in the tree, better public browsing experience.

### Improvements

- Added avatar URL to friend responses (avatar display in the friends list).
- Updated storage limit retrieval in the usage dashboard.

---

## v2.12.0 — 2026-04-27

### Fixes

- **HTTP 500 on public upload**: replaced the `ON CONFLICT` statement (which requires a UNIQUE constraint) with a SELECT → INSERT/UPDATE pattern in the `CompletePublicShareUploadHandler` and `CompleteSharedMultipartHandler` handlers. Files dropped by visitors or friends into a shared folder are now correctly recorded in the database.
- **Downloading files dropped by a friend**: the owner can now download files uploaded by their friends into a shared folder. A new `GET /files/:id/folder-key` endpoint returns the required key chain (`folder_encrypted_key` + `file_encrypted_key`). The client reconstructs the file key in two steps: `MasterKey → FolderKey → FileKey`.
- **Storage quota**: when replacing a file, the delta (new size − old size) is now used instead of the total size, avoiding artificial quota inflation.

### New Features

- **Permission guards**: any action forbidden by share rights (create a folder, delete, rename, drop a file) now displays a clear error message via `useUIStore().showError()` instead of failing silently.
- **ManageShareDialog interface**: permission chips are now color-coded — **green** if the right is granted, **red** if denied — both in the creation dialog and in the management view of an existing share. Neutral hover effect to indicate interactivity.
- **Default permissions**: when creating a new folder share, the rights granted by default are now **Download + Create** (previously Download only).

---

## v2.11.0 — 2026-04-27

### New Features

- **Folder upload**: upload an entire tree in a single operation, with a global progress bar and conflict handling (rename, skip, replace).
- **Improved search bar**: clicking a result navigates directly to the file's location in the tree with visual highlighting.
- **Folder sharing between friends**: granular folder sharing with a registered friend, permission management (download, create, delete, move), browsing shared content, uploading and deleting files according to rights.
- **Public share page**: the link access page lets you browse the tree of a shared folder and drop files into it (encrypted client-side).
- **P2P invitation email language**: the sender can choose to send the invitation in French or English, with enriched content in both languages.

### Fixes

- Progress bar performance: limited to 4 updates per second to free up resources during upload.
- Empty files: uploaded with a minimum size to work around an encryption constraint.
- P2P invitation email link: fixed the redirect URL.
- Analytics script (`umami`): added the missing CSP nonce.
- Dynamic statistics on the `send` subdomain.
