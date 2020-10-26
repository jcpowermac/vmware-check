package check

import (
	"context"
	"fmt"

	_ "github.com/davecgh/go-spew/spew"
	"github.com/jsafrane/vmware-check/pkg/vmware"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"k8s.io/legacy-cloud-providers/vsphere"
)

func CheckVirtualMachineHardwareVersion(vmClient *govmomi.Client, config *vsphere.VSphereConfig) error {
	finder := find.NewFinder(vmClient.Client, false)
	ctx, cancel := context.WithTimeout(context.Background(), *vmware.Timeout)

	defer cancel()

	// Locate the virtual machines
	if config.Workspace.Folder != "" {
		folder, err := finder.Folder(ctx, config.Workspace.Folder)
		if err != nil {
			return fmt.Errorf("failed to access Datacenter %s: %s", config.Workspace.Folder, err)

		}
		spew.Dump(folder)

		refs, err := folder.Children(ctx)
		if err != nil {
			return fmt.Errorf("failed to access folder children %s: %s", config.Workspace.Folder, err)
		}

		pc := property.DefaultCollector(vmClient.Client)

		for _, ref := range refs {

			if ref.Reference().Type == "VirtualMachine" {

				var virtualMachineMo mo.VirtualMachine

				err = pc.RetrieveOne(ctx, ref.Reference(), nil, &virtualMachineMo)

				if err != nil {
					return fmt.Errorf("failed to access virtual machine property collector %s: %s", ref.Reference().Value, err)
				}

				spew.Dump(virtualMachineMo.Config.Version)

				//var content []types.ObjectContent
				var content []types.ObjectContent
				// https://github.com/vmware/govmomi/blob/9a4f1f95eafd9f18b77919731a6e9f5b423d75b3/govc/vm/option/info.go
				err = property.DefaultCollector(vmClient.Client).RetrieveOne(ctx, ref.Reference(), []string{"environmentBrowser"}, &content)

				req := types.QueryConfigOptionEx{
					This: content[0].PropSet[0].Val.(types.ManagedObjectReference),
					Spec: &types.EnvironmentBrowserConfigOptionQuerySpec{},
				}

				opt, err := methods.QueryConfigOptionEx(ctx, vmClient, &req)
				if err != nil {
					return fmt.Errorf("failed to QueryConfigOptionEx %s: %s", ref.Reference().Value, err)
				}

				fmt.Printf("diskUuidEnabled: %t", *opt.Returnval.GuestOSDescriptor[0].DiskUuidEnabled)

			}
		}
	}

	return nil

}
