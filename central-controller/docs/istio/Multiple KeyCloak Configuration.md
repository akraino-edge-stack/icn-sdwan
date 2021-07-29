## Multiple KeyCloak Services Configuration

In many user cases, we may need different KeyCloak service for different enterprise. They themselves maintain their own authentication. So we simply list how to configure two independent KeyCloak services to do authentication on same overlay controller.

![arch](https://github.com/akraino-edge-stack/icn-sdwan/tree/master/central-controller/docs/istio/mutl-key.png)

### Create different namespace

```shell
# We will create two namespaces
kubectl create ns k1
kubectl create ns k2
```



### Deploy KeyCloak in different namespace

```shell
# Create secret for each KeyCloak service or use your own secret
openssl genrsa 2048 > k1.key
openssl req -new -x509 -nodes -sha256 -days 365 -key k1.key -out k1.crt

openssl genrsa 2048 > k2.key
openssl req -new -x509 -nodes -sha256 -days 365 -key k2.key -out k2.crt

kubectl create -n k1 secret tls ca-keycloak-certs --key k1.key --cert k1.crt
kubectl create -n k2 secret tls ca-keycloak-certs --key k2.key --cert k2.crt

# We will deploy two keycloak services into these two namespaces using the configuration file inside the `keycloak` dirctory.
kubectl apply -f keycloak/keycloak.yaml -n k1
kubectl apply -f keycloak/keycloak.yaml -n k2
```



### Configure the KeyCloak using Web GUI

```yaml
# Access each KeyCloak Web interface and configure as the following, and we do not need to distinguish the value in each KeyCloak Web Configuration.
- Create a new Realm - ex: enterprise1
- Add Users (as per customer requirement) - ex: "users" and configure the confidential password - ex:"test" in the first time
- Create a new Client under realm name - ex: "sdewan"
- Under Setting for client
      > Change assess type for client to confidential
      > Under Authentication Flow Overrides - Change Direct grant flow to direct grant
      > Update Valid Redirect URIs. # "https://<istio-ingress-url>/*".
      - In Roles tab:
            > Add roles (ex. admin and user)
      - Add Mappers # Under sdewan Client under mapper tab create a mapper
            > Mapper type - User Client role
            > Client-ID: sdewan
            > Token claim name: role
            > Claim JSON Type: string
- Under Users: assign roles from sdewan client to users ( Admin and User). Verify under sdewan Client roles for user are in the role.
```



### Enable Istio Authentication and Authorization Policy

Assume you have already configure the `istio` gateway and virtual services as `README`. Now we can Install an Authentication Policy for the Keycloak server being used.

```yaml
apiVersion: security.istio.io/v1beta1
kind: RequestAuthentication
metadata:
  name: request-keycloak-auth
  namespace: istio-system
spec:
  jwtRules:
    - issuer: "http://<keycloak-k1-url>/auth/realms/enterprise1"
      jwksUri: "http://<keycloak-k1-url>/auth/realms/enterprise1/protocol/openid-connect/certs"
    - issuer: "http://<keycloak-k2-url>/auth/realms/enterprise1"
      jwksUri: "http://<akeycloak-k2-url>/auth/realms/enterprise1/protocol/openid-connect/certs"
```



### Authorization Policies with Istio

A deny policy is added to ensure that only authenticated users (with right token) are allowed access.

```yaml
apiVersion: "security.istio.io/v1beta1"
kind: "AuthorizationPolicy"
metadata:
  name: "deny-auth-policy"
  namespace: istio-system
spec:
  selector:
    matchLabels:
      istio: ingressgateway
  action: DENY
  rules:
  - from:
    - source:
        notRequestPrincipals: ["*"]
```

Curl to the scc service url will get an error "403 : RBAC: access denied"

Retrieve access token from Keycloak and use it to access resources. Note that please replace the client secret with your keycloak `sdewan` client secret. Different enterprise have its own secret and they can get the token through their keycloak service.

```shell
export TOKEN=`curl --location --request POST 'http://<keycloack url>/auth/realms/<realm-name>/protocol/openid-connect/token' --header 'Content-Type: application/x-www-form-urlencoded' --data-urlencode 'grant_type=password' --data-urlencode 'client_id=sdewan' --data-urlencode 'username=users' --data-urlencode 'password=test' --data-urlencode 'client_secret=<secret>' | jq .access_token`

curl --header "Authorization: Bearer $TOKEN"  http://<istio-ingress-url>/scc/overlays
```

#### Authorization Policies with Istio

As specified in Keycloak section Role Mappers are created using Keycloak. Check Keycloak section on how to create Roles using Keycloak. These can be used apply authorizations based on Role of the user.

For example to allow Role "Admin" to perform any operations and Role "User" to only create/delete resources under a specified project following policies can be used.

```yaml
apiVersion: "security.istio.io/v1beta1"
kind: AuthorizationPolicy
metadata:
  name: allow-admin
  namespace: istio-system
spec:
  selector:
    matchLabels:
      app: istio-ingressgateway
  action: ALLOW
  rules:
  - when:
      # The value in `request.auth.claims[]` is specified in your client mapper with tag name `Token claim name`.
    - key: request.auth.claims[role]
      # The value in `[]` is defined as the roles in your client.
      values: ["admin"]

---
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-user
  namespace: istio-system
spec:
  selector:
    matchLabels:
      app: istio-ingressgateway
  action: ALLOW
  rules:
  - from:
    - source:
      # This is the issuer you defined in pre-steps as an istio RequestAuthentication with `/*`
        requestPrincipals: ["<issuer-defined-in-RequestAuthentication>/*"]
    to:
    - operation:
        methods: ["GET"]
        paths: ["/scc/v1/overlays/overlay1"]
    when:
      # The value in `request.auth.claims[]` is specified in your client mapper with tag name `Token claim name`.
      - key: request.auth.claims[role]
        # The value in `[]` is defined as the roles in your client.
  - from:
    - source:
        requestPrincipals: ["<issuer-defined-in-RequestAuthentication>/*"]
    to:
    - operation:
        methods: ["GET"]
        paths: ["/scc/v1/overlays/overlay2"]
    when:
      # The value in `request.auth.claims[]` is specified in your client mapper with tag name `Token claim name`.
      - key: request.auth.claims[role]
        # The value in `[]` is defined as the roles in your client.
```

Then you can find that the user in k1 enterprise1 can not access overlay2 which can be accessed by the user in k2 enterprise1. Meantime, the user in k2 enterprise1 can not access overlay1 which can be accessed by the user in k1 enterprise1.
