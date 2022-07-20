# istio-sdewan
This is for the integration of istio and sdewan

<!-- omit in toc -->
SDEWAN uses Istio and other open source solutions to leverage Istio Authorization and Authentication frameworks. Authentication for the SDEWAN users are done at the Istio Gateway, where all the traffic enters the cluster. Istio along with Authservice (Istio ecosystem project) enables request-level authentication with JSON Web Token (JWT) validation. This can be achieved using a custom authentication provider or any OpenID Connect providers like Keycloak, Auth0 etc.

Authservice (https://github.com/istio-ecosystem/authservice) is an istio-ecosystem project that works alongside with Envoy proxy. It is used along with Istio to work with external IAM systems (OAUTH2). Many Enterprises have their own OAUTH2 server for authenticating users and providing roles to users.

## Steps for setting up SDEWAN with Istio

These steps need to be followed in the Kubernetes Cluster where SDEWAN-CNF, SDEWAN-CRD-CONTROLLER and SDEWAN-OVERLAY-CONTROLLER are installed.

### Pre-Installation

Note: The `url` in the following configurations are recommended to be set as node port mode. For example, `<keycloak-url>` can be found from the command `kubectl get service -A`.

#### Istio

- Install istioctl and init

  ```shell
  # Download the latest istio release
  curl -L https://istio.io/downloadIstio | sh -

  # Download the specified istio release for target_arch
  # In the guide, we need istio version >= 1.7.4
  # curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.7.4 TARGET_ARCH=x86_64 sh -

  # Add the `istioctl` client binary to your path
  export PATH=$PWD/bin:$PATH

  # Deploy istio operator
  istioctl operator init
  ```

- Install the `istio` demo profile using operator

  ```shell
  # Create istio namespace
  kubectl create ns istio-system

  # Aplly the istio demo profile
  kubectl apply -f - <<EOF
  apiVersion: install.istio.io/v1alpha1
  kind: IstioOperator
  metadata:
    namespace: istio-system
    name: example-istiocontrolplane
  spec:
    profile: demo
  EOF
  ```

#### Keycloak

- Create the certificate file for keycloak

  ```shell
  # Use the key and cert in istio-sdewan/keycloak or using the commands to create
  openssl genrsa 2048 > keycloak.key
  openssl req -new -x509 -nodes -sha256 -days 365 -key keycloak.key -out keycloak.crt
  ```

- Configure and setup keycloak in k8s cluster

  ```shell
  # Create keycloak namespace
  kubectl create ns keycloak

  # Create secret for keycloak (key and cert here can be created or offered in repo)
  kubectl create -n keycloak secret tls ca-keycloak-certs --key keycloak.key --cert keycloak.crt

  # Deploy keycloak
  kubectl apply -f keycloak/keycloak.yaml -n keycloak
  ```

- Configure in Keycloak using its web interface

  ```
  - Create a new Realm - ex: enterprise1
  - Add Users (as per customer requirement)
  - Create a new Client under realm name - ex: sdewan
  - Under Setting for client
        > Change assess type for client to confidential
        > Under Authentication Flow Overrides - Change Direct grant flow to direct grant
        > Update Valid Redirect URIs. # "https://istio-ingress-url/*".
        - In Roles tab:
              > Add roles (ex. admin and user)
        - Add Mappers # Under sdewan Client under mapper tab create a mapper
              > Mapper type - User Client role
              > Client-ID: sdewan
              > Token claim name: role
              > Claim JSON Type: string
  - Under Users: assign roles from sdewan client to users ( Admin and User). Verify under sdewan Client roles for user are in the role.
  ```

### Configure and integrate

#### Istio Sidecar Injection for SDEWAN namespace

```Shell
# Label the namespace needed sidecar injection
kubectl label namespace sdewan-system istio-injection=enabled

# To exclude deployment - no need for sidecar
spec:
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
```
And then you can Install SDEWAN overlay controller in the `sdewan-system` namespace

#### Enable mTLS in target namespace

```Shell
# We set the namespace as `sdewan-system` and enable mTLS
kubectl apply -f istio-example-yaml/mTLS.yaml
```

The following is an example to enable mTLS in specified namespace, `sdewan-system` here.

```yaml
apiVersion: "security.istio.io/v1beta1"
kind: "PeerAuthentication"
metadata:
  name: "default"
  namespace: sdewan-system
spec:
  mtls:
    mode: STRICT
```

#### Configure Istio Ingress Gateway

Create certificate for Ingress Gateway and create secret for Istio Ingress Gateway
```shell
# Use the command to create key and cert for istio
openssl genrsa 2048 > v2.key
openssl req -new -x509 -nodes -sha256 -days 365 -key v2.key -out v2.crt

kubectl create -n istio-system secret tls sdewan-credential --key=v2.key --cert=v2.crt
```

Example Gateway yaml

```shell
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: sdewan-gateway
  namespace: sdewan-system
spec:
  selector:
    istio: ingressgateway # use Istio default gateway implementation
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: sdewan-credential
    hosts:
    - "*"
```

#### Create Istio VirtualServices Resources for SDEWAN
An Istio VirtualService Resource is required to be created for each of the microservices.

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: vs-sdewan-scc
  namespace: sdewan-system
spec:
  hosts:
  - "*"
  gateways:
  - sdewan-gateway
  http:
  - match:
    - uri:
        prefix: /scc/v1
    - uri:
        exact: /scc/v1/overlays
    - uri:
        regex: /scc/v1/overlays/[^\/]*
    - uri:
        exact: /scc/v1/provider
    - uri:
        regex: /scc/v1/provider/[^\/]*
    - uri:
        regex: /scc/v1/provider/ipranges
    - uri:
        regex: /scc/v1/provider/ipranges/[^\/]*
    - uri:
        regex: /scc/v1/overlays/.+/ipranges
    - uri:
        regex: /scc/v1/overlays/.+/ipranges/[^\/]*
    - uri:
        regex: /scc/v1/overlays/.+/certificates
    - uri:
        regex: /scc/v1/overlays/.+/certificates/[^\/]*
    - uri:
        regex: /scc/v1/overlays/.+/proposals
    - uri:
        regex: /scc/v1/overlays/.+/proposals/[^\/]*
    - uri:
        regex: /scc/v1/overlays/.+/hubs
    - uri:
        regex: /scc/v1/overlays/.+/hubs/[^\/]*
    - uri:
        regex: /scc/v1/overlays/.+/hubs/.+/cnfs
    - uri:
        regex: /scc/v1/overlays/.+/hubs/.+/devices
    - uri:
        regex: /scc/v1/overlays/.+/hubs/.+/devices/[^\/]*
    - uri:
        regex: /scc/v1/overlays/.+/hubs/.+/connections
    - uri:
        regex: /scc/v1/overlays/.+/hubs/.+/connections/[^\/]*
    - uri:
        regex: /scc/v1/overlays/.+/devices
    - uri:
        regex: /scc/v1/overlays/.+/devices/[^\/]*
    - uri:
        regex: /scc/v1/overlays/.+/devices/.+/cnfs
    - uri:
        regex: /scc/v1/overlays/.+/devices/.+/connections
    - uri:
        regex: /scc/v1/overlays/.+/devices/.+/connections/[^\/]*

    route:
    - destination:
        port:
          number: 9015
        host: scc
```

Make sure the SDEWAN overlay controller is accessible through Istio Ingress Gateway at this point.  "https://istio-ingress-url/scc/v1"


```shell
curl http://<istio-ingress-url>/scc/v1
200: ...
```

#### Enable Istio Authentication and Authorization Policy
Install an Authentication Policy for the Keycloak server being used.

```yaml
apiVersion: security.istio.io/v1beta1
kind: RequestAuthentication
metadata:
  name: request-keycloak-auth
  namespace: istio-system
spec:
  jwtRules:
  - issuer: "http://<keycloak-url>/auth/realms/enterprise1"
    jwksUri: "http://<keycloak-url>/auth/realms/enterprise1/protocol/openid-connect/certs"
# And you can create another keycloak service if you need
# jwtRules:
# - issuer: "http://<ano-keycloak-url>/auth/realms/ano-enterprise"
#   jwksUri: "http://<ano-keycloak-url>/auth/realms/ano-enterprise/protocol/openid-connect/certs"
```

#### Authorization Policies with Istio

A deny policy is added to ensure that only authenticated users (with right token) are allowed to access specified resources.

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

Curl to the scc url will give an error "403 : RBAC: access denied"

Retrieve access token from Keycloak and use it to access resources. Note that please replace the client secret with your keycloak `sdewan` client secret.

```shell
export TOKEN=`curl --location --request POST 'http://<keycloack url>/auth/realms/<realm-name>/protocol/openid-connect/token' --header 'Content-Type: application/x-www-form-urlencoded' --data-urlencode 'grant_type=password' --data-urlencode 'client_id=sdewan' --data-urlencode 'username=user1' --data-urlencode 'password=test' --data-urlencode 'client_secret=<secret>' | jq .access_token`

curl --header "Authorization: Bearer $TOKEN"  http://<istio-ingress-url>/scc/overlays
```
#### Authorization Policies with Istio

As specified in Keycloak section Role Mappers are created using Keycloak. Check Keycloak section on how to create Roles using Keycloak. These can be used to apply authorizations based on Role of the user.

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
  - to:
    - operation:
        paths: ["/scc/v1/*"]
    - operation:
        methods: ["GET"]
        paths: ["/scc/v1/overlays"]
    when:
        # The value in `request.auth.claims[]` is specified in your client mapper with tag name `Token claim name`.
      - key: request.auth.claims[role]
        # The value in `[]` is defined as the roles in your client.
        values: ["user"]

# If you have another keycloak service, you may create specified configuration for target mapper claim name.
```

Then you can only access specified resources with the roles your account have.

Note: For multi-keycloak services for different companies or tenants, you can refer to this [page](https://github.com/akraino-edge-stack/icn-sdwan/tree/master/central-controller/docs/istio/Multiple%20KeyCloak%20Configuration.md)
