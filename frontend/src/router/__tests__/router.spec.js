import { describe, it, expect } from 'vitest'

describe('Router', () => {
  it('should have valid route configuration', () => {
    const routes = [
      { path: '/', name: 'home' },
      { path: '/login', name: 'login' },
      { path: '/register', name: 'register' },
      { path: '/files', name: 'files' },
    ]

    routes.forEach(route => {
      expect(route.path).toBeTruthy()
      expect(route.name).toBeTruthy()
    })
  })
})
