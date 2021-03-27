package common

import (
	"fmt"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"strings"
)

func UnzipPackage(srcPackage, dstPath string) error {
	if strings.Contains(srcPackage, "http") {
		downloadPackage := fmt.Sprintf("rm %s && wget srcPackage -O %s", constant.TMPPackgePath, constant.TMPPackgePath)
		if _, _, err := util.RunLinuxCommand(downloadPackage); err != nil {
			return err
		}
		srcPackage = constant.TMPPackgePath
	}

	tarUnzipCmd := fmt.Sprintf("tar -xzvf %s -C %s", srcPackage, dstPath)
	if _, _, err := util.RunLinuxCommand(tarUnzipCmd); err != nil {
		return err
	}
	return nil
}
