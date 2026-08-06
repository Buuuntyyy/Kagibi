// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Global MFA step-up coordinator.
 *
 * When the backend answers a request with 403 {"error":"mfa_required"} (because the
 * user's account requires MFA for that action and the session is still aal1), the API
 * interceptor calls `requestStepUp()`. This shows a single global MFA modal
 * (mounted once in App.vue). Once the user verifies, the session token is upgraded to
 * aal2 and every request that was waiting is retried.
 *
 * Concurrent 403s are coalesced into one modal / one pending promise, so a page that
 * fires several gated requests at once only prompts the user a single time.
 */
import { ref } from 'vue'

// Drives the visibility and message of the single global MFA modal in App.vue.
export const stepUpVisible = ref(false)
export const stepUpContext = ref('general')

// Shared pending step-up, so concurrent prompts (login + background 403s) await the
// same modal — the user is never shown two MFA dialogs at once.
let pending = null

/**
 * Request an MFA step-up. Returns a promise that resolves when the user completes
 * MFA (session is now aal2) and rejects if they cancel. Concurrent callers share one
 * modal/promise; the first caller's context sets the displayed message.
 */
export function requestStepUp(context = 'general') {
  if (pending) return pending.promise
  let resolve, reject
  const promise = new Promise((res, rej) => {
    resolve = res
    reject = rej
  })
  pending = { promise, resolve, reject }
  stepUpContext.value = context
  stepUpVisible.value = true
  return promise
}

/** Called by the global modal on successful verification. */
export function completeStepUp() {
  stepUpVisible.value = false
  if (pending) {
    pending.resolve(true)
    pending = null
  }
}

/** Called by the global modal when the user cancels the step-up. */
export function cancelStepUp() {
  stepUpVisible.value = false
  if (pending) {
    pending.reject(new Error('mfa_cancelled'))
    pending = null
  }
}
