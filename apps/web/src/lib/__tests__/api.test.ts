import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import MockAdapter from 'axios-mock-adapter';
import { api, getAccessToken, setAccessToken } from '@/lib/api';

let mock: MockAdapter;

beforeEach(() => {
  mock = new MockAdapter(api);
  setAccessToken(null);
});

afterEach(() => {
  mock.restore();
});

describe('api baseURL — no prefix added to explicit paths', () => {
  it('auth request full URL must be /auth/register, not /api/auth/register', async () => {
    let capturedFullUrl: string | undefined;

    mock.onAny().reply((config) => {
      // Reproduce what axios does in the browser: merge baseURL + url
      const base = (config.baseURL ?? '').replace(/\/$/, '');
      const path = (config.url ?? '').replace(/^\//, '');
      capturedFullUrl = base ? `${base}/${path}` : `/${path}`;
      return [201, {}];
    });

    await api.post('/auth/register', { email: 'a@b.com', password: '12345678' });
    expect(capturedFullUrl).toBe('/auth/register');
  });
});

describe('api request interceptor', () => {
  it('attaches Authorization header when access token is set', async () => {
    setAccessToken('test-token');
    mock.onGet('/ping').reply(() => {
      return [200, { ok: true }];
    });

    const response = await api.get('/ping');
    expect(response.config.headers?.Authorization).toBe('Bearer test-token');
  });

  it('does not attach Authorization header when no token', async () => {
    mock.onGet('/ping').reply(() => {
      return [200, { ok: true }];
    });

    const response = await api.get('/ping');
    expect(response.config.headers?.Authorization).toBeUndefined();
  });
});

describe('api 401 response interceptor', () => {
  it('calls POST /auth/refresh and retries original request on 401', async () => {
    setAccessToken('expired-token');
    let callCount = 0;

    mock.onPost('/auth/refresh').reply(200, { accessToken: 'new-token' });
    mock.onGet('/protected').reply(() => {
      callCount++;
      if (callCount === 1) return [401, { message: 'Unauthorized' }];
      return [200, { data: 'secret' }];
    });

    const response = await api.get('/protected');
    expect(response.status).toBe(200);
    expect(response.data).toEqual({ data: 'secret' });
    expect(getAccessToken()).toBe('new-token');
    expect(callCount).toBe(2);
  });

  it('does not retry more than once (prevents infinite loop)', async () => {
    setAccessToken('expired-token');
    let refreshCallCount = 0;

    mock.onPost('/auth/refresh').reply(() => {
      refreshCallCount++;
      return [200, { accessToken: 'new-token' }];
    });
    mock.onGet('/protected').reply(401, { message: 'Unauthorized' });

    await expect(api.get('/protected')).rejects.toMatchObject({ response: { status: 401 } });
    expect(refreshCallCount).toBe(1);
  });

  it('rejects with 401 error when refresh itself fails', async () => {
    setAccessToken('expired-token');

    mock.onPost('/auth/refresh').reply(401, { message: 'Refresh token expired' });
    mock.onGet('/protected').reply(401, { message: 'Unauthorized' });

    await expect(api.get('/protected')).rejects.toMatchObject({ response: { status: 401 } });
    expect(getAccessToken()).toBeNull();
  });
});
