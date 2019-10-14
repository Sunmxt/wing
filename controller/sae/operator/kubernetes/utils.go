package kubernetes

import "strconv"

func DeploymentNameFromServiceName(serviceName string, specID int) string {
	return "wing-dp-" + strconv.FormatInt(int64(specID), 10) + "-" + serviceName
}

func TestingDeploymentNameFromServiceName(serviceName string, specID int) string {
	return "wing-dp-test-" + strconv.FormatInt(int64(specID), 10) + "-" + serviceName
}
