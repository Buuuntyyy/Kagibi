/**
 * Password strength validation — SaferCloud security policy
 *
 * Rules:
 *  - At least 20 characters
 *  - At least 2 digits (0-9)
 *  - At least 2 non-ambiguous special characters
 *
 * Non-ambiguous special characters exclude visually confusable symbols:
 *   Excluded: | \ ` ' "  (look like l, 1, or quotes can be confused)
 *   Allowed:  ! @ # $ % ^ & * ( ) - _ = + [ ] { } : ; < > , . ? / ~
 */
export const UNAMBIGUOUS_SPECIALS = '!@#$%^&*()-_=+[]{}:;<>,.?/~'

export function countDigits(password) {
  return (password.match(/[0-9]/g) || []).length
}

export function countSpecials(password) {
  let count = 0
  for (const char of password) {
    if (UNAMBIGUOUS_SPECIALS.includes(char)) count++
  }
  return count
}

/**
 * Returns a criteria object with individual pass/fail states
 * and a top-level `valid` boolean.
 */
export function checkPasswordCriteria(password) {
  const digits = countDigits(password)
  const specials = countSpecials(password)

  const criteria = {
    length: password.length >= 20,
    digits: digits >= 2,
    specials: specials >= 2,
  }

  return {
    valid: criteria.length && criteria.digits && criteria.specials,
    criteria,
    // Counts for live feedback
    currentLength: password.length,
    currentDigits: digits,
    currentSpecials: specials,
  }
}

/**
 * Returns an array of human-readable error strings (empty if password is valid).
 */
export function getPasswordErrors(password) {
  const { criteria, currentLength, currentDigits, currentSpecials } = checkPasswordCriteria(password)
  const errors = []
  if (!criteria.length)
    errors.push(`Minimum 20 caractères (actuellement : ${currentLength})`)
  if (!criteria.digits)
    errors.push(`Au moins 2 chiffres (actuellement : ${currentDigits})`)
  if (!criteria.specials)
    errors.push(`Au moins 2 caractères spéciaux non ambigus (actuellement : ${currentSpecials})`)
  return errors
}
