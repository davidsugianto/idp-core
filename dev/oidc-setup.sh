#!/bin/bash
set -e

KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8081}"
KEYCLOAK_ADMIN="${KEYCLOAK_ADMIN:-admin}"
KEYCLOAK_ADMIN_PASSWORD="${KEYCLOAK_ADMIN_PASSWORD:-admin}"

echo "==> Setting up Keycloak OIDC configuration..."
echo ""

# Step 1: Wait for Keycloak to be ready
echo "Step 1: Waiting for Keycloak to be ready..."
for i in $(seq 1 30); do
    if curl -sf "${KEYCLOAK_URL}/realms/master" >/dev/null 2>&1; then
        break
    fi
    sleep 2
done

if ! curl -sf "${KEYCLOAK_URL}/realms/master" >/dev/null 2>&1; then
    echo "❌ Keycloak is not responding at ${KEYCLOAK_URL}"
    echo "   Make sure Keycloak is running: make dev-oidc-up"
    exit 1
fi

echo "✅ Keycloak is ready"
echo ""

# Step 2: Get admin token
echo "Step 2: Getting admin token..."
TOKEN_RESPONSE=$(curl -s "${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token" \
    -d "client_id=admin-cli" \
    -d "username=${KEYCLOAK_ADMIN}" \
    -d "password=${KEYCLOAK_ADMIN_PASSWORD}" \
    -d "grant_type=password" 2>&1)

# Debug: show response if it fails
TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r .access_token 2>/dev/null)

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "❌ Failed to get admin token"
    echo "   Response: $TOKEN_RESPONSE"
    echo ""
    echo "   Troubleshooting:"
    echo "   1. Check Keycloak is running: docker-compose ps keycloak"
    echo "   2. Check Keycloak logs: make dev-oidc-logs"
    echo "   3. Verify admin credentials in docker-compose.yml"
    exit 1
fi

echo "✅ Got admin token"
echo ""

# Step 3: Create realm
echo "Step 3: Creating idp-core realm..."
REALM_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${KEYCLOAK_URL}/admin/realms" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "realm": "idp-core",
        "enabled": true,
        "displayName": "IDP Core",
        "registrationAllowed": false,
        "accessTokenLifespan": 3600,
        "ssoSessionMaxLifespan": 86400
    }' 2>&1)

REALM_HTTP_CODE=$(echo "$REALM_RESPONSE" | tail -1)
if [ "$REALM_HTTP_CODE" = "201" ]; then
    echo "✅ Realm created"
elif [ "$REALM_HTTP_CODE" = "409" ]; then
    echo "   (realm already exists)"
else
    echo "   Warning: Unexpected response code $REALM_HTTP_CODE"
fi
echo ""

# Step 4: Create client
echo "Step 4: Creating idp-core client..."
CLIENT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${KEYCLOAK_URL}/admin/realms/idp-core/clients" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "clientId": "idp-core",
        "name": "IDP Core Application",
        "enabled": true,
        "clientAuthenticatorType": "client-secret",
        "secret": "idp-core-secret-key",
        "redirectUris": ["http://localhost:8080/auth/callback", "http://localhost:8989/auth/callback"],
        "webOrigins": ["http://localhost:8080", "http://localhost:8989"],
        "standardFlowEnabled": true,
        "directAccessGrantsEnabled": true,
        "protocol": "openid-connect",
        "attributes": {"access.token.lifespan": "3600"},
        "defaultClientScopes": ["openid", "profile", "email"]
    }' 2>&1)

CLIENT_HTTP_CODE=$(echo "$CLIENT_RESPONSE" | tail -1)
if [ "$CLIENT_HTTP_CODE" = "201" ]; then
    echo "✅ Client created"
elif [ "$CLIENT_HTTP_CODE" = "409" ]; then
    echo "   (client already exists)"
else
    echo "   Warning: Unexpected response code $CLIENT_HTTP_CODE"
fi
echo ""

# Step 4b: Add group membership mapper to client
echo "Step 4b: Adding group membership mapper to idp-core client..."
CLIENT_UUID=$(curl -s "${KEYCLOAK_URL}/admin/realms/idp-core/clients?clientId=idp-core" \
    -H "Authorization: Bearer $TOKEN" | jq -r '.[0].id' 2>/dev/null)

if [ -n "$CLIENT_UUID" ] && [ "$CLIENT_UUID" != "null" ]; then
    MAPPER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
        "${KEYCLOAK_URL}/admin/realms/idp-core/clients/${CLIENT_UUID}/protocol-mappers/models" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "group-membership",
            "protocol": "openid-connect",
            "protocolMapper": "oidc-group-membership-mapper",
            "consentRequired": false,
            "config": {
                "full.path": "false",
                "id.token.claim": "true",
                "access.token.claim": "true",
                "claim.name": "groups"
            }
        }' 2>&1)

    MAPPER_HTTP_CODE=$(echo "$MAPPER_RESPONSE" | tail -1)
    if [ "$MAPPER_HTTP_CODE" = "201" ]; then
        echo "✅ Group membership mapper added"
    elif [ "$MAPPER_HTTP_CODE" = "409" ]; then
        echo "   (mapper already exists)"
    else
        echo "   Warning: Unexpected response code $MAPPER_HTTP_CODE"
        echo "   Response: $(echo "$MAPPER_RESPONSE" | head -1)"
    fi
else
    echo "   Warning: Could not find client UUID for idp-core"
fi
echo ""

# Step 5: Create platform-admins group
echo "Step 5: Creating platform-admins group..."
GROUP_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${KEYCLOAK_URL}/admin/realms/idp-core/groups" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name": "platform-admins"}' 2>&1)

GROUP_HTTP_CODE=$(echo "$GROUP_RESPONSE" | tail -1)
if [ "$GROUP_HTTP_CODE" = "201" ]; then
    echo "✅ Group created"
elif [ "$GROUP_HTTP_CODE" = "409" ]; then
    echo "   (group already exists)"
else
    echo "   Warning: Unexpected response code $GROUP_HTTP_CODE"
fi
echo ""

# Get the group ID for adding users
GROUP_ID=$(curl -s "${KEYCLOAK_URL}/admin/realms/idp-core/groups" \
    -H "Authorization: Bearer $TOKEN" | jq -r '.[] | select(.name=="platform-admins") | .id' 2>/dev/null)

# Step 6: Create test user (platform-admin)
echo "Step 6: Creating test user (platform-admin)..."
USER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${KEYCLOAK_URL}/admin/realms/idp-core/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "platform-admin",
        "enabled": true,
        "email": "admin@example.com",
        "firstName": "Platform",
        "lastName": "Admin",
        "credentials": [{"type": "password", "value": "admin123", "temporary": false}]
    }' 2>&1)

USER_HTTP_CODE=$(echo "$USER_RESPONSE" | tail -1)
if [ "$USER_HTTP_CODE" = "201" ]; then
    echo "✅ Test user created"
    # Get user ID and add to group
    USER_ID=$(curl -s "${KEYCLOAK_URL}/admin/realms/idp-core/users?username=platform-admin" \
        -H "Authorization: Bearer $TOKEN" | jq -r '.[0].id' 2>/dev/null)
    if [ -n "$USER_ID" ] && [ -n "$GROUP_ID" ]; then
        curl -s -X PUT "${KEYCLOAK_URL}/admin/realms/idp-core/users/${USER_ID}/groups/${GROUP_ID}" \
            -H "Authorization: Bearer $TOKEN" >/dev/null 2>&1
        echo "   Added to platform-admins group"
    fi
elif [ "$USER_HTTP_CODE" = "409" ]; then
    echo "   (user already exists)"
else
    echo "   Warning: Unexpected response code $USER_HTTP_CODE"
fi
echo ""

# Step 7: Create test user (developer)
echo "Step 7: Creating test user (developer)..."
DEV_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${KEYCLOAK_URL}/admin/realms/idp-core/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "developer",
        "enabled": true,
        "email": "developer@example.com",
        "firstName": "Dev",
        "lastName": "Eloper",
        "credentials": [{"type": "password", "value": "dev123", "temporary": false}]
    }' 2>&1)

DEV_HTTP_CODE=$(echo "$DEV_RESPONSE" | tail -1)
if [ "$DEV_HTTP_CODE" = "201" ]; then
    echo "✅ Developer user created"
elif [ "$DEV_HTTP_CODE" = "409" ]; then
    echo "   (user already exists)"
else
    echo "   Warning: Unexpected response code $DEV_HTTP_CODE"
fi
echo ""

echo "✅ Keycloak OIDC setup complete!"
echo ""
echo "=== Configuration Summary ==="
echo "Issuer URL:         ${KEYCLOAK_URL}/realms/idp-core"
echo "Client ID:          idp-core"
echo "Client Secret:      idp-core-secret-key"
echo "Redirect URI:       http://localhost:8080/auth/callback"
echo "Admin Group:        platform-admins"
echo ""
echo "Test Users:"
echo "  platform-admin / admin123 (member of platform-admins group)"
echo "  developer / dev123"
echo ""
echo "Add to configs/config.development.yaml:"
echo "  auth:"
echo "    oidc:"
echo "      enabled: true"
echo "      issuer_url: ${KEYCLOAK_URL}/realms/idp-core"
echo "      client_id: idp-core"
echo "      client_secret: idp-core-secret-key"
echo "      redirect_url: http://localhost:8080/auth/callback"
echo "      groups_claim: groups"
echo "      admin_group: platform-admins"
