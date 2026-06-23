import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useNavigate, Navigate, Link } from 'react-router-dom';
import { api } from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

interface LoginFields {
  email: string;
  password: string;
  rememberMe: boolean;
}

interface LoginResponse {
  accessToken: string;
}

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const { data: setupData, isLoading: setupLoading } = useQuery({
    queryKey: ['auth-setup'],
    queryFn: () => api.get<{ adminExists: boolean }>('/auth/setup').then((response) => response.data),
    staleTime: 30_000,
  });

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFields>();

  const mutation = useMutation({
    mutationFn: (data: LoginFields) => api.post<LoginResponse>('/auth/login', data).then((response) => response.data),
    onSuccess: (data) => {
      login(data.accessToken);
      navigate('/admin');
    },
    onError: () => {
      setErrorMsg('Invalid email or password.');
    },
  });

  if (setupLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground text-sm">Loading…</p>
      </div>
    );
  }

  if (setupData && !setupData.adminExists) {
    return <Navigate to="/register" replace />;
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="w-full max-w-sm space-y-6 px-4">
        <div className="space-y-1 text-center">
          <h1 className="text-2xl font-semibold">Sign in</h1>
          <p className="text-muted-foreground text-sm">Enter your credentials to continue</p>
        </div>

        {errorMsg && (
          <div role="alert" className="border-destructive/50 bg-destructive/10 text-destructive rounded-md border px-4 py-3 text-sm">
            {errorMsg}
          </div>
        )}

        <form onSubmit={handleSubmit((data) => mutation.mutate(data))} className="space-y-4" noValidate>
          <div className="space-y-1">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              aria-invalid={!!errors.email}
              {...register('email', {
                required: 'Email is required',
                pattern: { value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: 'Enter a valid email address' },
              })}
            />
            {errors.email && <p className="text-destructive text-xs">{errors.email.message}</p>}
          </div>

          <div className="space-y-1">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              aria-invalid={!!errors.password}
              {...register('password', {
                required: 'Password is required',
                minLength: { value: 8, message: 'Password must be at least 8 characters' },
              })}
            />
            {errors.password && <p className="text-destructive text-xs">{errors.password.message}</p>}
          </div>

          <div className="flex items-center gap-2">
            <input id="rememberMe" type="checkbox" className="border-border h-4 w-4 rounded" {...register('rememberMe')} />
            <Label htmlFor="rememberMe" className="cursor-pointer text-sm font-normal">
              Stay logged in
            </Label>
          </div>

          <Button type="submit" className="w-full" disabled={mutation.isPending}>
            {mutation.isPending ? 'Signing in…' : 'Sign in'}
          </Button>
        </form>

        <p className="text-muted-foreground text-center text-sm">
          Don&apos;t have an account?{' '}
          <Link to="/register" className="text-primary underline-offset-4 hover:underline">
            Create account
          </Link>
        </p>
      </div>
    </div>
  );
}
