import { useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useAcceptInvite } from '@/hooks/useInvites'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface AcceptFields {
  displayName: string
  password: string
  confirmPassword: string
}

export function InviteAcceptPage() {
  const { token } = useParams<{ token: string }>()
  const navigate = useNavigate()
  const [errorMsg, setErrorMsg] = useState<string | null>(null)
  const acceptInvite = useAcceptInvite()

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<AcceptFields>()

  const password = useWatch({ control, name: 'password' })

  function onSubmit(data: AcceptFields) {
    if (!token) return
    setErrorMsg(null)
    acceptInvite.mutate(
      { token, password: data.password, displayName: data.displayName },
      {
        onSuccess: () => navigate('/login'),
        onError: () => setErrorMsg('This invite link is invalid or has expired.'),
      },
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="w-full max-w-sm space-y-6 px-4">
        <div className="space-y-1 text-center">
          <h1 className="text-2xl font-semibold">Accept Invite</h1>
          <p className="text-muted-foreground text-sm">
            Set your password to complete your account setup.
          </p>
        </div>

        {errorMsg && (
          <div
            role="alert"
            className="rounded-md border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive"
          >
            {errorMsg}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
          <div className="space-y-1">
            <Label htmlFor="displayName">Display Name</Label>
            <Input
              id="displayName"
              type="text"
              autoComplete="name"
              aria-invalid={!!errors.displayName}
              {...register('displayName', {
                required: 'Display name is required',
                maxLength: { value: 100, message: 'Display name must be at most 100 characters' },
              })}
            />
            {errors.displayName && (
              <p className="text-destructive text-xs">{errors.displayName.message}</p>
            )}
          </div>

          <div className="space-y-1">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              autoComplete="new-password"
              aria-invalid={!!errors.password}
              {...register('password', {
                required: 'Password is required',
                minLength: { value: 8, message: 'Password must be at least 8 characters' },
              })}
            />
            {errors.password && (
              <p className="text-destructive text-xs">{errors.password.message}</p>
            )}
          </div>

          <div className="space-y-1">
            <Label htmlFor="confirmPassword">Confirm Password</Label>
            <Input
              id="confirmPassword"
              type="password"
              autoComplete="new-password"
              aria-invalid={!!errors.confirmPassword}
              {...register('confirmPassword', {
                required: 'Please confirm your password',
                validate: (value) => value === password || 'Passwords do not match',
              })}
            />
            {errors.confirmPassword && (
              <p className="text-destructive text-xs">{errors.confirmPassword.message}</p>
            )}
          </div>

          <Button type="submit" className="w-full" disabled={acceptInvite.isPending}>
            {acceptInvite.isPending ? 'Creating account…' : 'Create Account'}
          </Button>
        </form>

        <p className="text-muted-foreground text-center text-sm">
          Already have an account?{' '}
          <Link to="/login" className="text-primary underline-offset-4 hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  )
}
