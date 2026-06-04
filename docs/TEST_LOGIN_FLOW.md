# Test Login Flow

This document explains how to test the OIDC login flow for `idp-core` using Apidog and a browser.

## Overview

Use this split for reliable testing:

- **Apidog** starts the flow and tests backend APIs
- **Browser** handles the interactive Keycloak login page
- **Apidog** reuses the resulting session cookies for protected endpoints

This is the recommended approach because the OIDC Authorization Code flow depends on browser redirects, cookies, state, and an interactive Keycloak login form.

## Prerequisites

Make sure these services are running:

- Backend: `http://localhost:8989`
- Keycloak: `http://localhost:8081`
- Realm/client already seeded

Available local users:

- `platform-admin` / `admin123`
- `developer` / `dev123`

Use this redirect URI during testing so you do not need a frontend app running:

```text
http://localhost:8989/ping
```

This origin is already allowed in `configs/config.development.yaml`.

## Apidog Test Flow

## 1. Start login from Apidog

Create this request in Apidog:

### Request

**GET**

```text
http://localhost:8989/auth/oidc/login?redirect_uri=http://localhost:8989/ping
```

### Headers

```http
Accept: application/json
```

### Expected response

Status:

```http
200 OK
```

Response body shape:

```json
{
  "data": {
    "auth_url": "http://localhost:8081/realms/idp-core/protocol/openid-connect/auth?..."
  }
}
```

Also verify the response sets these cookies:

- `oidc_state`
- `oidc_redirect_uri`

## 2. Copy `auth_url`

From the Apidog response JSON, copy:

```text
data.auth_url
```

Open that URL in your browser.

## 3. Login in Keycloak

Use one of these credentials:

- `developer` / `dev123`
- `platform-admin` / `admin123`

After login, the backend should:

1. receive `/auth/oidc/callback`
2. verify the OIDC callback
3. exchange the authorization code for tokens
4. verify the ID token
5. create or update the local user
6. set cookies:
   - `auth_token`
   - `refresh_token`
7. redirect safely to:

```text
http://localhost:8989/ping
```

### Expected browser result

The browser should land on:

```text
http://localhost:8989/ping
```

and return a healthy response.

## 4. Verify cookies in browser

Open browser developer tools and inspect cookies for `http://localhost:8989`.

You should see:

- `auth_token`
- `refresh_token`

These cookies are `HttpOnly`, so they are not accessible from frontend JavaScript, but you can still inspect them in browser devtools.

## 5. Add cookies into Apidog

Use one of these approaches.

### Option A: Apidog cookie jar

If Apidog supports manual cookies for `localhost:8989`, add:

- `auth_token=<value>`
- `refresh_token=<value>`

### Option B: Raw Cookie header

Add this header manually:

```http
Cookie: auth_token=<JWT>; refresh_token=<REFRESH_TOKEN>
```

## 6. Test a protected endpoint in Apidog

### Request

**GET**

```text
http://localhost:8989/v1/users
```

### Headers

If you are not using Apidog's cookie jar:

```http
Cookie: auth_token=<JWT>; refresh_token=<REFRESH_TOKEN>
```

### Expected response

Status:

```http
200 OK
```

If this works, the backend OIDC login flow is working end-to-end.

## 7. Test refresh in Apidog

### Request

**POST**

```text
http://localhost:8989/auth/oidc/refresh
```

### Option 1: Use cookie only

No JSON body is required if `refresh_token` is already sent as a cookie.

Header:

```http
Cookie: refresh_token=<REFRESH_TOKEN>
```

### Option 2: Send JSON body

```json
{
  "refresh_token": "<REFRESH_TOKEN>"
}
```

### Expected response

Status:

```http
200 OK
```

Body shape:

```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "expires_in": 1234,
    "token_type": "Bearer",
    "user_id": "...",
    "email": "...",
    "is_admin": false
  }
}
```

The backend may also refresh cookies again.

## 8. Test logout in Apidog

### Request

**POST**

```text
http://localhost:8989/auth/oidc/logout
```

### Tip

Disable auto-follow redirects in Apidog if you want to inspect the initial backend response.

### Expected response

Status:

```http
302 Found
```

Expected headers include:

- `Location: http://localhost:8081/.../logout`
- clearing cookies for:
  - `auth_token`
  - `refresh_token`
  - `oidc_state`
  - `oidc_redirect_uri`

## Recommended Apidog Collection

Create these requests in order:

### 1. OIDC Login Start

**GET**

```text
{{BASE_URL}}/auth/oidc/login?redirect_uri=http://localhost:8989/ping
```

Headers:

```http
Accept: application/json
```

### 2. Protected Users

**GET**

```text
{{BASE_URL}}/v1/users
```

### 3. OIDC Refresh

**POST**

```text
{{BASE_URL}}/auth/oidc/refresh
```

### 4. OIDC Logout

**POST**

```text
{{BASE_URL}}/auth/oidc/logout
```

## Important Notes

Do **not** manually call `/auth/oidc/callback?...code=...&state=...` from Apidog after the browser already used it.

Reasons:

- the authorization code is one-time-use
- the state value must match the correct cookie
- Apidog and the browser usually do not share cookie state automatically

Use this reliable pattern instead:

- **Apidog** → request `/auth/oidc/login`
- **Browser** → complete Keycloak login
- **Apidog** → test protected APIs with the session cookies

## Pass/Fail Checklist

The backend login flow is working if all of these succeed:

- `/auth/oidc/login` returns a valid `auth_url`
- browser login succeeds
- browser lands on `/ping`
- `auth_token` cookie exists
- `refresh_token` cookie exists
- `/v1/users` works with those cookies
- `/auth/oidc/refresh` works
- `/auth/oidc/logout` clears cookies
