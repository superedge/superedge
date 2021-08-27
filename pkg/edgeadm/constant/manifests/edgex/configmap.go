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

package edgex

//configmap for each component, important, it should be built firstly and deleted lastly
const EDGEX_CONFIGMAP = "k8s-hanoi-redis-no-secty-configmap.yml"

const Edgex_CONFIGMAP_Yaml = `
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: {{.Namespace}}
  name: common-variables
data:
  EDGEX_SECURITY_SECRET_STORE: "false"
  Registry_Host: "edgex-core-consul"
  Clients_CoreData_Host: "edgex-core-data"
  Clients_Data_Host: "edgex-core-data"
  Clients_Notifications_Host: "edgex-support-notifications"
  Clients_Metadata_Host: "edgex-core-metadata"
  Clients_Command_Host: "edgex-core-command"
  Clients_Scheduler_Host: "edgex-support-scheduler"
  Clients_RulesEngine_Host: "edgex-kuiper"
  Clients_VirtualDevice_Host: "edgex-device-virtual"
  Databases_Primary_Host: "edgex-redis"
  Service_ServerBindAddr: "0.0.0.0"
  Logging_EnableRemote: "false"


`
