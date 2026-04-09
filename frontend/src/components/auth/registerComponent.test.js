import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import RegisterComponent from './registerComponent.vue'

// ── Stubs & Mocks ──────────────────────────────────────────────────────────

const AvatarSelectorStub = { template: '<div class="avatar-selector-stub" />' }
const PasswordCriteriaStub = {
  template: '<div class="password-criteria-stub" />',
  props: ['password', 'show'],
}

const mockRegister = vi.fn()
const mockEnsureRSAKeys = vi.fn().mockResolvedValue(undefined)

vi.mock('../../stores/auth', () => ({
  useAuthStore: () => ({
    register: mockRegister,
    ensureRSAKeys: mockEnsureRSAKeys,
    masterKey: null,
    isAuthenticated: false,
  }),
}))

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', name: 'Home', component: { template: '<div/>' } },
    { path: '/register', name: 'Register', component: { template: '<div/>' } },
  ],
})

function mountComponent() {
  return mount(RegisterComponent, {
    global: {
      plugins: [createPinia(), router],
      stubs: {
        AvatarSelector: AvatarSelectorStub,
        PasswordCriteria: PasswordCriteriaStub,
      },
    },
  })
}

async function fillAndSubmitForm(wrapper, { password = 'Abcdefghijklmnopqrst1!' } = {}) {
  await wrapper.find('input[type="text"]').setValue('TestUser')
  await wrapper.find('input[type="email"]').setValue('user@example.com')
  await wrapper.find('input[placeholder="••••••••"]').setValue(password)
  await wrapper.find('form').trigger('submit')
  await flushPromises()
}

// ── Tests ──────────────────────────────────────────────────────────────────

describe('registerComponent — registration form', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows duplicate email error when backend returns an error', async () => {
    mockRegister.mockRejectedValueOnce(new Error('Email already in use'))
    const wrapper = mountComponent()

    await fillAndSubmitForm(wrapper)

    expect(wrapper.find('.error-message').text()).toContain('Email already in use')
  })

  it('shows spinner on submit and hides after completion', async () => {
    mockRegister.mockResolvedValueOnce('testcode')
    const wrapper = mountComponent()

    // Before submit: no spinner
    expect(wrapper.find('.spinner').exists()).toBe(false)

    // Fill and start submit (don't await flushPromises yet)
    await wrapper.find('input[type="text"]').setValue('TestUser')
    await wrapper.find('input[type="email"]').setValue('user@example.com')
    await wrapper.find('input[placeholder="••••••••"]').setValue('Abcdefghijklmnopqrst1!')
    wrapper.find('form').trigger('submit')

    await flushPromises()

    // After completion: recovery code shown, no form
    expect(wrapper.find('.recovery-display').exists()).toBe(true)
  })
})

describe('registerComponent — recovery code display', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  async function mountWithRecoveryCode(code = 'testrecoverycode1234567890abcdef') {
    mockRegister.mockResolvedValueOnce(code)
    const wrapper = mountComponent()
    await fillAndSubmitForm(wrapper)
    return wrapper
  }

  it('Continue button is disabled until checkbox is checked', async () => {
    const wrapper = await mountWithRecoveryCode()

    const continueBtn = wrapper.find('.btn-submit')
    // Initially disabled (checkbox unchecked)
    expect(continueBtn.attributes('disabled')).toBeDefined()

    // Check the checkbox
    const checkbox = wrapper.find('#copied-checkbox')
    await checkbox.setValue(true)

    // Now enabled
    expect(continueBtn.attributes('disabled')).toBeUndefined()
  })

  it('copy button does NOT enable the Continue button on its own', async () => {
    const wrapper = await mountWithRecoveryCode()

    // Mock clipboard
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: vi.fn().mockResolvedValue(undefined) },
      configurable: true,
      writable: true,
    })

    // Click copy button — should NOT enable Continue
    await wrapper.find('.btn-secondary').trigger('click')
    await flushPromises()

    expect(wrapper.find('.btn-submit').attributes('disabled')).toBeDefined()
  })

  it('recovery display has role="alertdialog"', async () => {
    const wrapper = await mountWithRecoveryCode()
    expect(wrapper.find('.recovery-display').attributes('role')).toBe('alertdialog')
  })

  it('checkbox with correct id and label exists', async () => {
    const wrapper = await mountWithRecoveryCode()
    expect(wrapper.find('#copied-checkbox').exists()).toBe(true)
    expect(wrapper.find('#copied-checkbox').attributes('type')).toBe('checkbox')
    expect(wrapper.find('label[for="copied-checkbox"]').exists()).toBe(true)
  })

  it('displays the recovery code in the code-box', async () => {
    const code = 'myrecoverycode1234567890abcdefxy'
    const wrapper = await mountWithRecoveryCode(code)
    expect(wrapper.find('.code-box').text()).toContain(code)
  })
})
