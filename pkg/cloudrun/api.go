package cloudrun

import (
	"context"
	"fmt"

	runapi "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
)

func GetCurrentInstanceCount(project, region, service string) (int, error) {
	ctx := context.Background()
	c, err := runapi.NewServicesRESTClient(ctx)
	if err != nil {
		return 0, err
	}
	defer c.Close()

	req := &runpb.GetServiceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", project, region, service),
	}
	resp, err := c.GetService(ctx, req)
	if err != nil {
		return 0, err
	}
	if resp.Scaling != nil {
		fmt.Printf("Found min instaces for %s:%d\n", service, int(resp.Scaling.MinInstanceCount))
		return int(resp.Scaling.MinInstanceCount), nil
	}
	fmt.Printf("0 min instances configured for %s\n", service)
	return 0, nil
}
