package cloudrun

import (
	"context"
	"fmt"

	runapi "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
)

func GetCurrentInstanceCount(fqn string) (int, error) {
	ctx := context.Background()
	c, err := runapi.NewServicesClient(ctx)
	if err != nil {
		return 0, err
	}
	defer c.Close()

	req := &runpb.GetServiceRequest{Name: fqn}
	resp, err := c.GetService(ctx, req)
	if err != nil {
		return 0, err
	}
	if resp.Scaling != nil {
		fmt.Printf("Found min instaces for %s:%d\n", fqn, int(resp.Scaling.MinInstanceCount))
		return int(resp.Scaling.MinInstanceCount), nil
	}
	fmt.Printf("0 min instances configured for %s\n", fqn)
	return 0, nil
}

func SetMinInstanceCount(fqn string, count int) error {
	ctx := context.Background()
	c, err := runapi.NewServicesClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	req := &runpb.UpdateServiceRequest{
		Service: &runpb.Service{
			Name: fqn,
			Scaling: &runpb.ServiceScaling{
				MinInstanceCount: int32(count),
			},
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"scaling.min_instance_count"},
		},
	}
	op, err := c.UpdateService(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}
