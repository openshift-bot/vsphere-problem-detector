package check

import (
	"context"
	"flag"
	"time"

	ocpv1 "github.com/openshift/api/config/v1"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/legacy-cloud-providers/vsphere"
)

var (
	// Make the vSphere call timeout configurable.
	Timeout = flag.Duration("vmware-timeout", 10*time.Second, "Timeout of all VMware calls")

	// DefaultClusterChecks is the list of all checks.
	DefaultClusterChecks map[string]ClusterCheck = map[string]ClusterCheck{
		"CheckTaskPermissions":   CheckTaskPermissions,
		"ClusterInfo":            CollectClusterInfo,
		"CheckFolderPermissions": CheckFolderPermissions,
		"CheckDefaultDatastore":  CheckDefaultDatastore,
		"CheckPVs":               CheckPVs,
		"CheckStorageClasses":    CheckStorageClasses,
	}
	DefaultNodeChecks map[string]NodeCheck = map[string]NodeCheck{
		"CheckNodeDiskUUID": CheckNodeDiskUUID,
	}

	// NodeProperties is a list of properties that NodeCheck can rely on to be pre-filled.
	// Add a property to this list when a NodeCheck uses it.
	NodeProperties = []string{"config.extraConfig", "config.flags"}
)

// KubeClient is an interface between individual vSphere check and Kubernetes.
type KubeClient interface {
	// GetInfrastructure returns current Infrastructure instance.
	GetInfrastructure(ctx context.Context) (*ocpv1.Infrastructure, error)
	// ListNodes returns list of all nodes in the cluster.
	ListNodes(ctx context.Context) ([]v1.Node, error)
	// ListStorageClasses returns list of all storage classes in the cluster.
	ListStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error)
	// ListPVs returns list of all PVs in the cluster.
	ListPVs(ctx context.Context) ([]v1.PersistentVolume, error)
}

type CheckContext struct {
	Context    context.Context
	VMConfig   *vsphere.VSphereConfig
	VMClient   *vim25.Client
	KubeClient KubeClient
}

// Interface of a single vSphere cluster-level check. It gets connection to vSphere, vSphere config and connection to Kubernetes.
// It returns result of the check.
type ClusterCheck func(ctx *CheckContext) error

// Interface of a single vSphere node-level check. It gets connection to vSphere, vSphere config, connection to Kubernetes and a node to check.
// It returns result of the check. Reason for separate node-level checks:
// 1) We want to expose per node metrics what checks failed/succeeded.
// 2) When multiple checks need VM, we want to get it only once from the vSphere API.
type NodeCheck func(ctx *CheckContext, node *v1.Node, vm *mo.VirtualMachine) error
