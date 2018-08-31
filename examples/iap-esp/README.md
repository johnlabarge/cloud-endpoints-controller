# Cloud Endpoints and IAP Example with ESP

[![button](http://gstatic.com/cloudssh/images/open-btn.png)](https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/GoogleCloudPlatform/cloud-endpoints-operator&page=editor&tutorial=examples/iap-esp/README.md)

This example shows how to use the Cloud Endpoints Controller with IAP and an L7 Ingress Load Balancer.

**Figure 1.** *diagram of Google Cloud resources*

![architecture diagram](https://raw.githubusercontent.com/GoogleCloudPlatform/cloud-endpoints-operator/master/examples/iap-esp/diagram.png)

## Task 0 - Setup environment

1. Set the project, replace `YOUR_PROJECT` with your project ID:

```
gcloud config set project YOUR_PROJECT
```

2. Enable the Service Management API:

```
gcloud services enable servicemanagement.googleapis.com
```

3. Create GKE cluster:

```
VERSION=$(gcloud container get-server-config --zone us-central1-c --format='value(validMasterVersions[0])')
gcloud container clusters create dev --zone=us-central1-c --cluster-version=${VERSION} --scopes=cloud-platform
```

4. Change to the example directory:

```
[[ `basename $PWD` != iap-esp ]] && cd examples/iap-esp
```

## Task 1 - Install Cloud Endpoints Controller

1. Install helm

```sh
curl -L https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
kubectl create clusterrolebinding default-admin --clusterrole=cluster-admin --user=$(gcloud config get-value account);
kubectl create serviceaccount tiller --namespace kube-system;
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller;
helm init --service-account=tiller;
helm repo update
helm version
```

2. Install Cloud Endpoints Controller

```sh
kubectl create clusterrolebinding ${USER}-cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)

kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/metacontroller/master/manifests/metacontroller-rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/metacontroller/master/manifests/metacontroller.yaml

kubectl apply -f https://raw.githubusercontent.com/danisla/cloud-endpoints-controller/master/manifests/cloud-endpoints-controller-rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/danisla/cloud-endpoints-controller/master/manifests/cloud-endpoints-controller.yaml
```

3. Install cert-manager and ACME ClusterIssuers 

```sh
  helm install --name cert-manager --namespace kube-system stable/cert-manager
  EMAIL=$(gcloud config get-value account)

  cat <<EOF | kubectl apply -f -
apiVersion: certmanager.k8s.io/v1alpha1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
  namespace: kube-system
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: ${EMAIL}
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging
    # Enable the HTTP-01 challenge provider
    http01: {}
---
apiVersion: certmanager.k8s.io/v1alpha1
kind: ClusterIssuer
metadata:
  name: letsencrypt-production
  namespace: kube-system
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: ${EMAIL}
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-production
    # Enable the HTTP-01 challenge provider
    http01: {}
EOF
```

## Task 2 - Deploy Sample App

1. Deploy the target app that you will proxy with IAP:

```
kubectl run sample-app --image gcr.io/cloud-solutions-group/esp-sample-app:1.0.0 --port 8080
kubectl expose deploy sample-app --type ClusterIP
```

## Task 3 - Configure OAuth consent screen

1. Go to the [OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent).
2. Under __Email address__, select the email address you want to display as a public contact. This must be your email address, or a Google Group you own.
3. Enter the __Application name__ you would like to display.
4. Under __Authorized Domains__, add `iap-tutorial.endpoints.PROJECT_ID.cloud.goog`, replacing `PROJECT_ID` with the ID of your project.
5. Add any optional details you’d like.
6. Click __Save__.

## Task 4 - Set up IAP access

1. Go to the [Identity-Aware Proxy page](https://console.cloud.google.com/security/iap/project).
2. On the right side panel, next to __Access__, click __Add__.
3. In the __Add members__ dialog that appears, add the email addresses of groups or individuals to whom you want to grant the __IAP-Secured Web App User__ role for the project

    The following kinds of accounts can be members:
    - __Google Accounts__: user@gmail.com
    - __Google Groups__: admins@googlegroups.com
    - __Service accounts__: server@example.gserviceaccount.com
    - __G Suite domains__: example.com

    Make sure to add a Google account that you have access to.

4. When you're finished adding members, click __Add__.

## Task 5 - Create OAuth credentials

1. Go to the [Credentials page](https://console.cloud.google.com/apis/credentials)
2. Click __Create Credentials > OAuth client ID__,
3. Under __Application type__, select Web application. In the __Name__ box, enter `IAP Tutorial`, and in the __Authorized redirect URIs__ box, enter `https://iap-tutorial.endpoints.PROJECT_ID.cloud.goog/_gcp_gatekeeper/authenticate`, replacing `PROJECT_ID` with the ID of your project. These are domains you will be able to access your IAP-enabled BackendService’s from.
4. When you are finished, click __Create__. After your credentials are created, make note of the client ID and client secret that appear in the OAuth client window.
5. In Cloud Shell, create a Kubernetes secret with your OAuth credentials:

```
CLIENT_ID=YOUR_CLIENT_ID
CLIENT_SECRET=YOUR_CLIENT_SECRET
```

```
kubectl create secret generic iap-oauth --from-literal=client_id=${CLIENT_ID} --from-literal=client_secret=${CLIENT_SECRET}
```

## Task 6 - Deploy iap-ingress chart

1. Create values file for chart:

```
cat > values.yaml <<EOF
projectID: $(gcloud config get-value project)
endpointServiceName: iap-tutorial
targetServiceName: sample-app
targetServicePort: 8080
oauthSecretName: iap-oauth
esp:
  enabled: true
EOF
```

2. Deploy chart to create IAP aware ingress resource:

```
helm install --name iap-tutorial-ingress ../charts/iap-ingress -f values.yaml
```

3. Wait for the load balancer to be provisioned:

```
PROJECT=$(gcloud config get-value project)
COMMON_NAME="iap-tutorial.endpoints.${PROJECT}.cloud.goog"

(until [[ $(curl -sf -w "%{http_code}" https://${COMMON_NAME}) == "302" ]]; do echo "Waiting for LB with IAP..."; sleep 2; done)
```

4. Open your browser to `https://iap-tutorial.endpoints.PROJECT_ID.cloud.goog` replacing `PROJECT_ID` with your project id.
5. Login with your Google account and verify the the sample app is show.

> NOTE: It may take 10-15 minutes for the load balancer to be provisioned.

## Task 7 - Cleanup

1. Delete the chart:

```
helm delete --purge iap-tutorial-ingress
```

> This will trigger the load balancer cleanup. Wait a few moments before continuing.

2. Delete the GKE cluster:

```
gcloud container clusters delete dev --zone us-central1-c
```
