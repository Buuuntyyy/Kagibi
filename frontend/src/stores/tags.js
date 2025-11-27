import { defineStore } from 'pinia'
import api from '../api'

export const useTagStore = defineStore('tags', {
  state: () => ({
    tags: [],
  }),
  actions: {
    async fetchTags() {
      try {
        const response = await api.get('/tags/')
        this.tags = response.data || []
      } catch (error) {
        console.error('Error fetching tags:', error)
      }
    },
    async createTag(name, color) {
      try {
        const response = await api.post('/tags/', { name, color })
        this.tags.push(response.data)
        return response.data
      } catch (error) {
        console.error('Error creating tag:', error)
        throw error
      }
    },
    async deleteTag(id) {
      try {
        await api.delete(`/tags/${id}`)
        this.tags = this.tags.filter(t => t.id !== id)
      } catch (error) {
        console.error('Error deleting tag:', error)
        throw error
      }
    }
  }
})
