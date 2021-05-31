package hopsworksai

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestAccClusterDataSourceAWS_basic(t *testing.T) {
	testAccClusterDataSource_basic(t, api.AWS)
}

func TestAccClusterDataSourceAZURE_basic(t *testing.T) {
	testAccClusterDataSource_basic(t, api.AZURE)
}

func testAccClusterDataSource_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	dataSourceName := fmt.Sprintf("data.hopsworksai_cluster.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:  testAccPreCheck(t),
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig(cloud, rName, suffix),
				Check:  testAccClusterDataSourceCheckAllAttributes(resourceName, dataSourceName),
			},
		},
	})
}

func testAccClusterDataSourceCheckAllAttributes(resourceName string, dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		for k := range rs.Primary.Attributes {
			if k == "id" || k == "%" || k == "*" {
				continue
			}
			if err := resource.TestCheckResourceAttrPair(resourceName, k, dataSourceName, k)(s); err != nil {
				return fmt.Errorf("Error while checking %s  err: %s", k, err)
			}
		}
		return nil
	}
}

func testAccClusterDataSourceConfig(cloud api.CloudProvider, rName string, suffix string) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "%s" {
		name    = "%s%s%s"
		ssh_key = "%s"	  
		head {
		}
		
		%s
		

		tags = {
		  "%s" = "%s"
		}
	  }

	  data "hopsworksai_cluster" "%s" {
		  cluster_id = hopsworksai_cluster.%s.id
	  }
	`,
		rName,
		default_CLUSTER_NAME_PREFIX,
		strings.ToLower(cloud.String()),
		suffix,
		testAccClusterCloudSSHKeyAttribute(cloud),
		testAccClusterCloudConfigAttributes(cloud, 2),
		default_CLUSTER_TAG_KEY,
		default_CLUSTER_TAG_VALUE,
		rName,
		rName,
	)
}
