import { toast } from 'vue-sonner'
import { HttpError } from '@/api/http'

export function useToast() {
  function success(message: string) {
    toast.success(message)
  }

  function error(message: string) {
    toast.error(message)
  }

  function errorFromApi(err: unknown) {
    if (err instanceof HttpError) {
      toast.error(err.apiError.message)
    } else if (err instanceof Error) {
      toast.error(err.message)
    } else {
      toast.error('An unknown error occurred')
    }
  }

  return { success, error, errorFromApi }
}
