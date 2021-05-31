package hopsworksai

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/logicalclocks/terraform-provider-hopsworksai/hopsworksai/internal/api"
)

func TestAccClustersDataSourceAWS_basic(t *testing.T) {
	testAccClustersDataSource_basic(t, api.AWS)
}

func TestAccClustersDataSourceAZURE_basic(t *testing.T) {
	testAccClustersDataSource_basic(t, api.AZURE)
}

func testAccClustersDataSource_basic(t *testing.T, cloud api.CloudProvider) {
	suffix := acctest.RandString(5)
	rName := fmt.Sprintf("test_%s", suffix)
	resourceName := fmt.Sprintf("hopsworksai_cluster.%s", rName)
	dataSourceName := fmt.Sprintf("data.hopsworksai_clusters.%s", rName)
	parallelTest(t, cloud, resource.TestCase{
		PreCheck:  testAccPreCheck(t),
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig(cloud, rName, suffix),
				Check:  testAccClustersDataSourceCheckAllAttributes(cloud, resourceName, dataSourceName),
			},
		},
	})
}

func testAccClustersDataSourceCheckAllAttributes(cloud api.CloudProvider, resourceName string, dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found", dataSourceName)
		}

		var index string = ""
		listClustersTagPattern := regexp.MustCompile(`^clusters\.([0-9]*)\.tags.ListClusters$`)
		for k, v := range ds.Primary.Attributes {
			submatches := listClustersTagPattern.FindStringSubmatch(k)
			if len(submatches) == 2 && v == cloud.String() {
				index = submatches[1]
			}
		}

		if index == "" {
			return fmt.Errorf("no clusters returned")
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		for k := range rs.Primary.Attributes {
			if k == "id" || k == "%" || k == "*" {
				continue
			}
			dataSourceKey := fmt.Sprintf("clusters.%s.%s", index, k)
			if err := resource.TestCheckResourceAttrPair(resourceName, k, dataSourceName, dataSourceKey)(s); err != nil {
				return fmt.Errorf("Error while checking %s  err: %s", k, err)
			}
		}
		return nil
	}
}

func testAccClustersDataSourceConfig(cloud api.CloudProvider, rName string, suffix string) string {
	return fmt.Sprintf(`
	resource "hopsworksai_cluster" "%s" {
		name    = "%s%s%s"
		ssh_key = "%s"
		head {
		}

		%s


		tags = {
		  "ListClusters" = "%s"
		  "%s" = "%s"
		}
	  }

	  data "hopsworksai_clusters" "%s" {
		  depends_on = [
			hopsworksai_cluster.%s
		  ]
	  }
	`,
		rName,
		default_CLUSTER_NAME_PREFIX,
		strings.ToLower(cloud.String()),
		suffix,
		testAccClusterCloudSSHKeyAttribute(cloud),
		testAccClusterCloudConfigAttributes(cloud, 3),
		cloud.String(),
		default_CLUSTER_TAG_KEY,
		default_CLUSTER_TAG_VALUE,
		rName,
		rName,
	)
}
