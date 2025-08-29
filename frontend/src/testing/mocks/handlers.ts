import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/users/me', () => {
    return HttpResponse.json({
      id: 1,
      name: 'John Doe',
      email: 'john.doe@example.com',
    })
  }),
]
