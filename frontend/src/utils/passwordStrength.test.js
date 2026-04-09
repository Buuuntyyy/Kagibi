import { describe, it, expect } from 'vitest'
import {
  checkPasswordCriteria,
  getPasswordErrors,
  countDigits,
  countSpecials,
  UNAMBIGUOUS_SPECIALS,
} from './passwordStrength'

describe('countDigits', () => {
  it('counts digits correctly', () => {
    expect(countDigits('abc123')).toBe(3)
    expect(countDigits('no digits')).toBe(0)
    expect(countDigits('1')).toBe(1)
  })
})

describe('countSpecials', () => {
  it('counts unambiguous special characters', () => {
    expect(countSpecials('hello!')).toBe(1)
    expect(countSpecials('a@b#c')).toBe(2)
    expect(countSpecials('no specials')).toBe(0)
  })

  it('does not count ambiguous characters (pipe, backslash, backtick, quotes)', () => {
    expect(countSpecials('|')).toBe(0)
    expect(countSpecials('\\')).toBe(0)
    expect(countSpecials('`')).toBe(0)
    expect(countSpecials("'")).toBe(0)
    expect(countSpecials('"')).toBe(0)
  })
})

describe('checkPasswordCriteria', () => {
  it('length criterion passes at exactly 20 characters', () => {
    expect(checkPasswordCriteria('a'.repeat(20)).criteria.length).toBe(true)
    expect(checkPasswordCriteria('a'.repeat(19)).criteria.length).toBe(false)
  })

  it('uppercase criterion passes when at least 1 A-Z present', () => {
    expect(checkPasswordCriteria('A').criteria.uppercase).toBe(true)
    expect(checkPasswordCriteria('Z').criteria.uppercase).toBe(true)
    expect(checkPasswordCriteria('abc123').criteria.uppercase).toBe(false)
    expect(checkPasswordCriteria('').criteria.uppercase).toBe(false)
  })

  it('digits criterion passes when at least 1 digit present', () => {
    expect(checkPasswordCriteria('1').criteria.digits).toBe(true)
    expect(checkPasswordCriteria('abc').criteria.digits).toBe(false)
  })

  it('specials criterion passes when at least 1 unambiguous special present', () => {
    expect(checkPasswordCriteria('!').criteria.specials).toBe(true)
    expect(checkPasswordCriteria('@').criteria.specials).toBe(true)
    expect(checkPasswordCriteria('abc1A').criteria.specials).toBe(false)
  })

  it('all 4 criteria pass and fail independently', () => {
    // Only uppercase
    const r1 = checkPasswordCriteria('A')
    expect(r1.criteria.uppercase).toBe(true)
    expect(r1.criteria.length).toBe(false)
    expect(r1.criteria.digits).toBe(false)
    expect(r1.criteria.specials).toBe(false)

    // Only digit
    const r2 = checkPasswordCriteria('1')
    expect(r2.criteria.digits).toBe(true)
    expect(r2.criteria.uppercase).toBe(false)

    // Only special
    const r3 = checkPasswordCriteria('!')
    expect(r3.criteria.specials).toBe(true)
    expect(r3.criteria.uppercase).toBe(false)

    // Only length
    const r4 = checkPasswordCriteria('a'.repeat(20))
    expect(r4.criteria.length).toBe(true)
    expect(r4.criteria.uppercase).toBe(false)
  })

  it('valid requires all 4 criteria simultaneously', () => {
    // All 4 met
    expect(checkPasswordCriteria('Abcdefghijklmnopqrst1!').valid).toBe(true)

    // Missing uppercase
    expect(checkPasswordCriteria('abcdefghijklmnopqrst1!').valid).toBe(false)

    // Missing digit
    expect(checkPasswordCriteria('Abcdefghijklmnopqrst!').valid).toBe(false)

    // Missing special
    expect(checkPasswordCriteria('Abcdefghijklmnopqrst1').valid).toBe(false)

    // Too short (all others met)
    expect(checkPasswordCriteria('Ab1!').valid).toBe(false)
  })

  it('returns currentLength, currentDigits, currentSpecials counts', () => {
    const result = checkPasswordCriteria('Hello12!')
    expect(result.currentLength).toBe(8)
    expect(result.currentDigits).toBe(2)
    expect(result.currentSpecials).toBe(1)
  })
})

describe('getPasswordErrors', () => {
  it('returns empty array for a valid password', () => {
    expect(getPasswordErrors('Abcdefghijklmnopqrst1!')).toHaveLength(0)
  })

  it('returns error for missing length', () => {
    const errors = getPasswordErrors('Ab1!')
    expect(errors.some(e => e.includes('20'))).toBe(true)
  })

  it('returns error for missing uppercase', () => {
    const errors = getPasswordErrors('abcdefghijklmnopqrst1!')
    expect(errors.some(e => e.toLowerCase().includes('majuscule'))).toBe(true)
  })

  it('returns error for missing digit', () => {
    const errors = getPasswordErrors('Abcdefghijklmnopqrst!')
    expect(errors.some(e => e.includes('chiffre'))).toBe(true)
  })

  it('returns error for missing special', () => {
    const errors = getPasswordErrors('Abcdefghijklmnopqrst1')
    expect(errors.some(e => e.includes('spécial'))).toBe(true)
  })

  it('can return multiple errors at once', () => {
    const errors = getPasswordErrors('')
    expect(errors.length).toBeGreaterThan(1)
  })
})
