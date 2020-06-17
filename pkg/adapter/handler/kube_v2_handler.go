package handler

import (
	"context"
	"fmt"
	"github.com/mesh-operator/pkg/adapter/events"
	"github.com/mesh-operator/pkg/adapter/utils"
	v1 "github.com/mesh-operator/pkg/apis/mesh/v1"
	k8smanager "github.com/mesh-operator/pkg/k8s/manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KubeV2EventHandler it used for synchronizing the events which has been send by the adapter client
// to a kubernetes cluster which has an istio controller there.
// It usually uses a CRD group to depict both registered services and instances.
type KubeV2EventHandler struct {
	K8sMgr *k8smanager.ClusterManager
}

// replace
func (kubev2eh *KubeV2EventHandler) replace(event *events.ServiceEvent, amc *v1.AppMeshConfig) {
	var svc *events.Service
	if event.Service == nil {
		svc = event.Instance.Service
	} else {
		svc = event.Service
	}
	s := convertService(svc)
	if amc.Spec.Services == nil {
		var services []*v1.Service
		services = append(services, s)
		amc.Spec.Services = services
	} else {
		for index, as := range amc.Spec.Services {
			if as.Name == s.Name {
				// No matter what happen we replacing the existed service
				amc.Spec.Services[index] = s
				// break
			}
		}
	}
}

// convertService
func convertService(s *events.Service) *v1.Service {
	// Ports
	var ports []*v1.Port
	for _, p := range s.Ports {
		ports = append(ports, convertPort(p))
	}
	service := &v1.Service{
		Name:  s.Name,
		Ports: ports,
	}

	var instances []*v1.Instance
	for _, i := range s.Instances {
		ins := &v1.Instance{}
		ins.Host = utils.RemovePort(i.Host)
		ins.Port = convertPort(i.Port)
		ins.Labels = i.Labels
		instances = append(instances, ins)
	}
	service.Instances = instances

	return service
}

// AddService ...
func (kubev2eh *KubeV2EventHandler) AddService(event events.ServiceEvent) {
	fmt.Printf("Kube v2 event handler: Adding a service\n%v\n", event.Service)

	// Transform a service event that noticed by zookeeper to a Service CRD
	// TODO we should resolve the application name from the meta data placed in a zookeeper node.
	appIdentifier := resolveAppIdentifier(&event)
	if appIdentifier == "" {
		fmt.Printf("Can not find an application name with this adding event: %v\n", event.Service.Name)
		return
	}

	amc := &v1.AppMeshConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appIdentifier,
			Namespace: defaultNamespace,
		},
	}

	_, err := kubev2eh.GetAmc(amc)
	kubev2eh.replace(&event, amc)
	if err != nil {
		fmt.Printf("Can not find an existed amc CR: %v\n", err)
		kubev2eh.CreateAmc(amc)
	} else {
		kubev2eh.UpdateAmc(amc)
	}

	fmt.Printf("Create or update an AppMeshConfig CR after a service has beed created: %s\n", amc.Name)
}

func (kubev2eh *KubeV2EventHandler) AddInstance(event events.ServiceEvent) {
	fmt.Printf("Kube v2 event handler: Adding an instance\n%v\n", event.Service)
	kubev2eh.AddService(event)
}

// DeleteService we assume we need to remove the service Spec part of AppMeshConfig
// after received a service deleted notification.
func (kubev2eh *KubeV2EventHandler) DeleteService(event events.ServiceEvent) {
	fmt.Printf("Kube v2 event handler: Deleting a service: %s\n", event.Service.Name)

	// TODO we should resolve the application name from the meta data placed in a zookeeper node.
	appIdentifier := resolveAppIdentifier(&event)
	// There is a chance to delete a service with an empty instances manually, but it is not be sure that which
	// amc should be modified.
	if appIdentifier == "" {
		fmt.Printf("Can not find an application name with this deleting event: %v\n", event.Service.Name)
		return
	}

	amc := &v1.AppMeshConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appIdentifier,
			Namespace: defaultNamespace,
		},
	}

	_, err := kubev2eh.GetAmc(amc)
	if err != nil {
		fmt.Println("amc CR can not be found, ignore it")
		return
	} else {
		if amc.Spec.Services != nil && len(amc.Spec.Services) > 0 {
			for i, s := range amc.Spec.Services {
				if s.Name == event.Service.Name {
					result := utils.DeleteInSlice(amc.Spec.Services, i)
					amc.Spec.Services = result.([]*v1.Service)
					break
					// TODO break? Can I assume there is no duplicate services belongs to a same amc?
				}
			}

			if len(amc.Spec.Services) == 0 {
				amc.Spec.Services = nil
			}

			kubev2eh.UpdateAmc(amc)
		} else {
			fmt.Println("The services list belongs to this amc CR is empty, ignore it")
			return
		}
	}
}

// DeleteInstance ...
func (kubev2eh *KubeV2EventHandler) DeleteInstance(event events.ServiceEvent) {
	fmt.Printf("Kube v2 event handler: deleting an instance\n%v\n", event.Instance)

	appIdentifier := resolveAppIdentifier(&event)
	amc := &v1.AppMeshConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appIdentifier,
			Namespace: defaultNamespace,
		},
	}
	_, err := kubev2eh.GetAmc(amc)
	if err != nil {
		fmt.Printf("The applicatin mesh configruation can not be found with key: %s", appIdentifier)
		return
	} else {
		deleteInstance(&event, amc)
	}

	kubev2eh.UpdateAmc(amc)
}

// CreateAmc
func (kubev2eh *KubeV2EventHandler) CreateAmc(amc *v1.AppMeshConfig) {
	// TODO
	cluster, _ := kubev2eh.K8sMgr.Get("tcc-gz01-bj5-test")
	err := cluster.Client.Create(context.Background(), amc)
	if err != nil {
		fmt.Printf("Creating an acm has an error:%v\n", err)
		return
	}
}

// UpdateAmc
func (kubev2eh *KubeV2EventHandler) UpdateAmc(amc *v1.AppMeshConfig) {
	// TODO
	cluster, _ := kubev2eh.K8sMgr.Get("tcc-gz01-bj5-test")
	err := cluster.Client.Update(context.Background(), amc)
	if err != nil {
		fmt.Printf("Updating an acm has an error: %v\n", err)
		return
	}
}

// GetAmc
func (kubev2eh *KubeV2EventHandler) GetAmc(config *v1.AppMeshConfig) (*v1.AppMeshConfig, error) {
	cluster, _ := kubev2eh.K8sMgr.Get("tcc-gz01-bj5-test")
	key, _ := client.ObjectKeyFromObject(config)
	err := cluster.Client.Get(context.Background(), key, config)
	return config, err
}

// AddConfigEntry
func (kubev2eh *KubeV2EventHandler) AddConfigEntry(e *events.ConfigEvent, identifierFinder func(a string) string) {
	fmt.Printf("Simple event handler: adding a configuration\n%v\n", e.Path)

	serviceName := e.ConfigEntry.Key
	appIdentifier := identifierFinder(serviceName)

	amc := &v1.AppMeshConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appIdentifier,
			Namespace: defaultNamespace,
		},
	}
	_, err := kubev2eh.GetAmc(amc)
	if err != nil {
		fmt.Printf("Finding amc with name %s has an error: %v\n", appIdentifier, err)
		// TODO Is there a requirement to requeue this event?
	} else {
		//for _, ci := range cc.Configs {
		//for address := range ci.Addresses {
		//
		//}
		//}
	}

}

func (kubev2eh *KubeV2EventHandler) ChangeConfigEntry(e *events.ConfigEvent) {
	fmt.Printf("Simple event handler: change a configuration\n%v\n", e.Path)
}

func (kubev2eh *KubeV2EventHandler) DeleteConfigEntry(e *events.ConfigEvent) {
	fmt.Printf("Simple event handler: delete a configuration\n%v\n", e.Path)
}