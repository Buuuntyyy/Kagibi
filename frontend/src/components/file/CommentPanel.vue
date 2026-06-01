<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <Teleport to="body">
    <Transition name="panel-slide">
      <div v-if="commentStore.panelOpen" class="comment-panel-overlay" @click.self="commentStore.closePanel()">
        <div class="comment-panel">

          <!-- Header -->
          <div class="panel-header">
            <div class="panel-title">
              <MessageSquare :size="17" />
              <span>Commentaires</span>
              <span v-if="commentStore.panelFile" class="panel-filename">
                — {{ commentStore.panelFile.Name ?? commentStore.panelFile.name }}
              </span>
            </div>
            <button class="panel-close" @click="commentStore.closePanel()">
              <X :size="20" />
            </button>
          </div>

          <!-- Comment list -->
          <div class="panel-body" ref="bodyRef">
            <div v-if="commentStore.loading" class="panel-state">Chargement…</div>

            <div v-else-if="!topLevel.length" class="panel-state empty">
              <MessageSquare :size="38" class="empty-icon" />
              <p>Aucun commentaire.</p>
              <p class="empty-sub">Soyez le premier à commenter.</p>
            </div>

            <div v-else class="comment-list">
              <div v-for="comment in topLevel" :key="comment.id" class="comment-thread">

                <!-- Parent comment -->
                <div class="comment-item" :class="{ unread: !comment.is_read, resolved: comment.is_resolved }">
                  <div class="comment-meta">
                    <div class="comment-avatar">{{ initials(comment.author_name) }}</div>
                    <div class="comment-info">
                      <span class="comment-author">{{ comment.author_name || 'Utilisateur' }}</span>
                      <span class="comment-date">{{ fmtDate(comment.created_at) }}</span>
                    </div>
                    <div class="comment-badges">
                      <span v-if="comment.is_resolved" class="badge resolved-badge">Résolu</span>
                      <span v-if="!comment.is_read" class="badge unread-badge">Non lu</span>
                    </div>
                  </div>

                  <div v-if="editingID !== comment.id" class="comment-content">{{ comment.content }}</div>
                  <div v-else class="comment-edit-form">
                    <textarea v-model="editContent" class="comment-textarea" rows="3" />
                    <div class="edit-actions">
                      <button class="btn-save" @click="saveEdit(comment.id)">Enregistrer</button>
                      <button class="btn-cancel" @click="cancelEdit">Annuler</button>
                    </div>
                  </div>

                  <div class="comment-actions">
                    <button v-if="!comment.is_read" class="action-btn" @click="markRead(comment)">
                      <CheckCheck :size="13" /> Lu
                    </button>
                    <button class="action-btn" @click="toggleResolve(comment)">
                      <component :is="comment.is_resolved ? CircleDot : CircleCheck" :size="13" />
                      {{ comment.is_resolved ? 'Rouvrir' : 'Résoudre' }}
                    </button>
                    <button class="action-btn" @click="startReply(comment.id)">
                      <Reply :size="13" /> Répondre
                    </button>
                    <button v-if="comment.author_id === currentUserID" class="action-btn" @click="startEdit(comment)">
                      <Pencil :size="13" /> Modifier
                    </button>
                    <button v-if="comment.author_id === currentUserID" class="action-btn danger" @click="deleteComment(comment.id)">
                      <Trash2 :size="13" />
                    </button>
                  </div>
                </div>

                <!-- Inline reply form for this comment -->
                <div v-if="replyingTo === comment.id" class="reply-form">
                  <div class="reply-avatar">{{ currentInitials }}</div>
                  <div class="reply-input-wrap">
                    <textarea
                      v-model="replyContent"
                      class="comment-textarea reply-textarea"
                      :placeholder="`Répondre à ${comment.author_name}…`"
                      rows="2"
                      @keydown.ctrl.enter.prevent="submitReply(comment.id)"
                      ref="replyInputRef"
                    />
                    <div class="reply-actions">
                      <button class="btn-cancel" @click="cancelReply">Annuler</button>
                      <button class="btn-save" :disabled="!replyContent.trim() || submitting" @click="submitReply(comment.id)">
                        <Send :size="13" /> Envoyer
                      </button>
                    </div>
                  </div>
                </div>

                <!-- Replies -->
                <div v-if="replies[comment.id]?.length" class="reply-list">
                  <div
                    v-for="reply in replies[comment.id]"
                    :key="reply.id"
                    class="comment-item reply-item"
                    :class="{ unread: !reply.is_read }"
                  >
                    <div class="comment-meta">
                      <div class="comment-avatar reply-avatar-sm">{{ initials(reply.author_name) }}</div>
                      <div class="comment-info">
                        <span class="comment-author">{{ reply.author_name || 'Utilisateur' }}</span>
                        <span class="comment-date">{{ fmtDate(reply.created_at) }}</span>
                      </div>
                      <span v-if="!reply.is_read" class="badge unread-badge">Non lu</span>
                    </div>

                    <div v-if="editingID !== reply.id" class="comment-content">{{ reply.content }}</div>
                    <div v-else class="comment-edit-form">
                      <textarea v-model="editContent" class="comment-textarea" rows="2" />
                      <div class="edit-actions">
                        <button class="btn-save" @click="saveEdit(reply.id)">Enregistrer</button>
                        <button class="btn-cancel" @click="cancelEdit">Annuler</button>
                      </div>
                    </div>

                    <div class="comment-actions">
                      <button v-if="!reply.is_read" class="action-btn" @click="markRead(reply)">
                        <CheckCheck :size="13" /> Lu
                      </button>
                      <button v-if="reply.author_id === currentUserID" class="action-btn" @click="startEdit(reply)">
                        <Pencil :size="13" /> Modifier
                      </button>
                      <button v-if="reply.author_id === currentUserID" class="action-btn danger" @click="deleteComment(reply.id)">
                        <Trash2 :size="13" />
                      </button>
                    </div>
                  </div>
                </div>

              </div>
            </div>
          </div>

          <!-- New top-level comment form -->
          <div class="panel-footer">
            <textarea
              v-model="newContent"
              class="comment-textarea"
              placeholder="Ajouter un commentaire…"
              rows="3"
              @keydown.ctrl.enter.prevent="submitComment"
            />
            <div class="footer-actions">
              <span class="hint">Ctrl+Entrée pour envoyer</span>
              <button class="btn-submit" :disabled="!newContent.trim() || submitting" @click="submitComment">
                <Send :size="15" /> Envoyer
              </button>
            </div>
          </div>

        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { useCommentStore } from '../../stores/comments'
import { useAuthStore } from '../../stores/auth'
import {
  MessageSquare, X, CheckCheck, CircleCheck, CircleDot,
  Pencil, Trash2, Send, Reply
} from 'lucide-vue-next'

const commentStore = useCommentStore()
const authStore = useAuthStore()

const newContent = ref('')
const submitting = ref(false)
const editingID = ref(null)
const editContent = ref('')
const replyingTo = ref(null) // comment ID being replied to
const replyContent = ref('')
const replyInputRef = ref(null)

const currentUserID = computed(() => authStore.user?.id)
const currentInitials = computed(() => {
  const name = authStore.user?.name
  if (!name) return '?'
  return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()
})

// Separate top-level comments from replies
const topLevel = computed(() =>
  commentStore.comments.filter(c => !c.parent_id)
)

const replies = computed(() => {
  const map = {}
  commentStore.comments
    .filter(c => c.parent_id)
    .forEach(c => {
      if (!map[c.parent_id]) map[c.parent_id] = []
      map[c.parent_id].push(c)
    })
  return map
})

// ── Helpers ──────────────────────────────────────────────────────────────────

function initials(name) {
  if (!name) return '?'
  return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()
}

function fmtDate(iso) {
  if (!iso) return ''
  return new Date(iso).toLocaleString('fr-FR', {
    day: '2-digit', month: '2-digit', year: 'numeric',
    hour: '2-digit', minute: '2-digit'
  })
}

function fileID() { return commentStore.panelFile?.ID ?? commentStore.panelFile?.id }

// ── Actions ───────────────────────────────────────────────────────────────────

async function submitComment() {
  const content = newContent.value.trim()
  if (!content || submitting.value) return
  submitting.value = true
  try {
    await commentStore.addComment(fileID(), content, commentStore.panelType, commentStore.panelOrgID)
    newContent.value = ''
  } finally {
    submitting.value = false
  }
}

function startReply(commentID) {
  replyingTo.value = commentID
  replyContent.value = ''
  cancelEdit()
  nextTick(() => replyInputRef.value?.focus())
}

function cancelReply() {
  replyingTo.value = null
  replyContent.value = ''
}

async function submitReply(parentID) {
  const content = replyContent.value.trim()
  if (!content || submitting.value) return
  submitting.value = true
  try {
    await commentStore.addComment(fileID(), content, commentStore.panelType, commentStore.panelOrgID, parentID)
    cancelReply()
  } finally {
    submitting.value = false
  }
}

function startEdit(comment) {
  editingID.value = comment.id
  editContent.value = comment.content
  cancelReply()
}

async function saveEdit(commentID) {
  if (!editContent.value.trim()) return
  await commentStore.editComment(commentID, editContent.value.trim())
  cancelEdit()
}

function cancelEdit() {
  editingID.value = null
  editContent.value = ''
}

async function markRead(comment) {
  await commentStore.markRead(comment.id, fileID(), commentStore.panelType, commentStore.panelOrgID)
}

async function toggleResolve(comment) {
  await commentStore.resolveComment(comment.id, !comment.is_resolved, fileID(), commentStore.panelType, commentStore.panelOrgID)
}

async function deleteComment(commentID) {
  await commentStore.deleteComment(commentID)
}
</script>

<style scoped>
.comment-panel-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(0, 0, 0, 0.32);
  display: flex;
  justify-content: flex-end;
}

.comment-panel {
  width: 420px;
  max-width: 100vw;
  height: 100%;
  background: var(--card-color);
  display: flex;
  flex-direction: column;
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.18);
}

/* Slide animation */
.panel-slide-enter-active,
.panel-slide-leave-active { transition: opacity 0.2s ease; }
.panel-slide-enter-active .comment-panel,
.panel-slide-leave-active .comment-panel { transition: transform 0.25s cubic-bezier(0.4,0,0.2,1); }
.panel-slide-enter-from .comment-panel,
.panel-slide-leave-to .comment-panel { transform: translateX(100%); }
.panel-slide-enter-from,
.panel-slide-leave-to { opacity: 0; }

/* Header */
.panel-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}
.panel-title {
  display: flex; align-items: center; gap: 0.5rem;
  font-weight: 600; font-size: 0.95rem; color: var(--main-text-color); min-width: 0;
}
.panel-filename {
  color: var(--secondary-text-color); font-weight: 400; font-size: 0.85rem;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.panel-close {
  background: none; border: none; cursor: pointer;
  color: var(--secondary-text-color); padding: 0.25rem; border-radius: 4px;
  display: flex; align-items: center; flex-shrink: 0;
}
.panel-close:hover { background: var(--hover-background-color); }

/* Body */
.panel-body {
  flex: 1; overflow-y: auto; padding: 0.75rem 1rem;
  display: flex; flex-direction: column; gap: 0; min-height: 0;
}

.panel-state {
  flex: 1; display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  color: var(--secondary-text-color); text-align: center; gap: 0.5rem;
}
.empty-icon { opacity: 0.3; margin-bottom: 0.25rem; }
.empty-sub { font-size: 0.82rem; opacity: 0.7; }

/* Thread */
.comment-thread { margin-bottom: 1rem; }

/* Comment item */
.comment-item {
  background: var(--background-color);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 0.75rem;
  display: flex; flex-direction: column; gap: 0.45rem;
}
.comment-item.unread { border-left: 3px solid var(--primary-color); }
.comment-item.resolved { opacity: 0.6; }

.comment-meta { display: flex; align-items: center; gap: 0.45rem; }
.comment-avatar {
  width: 28px; height: 28px; border-radius: 50%;
  background: var(--primary-color); color: #fff;
  font-size: 0.65rem; font-weight: 700;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.reply-avatar-sm { width: 24px; height: 24px; font-size: 0.6rem; }
.comment-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.comment-author { font-weight: 600; font-size: 0.83rem; color: var(--main-text-color); }
.comment-date { font-size: 0.72rem; color: var(--secondary-text-color); }
.comment-badges { display: flex; gap: 0.25rem; flex-shrink: 0; }

.badge { font-size: 0.62rem; padding: 2px 5px; border-radius: 99px; font-weight: 600; }
.resolved-badge { background: #d1fae5; color: #065f46; }
.unread-badge { background: #dbeafe; color: #1e40af; }

.comment-content {
  font-size: 0.85rem; color: var(--main-text-color);
  line-height: 1.5; white-space: pre-wrap; word-break: break-word;
}

.comment-actions { display: flex; gap: 0.35rem; flex-wrap: wrap; }
.action-btn {
  display: flex; align-items: center; gap: 0.2rem;
  background: none; border: none; cursor: pointer;
  font-size: 0.72rem; color: var(--secondary-text-color);
  padding: 0.15rem 0.35rem; border-radius: 4px;
}
.action-btn:hover { background: var(--hover-background-color); color: var(--main-text-color); }
.action-btn.danger:hover { background: #fee2e2; color: #dc2626; }

/* Edit form */
.comment-edit-form { display: flex; flex-direction: column; gap: 0.4rem; }
.edit-actions { display: flex; gap: 0.4rem; }

/* Inline reply form */
.reply-form {
  display: flex; gap: 0.5rem;
  margin-top: 0.4rem; padding-left: 1.5rem;
}
.reply-avatar {
  width: 24px; height: 24px; border-radius: 50%;
  background: var(--primary-color); color: #fff;
  font-size: 0.6rem; font-weight: 700;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0; margin-top: 4px;
}
.reply-input-wrap { flex: 1; display: flex; flex-direction: column; gap: 0.35rem; }
.reply-textarea { min-height: 52px; }
.reply-actions { display: flex; gap: 0.4rem; justify-content: flex-end; }

/* Replies */
.reply-list {
  padding-left: 1.75rem; margin-top: 0.35rem;
  display: flex; flex-direction: column; gap: 0.35rem;
  border-left: 2px solid var(--border-color);
}
.reply-item { background: color-mix(in srgb, var(--card-color) 60%, var(--background-color)); }

/* Shared textarea */
.comment-textarea {
  width: 100%; resize: vertical; box-sizing: border-box;
  border: 1px solid var(--border-color); border-radius: 6px;
  padding: 0.5rem 0.6rem; font-size: 0.85rem;
  color: var(--main-text-color); background: var(--background-color);
  font-family: inherit;
}
.comment-textarea:focus { outline: none; border-color: var(--primary-color); }

.btn-save, .btn-cancel {
  padding: 0.28rem 0.65rem; border-radius: 5px;
  font-size: 0.78rem; cursor: pointer; border: 1px solid var(--border-color);
  display: flex; align-items: center; gap: 0.25rem;
}
.btn-save { background: var(--primary-color); color: #fff; border-color: transparent; }
.btn-cancel { background: var(--card-color); color: var(--main-text-color); }

/* Footer */
.panel-footer {
  flex-shrink: 0; padding: 0.7rem 1rem;
  border-top: 1px solid var(--border-color);
  display: flex; flex-direction: column; gap: 0.4rem;
}
.footer-actions { display: flex; align-items: center; justify-content: space-between; }
.hint { font-size: 0.72rem; color: var(--secondary-text-color); }
.btn-submit {
  display: flex; align-items: center; gap: 0.35rem;
  padding: 0.4rem 0.9rem; background: var(--primary-color);
  color: #fff; border: none; border-radius: 6px; cursor: pointer; font-size: 0.83rem; font-weight: 500;
}
.btn-submit:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-submit:not(:disabled):hover { filter: brightness(1.08); }
</style>
