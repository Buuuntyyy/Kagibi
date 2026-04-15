// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

const hostname = window.location.hostname

export const isP2PSubdomain =
  hostname === 'send.kagibi.cloud' ||
  hostname === 'send-dev.kagibi.cloud'

export function useSubdomain() {
  return { isP2PSubdomain }
}
