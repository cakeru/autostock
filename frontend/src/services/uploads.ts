import api from './api'

// Uploads a photo (e.g. a supplier invoice) before the record that references
// it exists. Returns the URL to persist on the batch.
export const uploadsApi = {
  uploadImage: async (file: File): Promise<string> => {
    const form = new FormData()
    form.append('image', file)
    const res = await api.post('/uploads', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data.data.url as string
  },
}