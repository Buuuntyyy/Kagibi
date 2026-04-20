// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

const hostname = window.location.hostname

export const isP2PSubdomain =
  hostname === 'send.kagibi.cloud' ||
  hostname === 'send-dev.kagibi.cloud'

// Base URL of the Drive app — used to redirect out of the P2P subdomain
const DRIVE_BASE_URLS = {
  'send.kagibi.cloud':     'https://kagibi.cloud',
  'send-dev.kagibi.cloud': 'https://dev.kagibi.cloud',
}

export const driveBaseUrl = DRIVE_BASE_URLS[hostname] ?? null

// Base URL of the P2P send app — used to redirect to it from the main domain
const SEND_BASE_URLS = {
  'kagibi.cloud':     'https://send.kagibi.cloud',
  'dev.kagibi.cloud': 'https://send-dev.kagibi.cloud',
}

export const sendBaseUrl = SEND_BASE_URLS[hostname] ?? 'https://send.kagibi.cloud'

export function useSubdomain() {
  return { isP2PSubdomain, driveBaseUrl, sendBaseUrl }
}
