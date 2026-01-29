import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

describe('Store Tests', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('should initialize pinia', () => {
    const pinia = createPinia()
    expect(pinia).toBeDefined()
  })
})
