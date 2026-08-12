// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

// Kit de récupération : fichier texte imprimable contenant le code de
// récupération et les instructions. Généré entièrement côté client — le code
// ne transite jamais en clair vers le serveur.

import i18n from '../i18n'

function kitLine(char = '=', len = 62) {
  return char.repeat(len)
}

export function buildRecoveryKitText({ code, email }) {
  const t = i18n.global.t
  const date = new Date().toLocaleDateString()
  const origin = typeof window !== 'undefined' ? window.location.origin : 'https://kagibi.cloud'

  return [
    kitLine(),
    `  KAGIBI — ${t('recoveryKit.kitTitle')}`,
    kitLine(),
    '',
    `${t('recoveryKit.kitAccount')} : ${email || '-'}`,
    `${t('recoveryKit.kitDate')} : ${date}`,
    '',
    `${t('recoveryKit.kitCodeLabel')} :`,
    '',
    kitLine('-'),
    `  ${code}`,
    kitLine('-'),
    '',
    t('recoveryKit.kitInstructions1'),
    t('recoveryKit.kitInstructions2'),
    t('recoveryKit.kitInstructions3'),
    '',
    `${t('recoveryKit.kitRecoverAt')} : ${origin}/login`,
    '',
    kitLine(),
    `  ${t('recoveryKit.kitWarning')}`,
    kitLine(),
    '',
  ].join('\n')
}

export function downloadRecoveryKit({ code, email }) {
  const text = buildRecoveryKitText({ code, email })
  const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'kagibi-recovery-kit.txt'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
