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

package controller

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/pelletier/go-toml"
	"github.com/superedge/superedge/pkg/penetrator/apis/nodetask.apps.superedge.io/v1beta1"
	clientset "github.com/superedge/superedge/pkg/penetrator/client/clientset/versioned"
	"github.com/superedge/superedge/pkg/penetrator/client/clientset/versioned/scheme"
	informers "github.com/superedge/superedge/pkg/penetrator/client/informers/externalversions/nodetask.apps.superedge.io/v1beta1"
	listers "github.com/superedge/superedge/pkg/penetrator/client/listers/nodetask.apps.superedge.io/v1beta1"
	"github.com/superedge/superedge/pkg/penetrator/constants"
	"github.com/superedge/superedge/pkg/penetrator/job/conf"
	"github.com/superedge/superedge/pkg/penetrator/operator/context"
	"github.com/superedge/superedge/pkg/util"
	kubecli "github.com/superedge/superedge/pkg/util/kubeclient"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/workqueue"
	tokenapi "k8s.io/cluster-bootstrap/token/api"
	tokenutil "k8s.io/cluster-bootstrap/token/util"
	"k8s.io/klog/v2"
	"net/url"
	"reflect"
	"strings"
	"time"
)

var (
	KeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
)

type NodeTaskController struct {
	kubeClient        kubernetes.Interface
	nodeTaskClientset clientset.Interface
	nodeTaskLister    listers.NodeTaskLister
	nodeTaskSynced    cache.InformerSynced
	workqueue         workqueue.RateLimitingInterface
	metaRecorder      record.EventRecorder
	ctx               *context.NodeTaskContext
}

func NewNodeTaskController(kube kubernetes.Interface, ntclient clientset.Interface, ntInformer informers.NodeTaskInformer, nodetaskctx *context.NodeTaskContext, recorder record.EventRecorder) *NodeTaskController {
	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))

	controller := &NodeTaskController{
		kubeClient:        kube,
		nodeTaskClientset: ntclient,
		nodeTaskLister:    ntInformer.Lister(),
		nodeTaskSynced:    ntInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "NodeTask"),
		metaRecorder:      recorder,
		ctx:               nodetaskctx,
	}
	ntInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueue,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldNodeTask, ok1 := oldObj.(*v1beta1.NodeTask)
			newNodeTask, ok2 := newObj.(*v1beta1.NodeTask)
			if ok1 && ok2 {
				// Check whether it is a delete event
				if !reflect.DeepEqual(oldNodeTask.ObjectMeta.DeletionTimestamp, newNodeTask.ObjectMeta.DeletionTimestamp) {
					controller.enqueue(newNodeTask)
					return
				}

				// Check if the nodes are added
				if newNodeTask.Status.NodeTaskStatus == v1beta1.NodeTaskStatusCreating || reflect.DeepEqual(newNodeTask.Status, v1beta1.NodeTaskStatus{}) {
					controller.enqueue(newNodeTask)
				}
			}
		},
	}, 60*time.Second)
	return controller
}

func (ntController *NodeTaskController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	klog.Info("start nodetask controller")

	klog.Info("waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, ntController.nodeTaskSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(ntController.runWork, time.Second, stopCh)
	}

	return nil
}

func (ntController *NodeTaskController) enqueue(obj interface{}) {
	key, err := KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}
	ntController.workqueue.Add(key)
}

func (ntController *NodeTaskController) runWork() {
	for ntController.processNextWorkItem() {

	}
}

func (ntController *NodeTaskController) processNextWorkItem() bool {
	obj, shutdown := ntController.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(item interface{}) error {
		defer ntController.workqueue.Done(item)
		var key string
		var ok bool

		if key, ok = item.(string); !ok {
			ntController.workqueue.Forget(key)
			utilruntime.HandleError(fmt.Errorf("excepted string in workqueue but got %#v", item))
			return nil
		}

		if err := ntController.syncHandler(key); err != nil {
			ntController.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s':%s, requeuing", key, err.Error())
		}

		ntController.workqueue.Forget(item)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
	}

	return true
}

func (ntController *NodeTaskController) syncHandler(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return err
	}
	nt, err := ntController.nodeTaskLister.Get(name)

	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Infof("nodetask has been deleted, namespace: %s name: %s", ns, name)
		}
		return err
	}

	// Check whether the job for installing the node exists
	nodeJob, err := ntController.kubeClient.BatchV1().Jobs(constants.NameSpaceEdge).Get(ntController.ctx, nt.Annotations[constants.AnnotationAddNodeJobName], metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			ntCopy := nt.DeepCopy()
			err = ntController.prepareJob(ntCopy)
			if err != nil {
				klog.Errorf("error in preparation before job creation, error: %v", err)
				return err
			}
			if ntCopy.Status.NodeTaskStatus == v1beta1.NodeTaskStatusCreating {
				node, err := ntController.kubeClient.CoreV1().Nodes().Get(ntController.ctx, ntCopy.Spec.ProxyNode, metav1.GetOptions{})
				if err != nil {
					klog.Errorf("Failed to get node: %s", ntCopy.Spec.ProxyNode)
					return err
				}
				if _, hasMasterRoleLabel := node.Labels[util.KubernetesControlPlaneLabel]; hasMasterRoleLabel {
					err = createNodeJob(ntController.kubeClient, nt, true)
				} else {
					err = createNodeJob(ntController.kubeClient, nt, false)
				}
				if err != nil {
					klog.Errorf("failed to create NodeJob, error: %v", err)
					return err
				}
			}

			if !reflect.DeepEqual(nt.Status, ntCopy.Status) {
				_, err = ntController.nodeTaskClientset.NodestaskV1beta1().NodeTasks().UpdateStatus(ntController.ctx, ntCopy, metav1.UpdateOptions{})
				if err != nil {
					klog.Errorf("failed to update NodeTaskStatus, error: %v", err)
					return err
				}
			}
		}
	} else {
		if kubecli.IsJobFinished(nodeJob) {
			err = ntController.kubeClient.BatchV1().Jobs(constants.NameSpaceEdge).Delete(ntController.ctx, nt.Annotations[constants.AnnotationAddNodeJobName], metav1.DeleteOptions{})
			if err != nil {
				klog.Errorf("failed to delete NodeJob, error: %v", err)
				return err
			}
		}
	}

	return nil
}

func (ntController *NodeTaskController) prepareJob(nt *v1beta1.NodeTask) error {
	err := filterNodeIps(nt, ntController.kubeClient, ntController.ctx)
	if err != nil {
		klog.Errorf("Failed to filter nodeIps, error: %v", err)
		return err
	}

	if nt.Status.NodeTaskStatus == v1beta1.NodeTaskStatusReady {
		return nil
	}

	bootStrapToken, err := getBootStrapToken(ntController.ctx, ntController.kubeClient, constants.Expiration)
	if err != nil {
		klog.Errorf("Failed to get bootstraptoken, error: %v", err)
		return err
	}

	jobConf, err := ntController.kubeClient.CoreV1().ConfigMaps(constants.NameSpaceEdge).Get(ntController.ctx, nt.Annotations[constants.AnnotationAddNodeConfigmapName], metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			jobConf = &corev1.ConfigMap{}
			jobConf.Name = nt.Annotations[constants.AnnotationAddNodeJobName]
			jobConf.Namespace = nt.Namespace
			block := true
			jobConf.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion:         "batch/v1",
					Kind:               "Job",
					Name:               nt.Annotations[constants.AnnotationAddNodeJobName],
					Controller:         &block,
					BlockOwnerDeletion: &block,
					UID:                nt.UID,
				},
			}
			jobConf.Data = make(map[string]string)

			//Fill in the configuration file of the job
			jobConfig := conf.JobConfig{}
			jobConfig.NodeLabel = nt.Labels[constants.NodeLabel]
			jobConfig.NodesIps = nt.Status.NodeStatus
			jobConfig.SSHPort = nt.Spec.SSHPort
			jobConfig.AdmToken = bootStrapToken
			caHash, apiserverAddr, apiserverPort, err := getCaHashAndApiserverAddr(ntController.ctx, ntController.kubeClient)
			if err != nil {
				klog.Errorf("Failed to get caHash, error: %v", err)
				return err
			}
			jobConfig.ApiserverAddr = apiserverAddr
			jobConfig.ApiserverPort = apiserverPort
			jobConfig.CaHash = caHash

			jobConfigBD, err := toml.Marshal(jobConfig)
			if err != nil {
				klog.Errorf("failed to marshal job config, error: %v", err)
				return err
			}
			jobConf.Data[constants.JobConf] = string(jobConfigBD)

			_, err = ntController.kubeClient.CoreV1().ConfigMaps(constants.NameSpaceEdge).Create(ntController.ctx, jobConf, metav1.CreateOptions{})
			if err != nil {
				klog.Errorf("failed to create job config configmap, error: %v", err)
				return err
			}
		} else {
			klog.Errorf("Failed to get  configmap %s, error: %s", nt.Annotations[constants.AnnotationAddNodeConfigmapName], err)
			return err
		}
	} else {
		cmconf := &conf.JobConfig{}
		err = toml.Unmarshal([]byte(jobConf.Data[constants.JobConf]), cmconf)
		if err != nil {
			klog.Errorf("failed to get jobconf from configmap, error: %v", err)
			return err
		}

		cmconf.AdmToken = bootStrapToken
		cmconf.NodesIps = nt.Status.NodeStatus

		cmjb, err := toml.Marshal(*cmconf)
		if err != nil {
			klog.Errorf("failed to marshal job config, error: %v", err)
			return err
		}
		jobConf.Data[constants.JobConf] = string(cmjb)

		_, err = ntController.kubeClient.CoreV1().ConfigMaps(constants.NameSpaceEdge).Update(ntController.ctx, jobConf, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("failed to update config configmap %s, error: %v", nt.Annotations[constants.AnnotationAddNodeConfigmapName], err)
			return err
		}
	}

	return nil
}

func filterNodeIps(nt *v1beta1.NodeTask, kubeclient kubernetes.Interface, ctx *context.NodeTaskContext) error {
	nl, err := kubeclient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: constants.NodeLabel + "=" + nt.Labels[constants.NodeLabel]})
	if err != nil {
		klog.Errorf("failed to get existing nodes of the user cluster, error: %v", err)
		return err
	}

	//The status of the node to be installed is synchronized from all nodes to be installed
	ipMap := make(map[string]string)

	// Filter duplicate ip of node name
	for k, v := range nt.Spec.NodeNamesOverride {
		ipMap[v] = k
	}

	if len(ipMap) == len(nt.Spec.NodeNamesOverride) {
		nt.Status.NodeStatus = nt.DeepCopy().Spec.NodeNamesOverride
	} else {
		nameMap := make(map[string]string)
		for k, v := range ipMap {
			nameMap[v] = k
		}
		nt.Status.NodeStatus = nameMap
	}

	for _, node := range nl.Items {
		delete(nt.Status.NodeStatus, node.Name)
	}

	if len(nt.Status.NodeStatus) == 0 {
		nt.Status.NodeTaskStatus = v1beta1.NodeTaskStatusReady
	} else {
		nt.Status.NodeTaskStatus = v1beta1.NodeTaskStatusCreating
	}
	return nil
}

func createNodeJob(kubeclient kubernetes.Interface, nt *v1beta1.NodeTask, hasMasterRoleLabel bool) error {
	options := map[string]interface{}{
		"JobName":      nt.Annotations[constants.AnnotationAddNodeJobName],
		"NameSpace":    constants.NameSpaceEdge,
		"SecretName":   nt.Spec.SSHCredential,
		"JobConfig":    nt.Annotations[constants.AnnotationAddNodeConfigmapName],
		"NodeName":     nt.Spec.ProxyNode,
		"NodeTaskName": nt.Annotations[constants.AnnotationAddNodeJobName],
		"Uid":          nt.UID,
	}
	var secretTemp []byte
	var err error
	if hasMasterRoleLabel {
		secretTemp, err = util.ReadFile(constants.DirectAddNodeJob)
		if err != nil {
			klog.Errorf("Failed to read file:%s, error: %v", constants.DirectAddNodeJob, err)
			return err
		}
	} else {
		secretTemp, err = util.ReadFile(constants.SpringboardAddNodeJob)
		if err != nil {
			klog.Errorf("Failed to read file:%s, error: %v", constants.SpringboardAddNodeJob, err)
			return err
		}
	}
	err = kubecli.CreateResourceWithFile(kubeclient, string(secretTemp), options)
	if err != nil {
		klog.Errorf("Failed to create a job that adds nodes through nodes in the cluster, error: %v", err)
		return err
	}
	return nil
}

func getBootStrapToken(ctx *context.NodeTaskContext, kubeclient kubernetes.Interface, expiration time.Duration) (string, error) {
	var token string
	secrets, err := kubeclient.CoreV1().Secrets(metav1.NamespaceSystem).List(ctx, metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(map[string]string{"type": string(tokenapi.SecretTypeBootstrapToken)}).String(),
	})
	if err != nil {
		klog.Errorf("Failed to list bootstraptoken, error: %v", err)
		return token, err
	}
	if len(secrets.Items) == 0 {
		token, err = tokenutil.GenerateBootstrapToken()
		if err != nil {
			klog.Errorf("Failed to generate bootstraptoken, error: %v", err)
			return token, err
		}
		tokens := strings.Split(token, ".")
		base64ToenId := base64.StdEncoding.EncodeToString([]byte(tokens[0]))
		base64TokenSecret := base64.StdEncoding.EncodeToString([]byte(tokens[1]))
		base64Expiration := base64.StdEncoding.EncodeToString([]byte(time.Now().Add(expiration).Format(time.RFC3339)))

		options := map[string]string{
			"Base64TokenId":     base64ToenId,
			"Base64TokenSecret": base64TokenSecret,
			"TokenId":           tokens[0],
			"Base64Expiration":  base64Expiration,
		}
		secretTmep, err := util.ReadFile(constants.BootStrapTokenSecert)
		if err != nil {
			klog.Errorf("Failed to read file:%s, error: %v", constants.BootStrapTokenSecert, err)
			return token, err
		}
		err = kubecli.CreateResourceWithFile(kubeclient, string(secretTmep), options)
		if err != nil {
			klog.Errorf("Failed to create bootstraptoken, error: %v", err)
			return token, err
		}
	} else {
		var secret corev1.Secret
		var t time.Time
		for k, item := range secrets.Items {
			if k == 0 {
				secret = item
				t, err = time.Parse(time.RFC3339, string(item.Data[tokenapi.BootstrapTokenExpirationKey]))
				if err != nil {
					klog.Errorf("Failed to parse the expiration time of bootstraptoken, error: %v", err)
					return token, err
				}
			} else if k > 0 {
				tmp, err := time.Parse(time.RFC3339, string(item.Data[tokenapi.BootstrapTokenExpirationKey]))
				if err != nil {
					klog.Errorf("Failed to parse the expiration time of bootstraptoken, error: %v", err)
					return token, err
				}
				if t.Before(tmp) {
					secret = item
					t = tmp
				}

			}
		}
		token = tokenutil.TokenFromIDAndSecret(string(secret.Data[tokenapi.BootstrapTokenIDKey]), string(secret.Data[tokenapi.BootstrapTokenSecretKey]))
	}
	return token, nil
}

func getCaHashAndApiserverAddr(ctx *context.NodeTaskContext, kubeclient kubernetes.Interface) (string, string, string, error) {
	var caHash string
	var apiserverAddr string
	var apiserverPort string
	clusterInfo, err := kubeclient.CoreV1().ConfigMaps(util.NamespaceKubePublic).Get(ctx, tokenapi.ConfigMapClusterInfo, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get configmap %s, error: %v", tokenapi.ConfigMapClusterInfo, err)
		return caHash, apiserverAddr, apiserverPort, err
	}
	config, err := clientcmd.Load([]byte(clusterInfo.Data[tokenapi.KubeConfigKey]))
	if err != nil {
		klog.Errorf("Failed to get kubeconfig, error: %v", err)
		return caHash, apiserverAddr, apiserverPort, err
	}

	cacert, err := certutil.ParseCertsPEM([]byte(config.Clusters[""].CertificateAuthorityData))
	if err != nil {
		klog.Errorf("Failed to parse cacert, error: %v", err)
		return caHash, apiserverAddr, apiserverPort, err
	}
	apiserverUrl, err := url.Parse(config.Clusters[""].Server)
	if err != nil {
		return caHash, apiserverAddr, apiserverPort, err
	}
	apiserverAddr = apiserverUrl.Host
	apiserverPort = apiserverUrl.Port()
	caHash = Hash(cacert[0])
	return caHash, apiserverAddr, apiserverPort, nil
}

func Hash(certificate *x509.Certificate) string {
	spkiHash := sha256.Sum256(certificate.RawSubjectPublicKeyInfo)
	return "sha256" + ":" + strings.ToLower(hex.EncodeToString(spkiHash[:]))
}
