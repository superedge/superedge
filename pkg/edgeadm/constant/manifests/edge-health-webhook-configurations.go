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
      caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNwRENDQVl3Q0NRQ2RaL0w2akZSSkdqQU5CZ2txaGtpRzl3MEJBUXNGQURBVU1SSXdFQVlEVlFRRERBbFgKYVhObE1tTWdRMEV3SGhjTk1qQXdOekU0TURRek9ERTNXaGNOTkRjeE1qQTBNRFF6T0RFM1dqQVVNUkl3RUFZRApWUVFEREFsWGFYTmxNbU1nUTBFd2dnRWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUUNSCnhHT2hrODlvVkRHZklyVDBrYVkwajdJQVJGZ2NlVVFmVldSZVhVcjh5eEVOQkF6ZnJNVVZyOWlCNmEwR2VFL3cKZzdVdW8vQWtwUEgrbzNQNjFxdWYrTkg1UDBEWHBUd1pmWU56VWtyaUVja3FOSkYzL2liV0o1WGpFZUZSZWpidgpST1V1VEZabmNWOVRaeTJISVF2UzhTRzRBTWJHVmptQXlDMStLODBKdDI3QUl4YmdndmVVTW8xWFNHYnRxOXlJCmM3Zk1QTXJMSHhaOUl5aTZla3BwMnJrNVdpeU5YbXZhSVA4SmZMaEdnTU56YlJaS1RtL0ZKdDdyV0dhQ1orNXgKV0kxRGJYQ2MyWWhmbThqU1BqZ3NNQTlaNURONDU5ellJSkVhSTFHeFI3MlhaUVFMTm8zdE5jd3IzVlQxVlpiTgo1cmhHQlVaTFlrMERtd25vWTBCekFnTUJBQUV3RFFZSktvWklodmNOQVFFTEJRQURnZ0VCQUhuUDJibnJBcWlWCjYzWkpMVzM0UWFDMnRreVFScTNVSUtWR3RVZHFobWRVQ0I1SXRoSUlleUdVRVdqVExpc3BDQzVZRHh4YVdrQjUKTUxTYTlUY0s3SkNOdkdJQUdQSDlILzRaeXRIRW10aFhiR1hJQ3FEVUVmSUVwVy9ObUgvcnBPQUxhYlRvSUVzeQpVNWZPUy9PVVZUM3ZoSldlRjdPblpIOWpnYk1SZG9zVElhaHdQdTEzZEtZMi8zcEtxRW1Cd1JkbXBvTExGbW9MCmVTUFQ4SjREZExGRkh2QWJKalFVbjhKQTZjOHUrMzZJZDIrWE1sTGRZYTdnTnhvZTExQTl6eFJQczRXdlpiMnQKUXZpbHZTbkFWb0ZUSVozSlpjRXVWQXllNFNRY1dKc3FLMlM0UER1VkNFdlg0SmRCRlA2NFhvU08zM3pXaWhtLworMXg3OXZHMUpFcz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
      service:
        namespace: kube-system
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
      caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNwRENDQVl3Q0NRQ2RaL0w2akZSSkdqQU5CZ2txaGtpRzl3MEJBUXNGQURBVU1SSXdFQVlEVlFRRERBbFgKYVhObE1tTWdRMEV3SGhjTk1qQXdOekU0TURRek9ERTNXaGNOTkRjeE1qQTBNRFF6T0RFM1dqQVVNUkl3RUFZRApWUVFEREFsWGFYTmxNbU1nUTBFd2dnRWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUUNSCnhHT2hrODlvVkRHZklyVDBrYVkwajdJQVJGZ2NlVVFmVldSZVhVcjh5eEVOQkF6ZnJNVVZyOWlCNmEwR2VFL3cKZzdVdW8vQWtwUEgrbzNQNjFxdWYrTkg1UDBEWHBUd1pmWU56VWtyaUVja3FOSkYzL2liV0o1WGpFZUZSZWpidgpST1V1VEZabmNWOVRaeTJISVF2UzhTRzRBTWJHVmptQXlDMStLODBKdDI3QUl4YmdndmVVTW8xWFNHYnRxOXlJCmM3Zk1QTXJMSHhaOUl5aTZla3BwMnJrNVdpeU5YbXZhSVA4SmZMaEdnTU56YlJaS1RtL0ZKdDdyV0dhQ1orNXgKV0kxRGJYQ2MyWWhmbThqU1BqZ3NNQTlaNURONDU5ellJSkVhSTFHeFI3MlhaUVFMTm8zdE5jd3IzVlQxVlpiTgo1cmhHQlVaTFlrMERtd25vWTBCekFnTUJBQUV3RFFZSktvWklodmNOQVFFTEJRQURnZ0VCQUhuUDJibnJBcWlWCjYzWkpMVzM0UWFDMnRreVFScTNVSUtWR3RVZHFobWRVQ0I1SXRoSUlleUdVRVdqVExpc3BDQzVZRHh4YVdrQjUKTUxTYTlUY0s3SkNOdkdJQUdQSDlILzRaeXRIRW10aFhiR1hJQ3FEVUVmSUVwVy9ObUgvcnBPQUxhYlRvSUVzeQpVNWZPUy9PVVZUM3ZoSldlRjdPblpIOWpnYk1SZG9zVElhaHdQdTEzZEtZMi8zcEtxRW1Cd1JkbXBvTExGbW9MCmVTUFQ4SjREZExGRkh2QWJKalFVbjhKQTZjOHUrMzZJZDIrWE1sTGRZYTdnTnhvZTExQTl6eFJQczRXdlpiMnQKUXZpbHZTbkFWb0ZUSVozSlpjRXVWQXllNFNRY1dKc3FLMlM0UER1VkNFdlg0SmRCRlA2NFhvU08zM3pXaWhtLworMXg3OXZHMUpFcz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
      service:
        namespace: kube-system
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
