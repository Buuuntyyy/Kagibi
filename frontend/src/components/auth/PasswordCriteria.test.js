import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PasswordCriteria from './PasswordCriteria.vue'

describe('PasswordCriteria', () => {
  it('renders 4 strength segments when password is non-empty', () => {
    const wrapper = mount(PasswordCriteria, {
      props: { password: 'Ab1!', show: true },
    })
    const segments = wrapper.findAll('.strength-segment')
    expect(segments.length).toBe(4)
  })

  it('does not render segments when password is empty', () => {
    const wrapper = mount(PasswordCriteria, {
      props: { password: '', show: true },
    })
    expect(wrapper.find('.strength-segments').exists()).toBe(false)
  })

  it('shows 4 criteria list items', () => {
    const wrapper = mount(PasswordCriteria, {
      props: { password: '', show: true },
    })
    const items = wrapper.findAll('.criteria-list li')
    expect(items.length).toBe(4)
  })

  it('is visible when show=true', () => {
    const wrapper = mount(PasswordCriteria, {
      props: { password: 'test', show: true },
    })
    expect(wrapper.find('.password-criteria').isVisible()).toBe(true)
  })

  it('is hidden when show=false', () => {
    const wrapper = mount(PasswordCriteria, {
      props: { password: 'test', show: false },
    })
    expect(wrapper.find('.password-criteria').isVisible()).toBe(false)
  })

  it('marks length criterion as met when password >= 20 chars', () => {
    const wrapper = mount(PasswordCriteria, {
      props: { password: 'a'.repeat(20), show: true },
    })
    const items = wrapper.findAll('.criteria-list li')
    expect(items[0].classes()).toContain('met')
  })

  it('marks uppercase criterion (second item) as met when uppercase present', () => {
    const wrapper = mount(PasswordCriteria, {
      props: { password: 'A', show: true },
    })
    const items = wrapper.findAll('.criteria-list li')
    expect(items[1].classes()).toContain('met')
  })

  it('marks all segments green when all 4 criteria met', () => {
    const wrapper = mount(PasswordCriteria, {
      props: { password: 'Abcdefghijklmnopqrst1!', show: true },
    })
    const segments = wrapper.findAll('.strength-segment')
    expect(segments.every(s => s.classes().includes('segment-green'))).toBe(true)
  })

  it('marks 1 segment red when only 1 criterion met', () => {
    // Only uppercase met (1 criteria)
    const wrapper = mount(PasswordCriteria, {
      props: { password: 'A', show: true },
    })
    const segments = wrapper.findAll('.strength-segment')
    expect(segments[0].classes()).toContain('segment-red')
    expect(segments[1].classes()).toContain('segment-inactive')
    expect(segments[2].classes()).toContain('segment-inactive')
    expect(segments[3].classes()).toContain('segment-inactive')
  })

  it('marks 2 segments orange when 2 criteria met', () => {
    // uppercase + digit (2 criteria)
    const wrapper = mount(PasswordCriteria, {
      props: { password: 'A1', show: true },
    })
    const segments = wrapper.findAll('.strength-segment')
    expect(segments[0].classes()).toContain('segment-orange')
    expect(segments[1].classes()).toContain('segment-orange')
    expect(segments[2].classes()).toContain('segment-inactive')
    expect(segments[3].classes()).toContain('segment-inactive')
  })
})
