/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manifests

const APP_EDGE_HEALTH_WEBHOOK = "edge-health-webhook.yaml"

const EdgeHealthWebhookConfigYaml = `
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: edge-health-admission
webhooks:
  - admissionReviewVersions:
      - v1beta1
    clientConfig:
      caBundle: {{.CABundle}}
      service:
        namespace: {{.Namespace}}
        name: edge-health-admission
        path: /node-taint
    failurePolicy: Ignore
    matchPolicy: Exact
    name: node-taint.k8s.io
    namespaceSelector: {}
    objectSelector: {}
    reinvocationPolicy: Never
    rules:
      - apiGroups:
          - '*'
        apiVersions:
          - '*'
        operations:
          - UPDATE
        resources:
          - nodes
        scope: '*'
    sideEffects: None
    timeoutSeconds: 5
  - admissionReviewVersions:
      - v1beta1
    clientConfig:
      caBundle: {{.CABundle}}
      service:
        namespace: {{.Namespace}}
        name: edge-health-admission
        path: /endpoint
    failurePolicy: Ignore
    matchPolicy: Exact
    name: endpoint.k8s.io
    namespaceSelector: {}
    objectSelector: {}
    reinvocationPolicy: Never
    rules:
      - apiGroups:
          - '*'
        apiVersions:
          - '*'
        operations:
          - UPDATE
        resources:
          - endpoints
        scope: '*'
    sideEffects: None
    timeoutSeconds: 5
`
