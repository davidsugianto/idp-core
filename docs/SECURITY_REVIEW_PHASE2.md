# Security Review Report - Phase 2

**Date:** May 2026
**Scope:** OIDC, RBAC, API Keys, Audit Log, and general security considerations

---

## Executive Summary

The Phase 2 implementation has been reviewed for security vulnerabilities. Overall, the codebase follows security best practices with a few areas that warrant attention.

**Summary:**
- Critical Issues: 0
- High Issues: 2
- Medium Issues: 3
- Low Issues: 4

---

## Findings

### 1. OIDC Implementation

#### 1.1 Token Verification - ✅ SECURE

**Location:** `internal/pkg/oidc/client.go:78-80`

The implementation uses the official `go-oidc` library which properly validates:
- Token signature against JWKS
- Token expiration (`exp` claim)
- Issuer (`iss` claim)
- Audience (`aud` claim)

```go
func (c *Client) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
    return c.verifier.Verify(ctx, rawIDToken)
}
```

**Status:** No issues found.

#### 1.2 Missing Token Replay Protection - MEDIUM

**Location:** `internal/pkg/oidc/verifier.go:56-63`

The implementation does not check for token replay attacks. While the OIDC library checks expiration, a token could be replayed within its validity window.

**Recommendation:** Implement a token nonce check or use short-lived tokens with refresh token rotation.

#### 1.3 Group Claim Injection - LOW

**Location:** `internal/pkg/oidc/verifier.go:112-135`

Groups are extracted from the token claims without validation. A malicious OIDC provider could inject unexpected group values.

**Recommendation:** Validate group names against a whitelist or sanitize before using in RBAC decisions.

---

### 2. RBAC Implementation

#### 2.1 Privilege Escalation Risk - HIGH

**Location:** `internal/usecase/auth/rbac.go:23-41`

The `CheckPermission` function checks for both specific action permission and `manage` permission. However, there's no check to prevent users with `manage` permission from assigning roles they don't have themselves.

**Recommendation:** Add a constraint that users can only assign roles/permissions they possess:
```go
// Before assigning a role, verify the assigner has all permissions in that role
func (e *RBACEngine) CanAssignRole(ctx context.Context, assignerID, roleID string) (bool, error) {
    // Get role permissions
    rolePerms, err := e.roleRepo.GetRolePermissions(ctx, roleID)
    // Check if assigner has all of them
    for _, perm := range rolePerms {
        has, err := e.CheckPermission(ctx, assignerID, perm.Resource, perm.Action)
        if !has || err != nil {
            return false, err
        }
    }
    return true, nil
}
```

#### 2.2 Team Isolation Verification - MEDIUM

**Location:** `internal/usecase/auth/rbac.go:44-62`

The `CheckTeamPermission` properly checks team-scoped permissions, but the team ID is passed as a parameter without verification that the user actually belongs to that team.

**Recommendation:** Add explicit team membership check before permission check:
```go
func (e *RBACEngine) CheckTeamPermission(ctx context.Context, userID, teamID, resource, action string) (bool, error) {
    // First verify user is a member of the team
    isMember, err := e.roleRepo.IsTeamMember(ctx, userID, teamID)
    if err != nil || !isMember {
        return false, err
    }
    // Then check permissions...
}
```

---

### 3. API Key Implementation

#### 3.1 Key Generation Entropy - ✅ SECURE

**Location:** `internal/usecase/apikey/apikey.go:31-38`

Key generation uses `crypto/rand` which provides cryptographically secure random bytes:

```go
func generateKey() (string, error) {
    bytes := make([]byte, keyLength)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("failed to generate random key: %w", err)
    }
    return keyPrefix + hex.EncodeToString(bytes), nil
}
```

This produces 40 bytes of entropy (320 bits), which is sufficient for security.

**Status:** No issues found.

#### 3.2 Key Hashing - ✅ SECURE

**Location:** `internal/usecase/apikey/apikey.go:40-44`

Keys are hashed with SHA-256 before storage:

```go
func hashKey(key string) string {
    hash := sha256.Sum256([]byte(key))
    return hex.EncodeToString(hash[:])
}
```

While SHA-256 is not a password hashing algorithm, it's acceptable for API keys because:
- API keys have high entropy (320 bits)
- They are randomly generated, not user-chosen
- SHA-256 is collision-resistant

**Status:** No issues found.

#### 3.3 Timing Attack on Validation - LOW

**Location:** `internal/usecase/apikey/apikey.go:170-194`

The key validation hashes the input and compares against stored hash. While the database lookup may be constant-time, the hash comparison depends on the database implementation.

**Recommendation:** Use `crypto/subtle.ConstantTimeCompare` for hash comparison if doing in-memory comparison.

#### 3.4 Key Exposure in Response - ✅ SECURE

**Location:** `internal/usecase/apikey/apikey.go:86-87`

The plain key is only returned once during creation and never stored:

```go
resp := apikey.ToAPIKeyResponse(key, true)
resp.Key = plainKey // return the plain key once
```

Subsequent calls to `Get()` return `ToAPIKeyResponse(key, false)` which does not include the key.

**Status:** No issues found.

---

### 4. Audit Log Implementation

#### 4.1 Sensitive Data Logging - MEDIUM

**Location:** `internal/usecase/auditlog/auditlog.go:11-42`

The audit log stores `OldValues` and `NewValues` as JSON without filtering. This could potentially log sensitive data like passwords, API keys, or tokens if passed in.

**Recommendation:** Add a sensitive field filter:
```go
var sensitiveFields = []string{"password", "secret", "token", "api_key", "client_secret"}

func sanitizeValues(values auditlog.Map) auditlog.Map {
    if values == nil {
        return nil
    }
    result := make(auditlog.Map)
    for k, v := range values {
        if isSensitive(k) {
            result[k] = "[REDACTED]"
        } else {
            result[k] = v
        }
    }
    return result
}
```

#### 4.2 Log Injection - LOW

**Location:** `internal/usecase/auditlog/auditlog.go:11-42`

User-provided fields like `UserAgent`, `RequestPath`, and `ErrorMessage` are stored without sanitization. When logs are displayed, this could allow log injection if not properly escaped at display time.

**Recommendation:** Sanitize or escape user-controlled fields before storage or display.

---

### 5. General Security Considerations

#### 5.1 Input Validation - ✅ GOOD

The codebase uses Gin's `ShouldBindJSON` for input parsing which handles malformed JSON safely. Request structs use appropriate types (string, int, time.Time) rather than interface{}.

**Status:** Good practices followed.

#### 5.2 SQL Injection - ✅ SECURE

The codebase uses GORM which parameterizes queries by default. All database queries go through the repository layer which uses GORM's query builder.

**Status:** No issues found.

#### 5.3 Authorization Bypass - HIGH

**Location:** Handler layer (not reviewed in detail)

Each handler should verify authorization before processing requests. Ensure all endpoints have proper RBAC checks and team isolation.

**Recommendation:** Audit all HTTP handlers to ensure:
1. JWT/API key validation is performed
2. RBAC check is done before processing
3. Team isolation is enforced for team-scoped resources
4. No orphaned endpoints without auth middleware

#### 5.4 Error Information Leakage - LOW

**Location:** Various usecase files

Error messages returned to users may contain internal details (e.g., "failed to connect to database"). While helpful for debugging, this can leak internal architecture.

**Recommendation:** Use a centralized error handler that logs detailed errors but returns generic messages to users.

---

## Recommendations Summary

### High Priority

1. **Privilege Escalation Prevention:** Add constraint that users can only assign roles with permissions they possess.

2. **Authorization Audit:** Review all HTTP handlers to ensure proper RBAC checks are in place.

### Medium Priority

3. **Token Replay Protection:** Implement nonce checking or use short-lived tokens.

4. **Team Membership Verification:** Verify team membership before checking team permissions.

5. **Sensitive Data Filtering:** Filter sensitive fields from audit log storage.

### Low Priority

6. **Group Validation:** Validate group names from OIDC provider.

7. **Timing Attack Mitigation:** Use constant-time comparison for key validation.

8. **Log Injection Prevention:** Sanitize user-controlled audit log fields.

9. **Error Message Sanitization:** Return generic error messages to users.

---

## Security Best Practices Followed

✅ Cryptographically secure random number generation for API keys
✅ SHA-256 hashing for API key storage
✅ Proper OIDC token verification using official library
✅ API keys only returned once during creation
✅ GORM for SQL injection prevention
✅ Proper type binding for input validation
✅ Team-scoped RBAC permissions

---

## Conclusion

The Phase 2 implementation demonstrates good security practices overall. The most critical areas requiring attention are:

1. **Privilege escalation prevention** - Users should not be able to assign roles with permissions they don't have.
2. **Comprehensive authorization audit** - Ensure all endpoints have proper auth checks.

These should be addressed before production deployment. The medium and low priority items can be addressed in a subsequent security hardening pass.
