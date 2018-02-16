/* Copyright © 2017 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vmware-nsxt"
	"net/http"
	"testing"
)

func TestNSXSNATRuleBasic(t *testing.T) {

	ruleName := fmt.Sprintf("test-nsx-snat-rule")
	updateRuleName := fmt.Sprintf("%s-update", ruleName)
	testResourceName := "nsxt_nat_rule.test"
	edgeClusterName := EdgeClusterDefaultName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNATRuleCheckDestroy(state, ruleName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXSNATRuleCreateTemplate(ruleName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNATRuleCheckExists(ruleName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", ruleName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "logging", "true"),
					resource.TestCheckResourceAttr(testResourceName, "nat_pass", "false"),
					resource.TestCheckResourceAttr(testResourceName, "action", "SNAT"),
					resource.TestCheckResourceAttr(testResourceName, "translated_network", "4.4.4.0/24"),
					resource.TestCheckResourceAttr(testResourceName, "match_destination_network", "3.3.3.0/24"),
					resource.TestCheckResourceAttr(testResourceName, "match_source_network", "5.5.5.0/24"),
				),
			},
			{
				Config: testAccNSXSNATRuleUpdateTemplate(updateRuleName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNATRuleCheckExists(updateRuleName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateRuleName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "logging", "true"),
					resource.TestCheckResourceAttr(testResourceName, "nat_pass", "true"),
					resource.TestCheckResourceAttr(testResourceName, "action", "SNAT"),
					resource.TestCheckResourceAttr(testResourceName, "translated_network", "4.4.4.0/24"),
					resource.TestCheckResourceAttr(testResourceName, "match_destination_network", "3.3.3.0/24"),
					resource.TestCheckResourceAttr(testResourceName, "match_source_network", "6.6.6.0/24"),
				),
			},
		},
	})
}

func TestNSXDNATRuleBasic(t *testing.T) {

	ruleName := fmt.Sprintf("test-nsx-dnat-rule")
	updateRuleName := fmt.Sprintf("%s-update", ruleName)
	testResourceName := "nsxt_nat_rule.test"
	edgeClusterName := EdgeClusterDefaultName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNATRuleCheckDestroy(state, ruleName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDNATRuleCreateTemplate(ruleName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNATRuleCheckExists(ruleName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", ruleName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "logging", "true"),
					resource.TestCheckResourceAttr(testResourceName, "nat_pass", "true"),
					resource.TestCheckResourceAttr(testResourceName, "action", "DNAT"),
					resource.TestCheckResourceAttr(testResourceName, "translated_network", "4.4.4.4"),
					resource.TestCheckResourceAttr(testResourceName, "match_destination_network", "3.3.3.0/24"),
				),
			},
			{
				Config: testAccNSXDNATRuleUpdateTemplate(updateRuleName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNATRuleCheckExists(updateRuleName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateRuleName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "logging", "true"),
					resource.TestCheckResourceAttr(testResourceName, "nat_pass", "true"),
					resource.TestCheckResourceAttr(testResourceName, "action", "DNAT"),
					resource.TestCheckResourceAttr(testResourceName, "translated_network", "4.4.4.4"),
					resource.TestCheckResourceAttr(testResourceName, "match_destination_network", "7.7.7.0/24"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
				),
			},
		},
	})
}

func testAccNSXNATRuleCheckExists(display_name string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(*nsxt.APIClient)

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX nat rule resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX nat rule resource ID not set in resources ")
		}
		router_id := rs.Primary.Attributes["logical_router_id"]
		if router_id == "" {
			return fmt.Errorf("NSX nat rule router_id not set in resources ")
		}

		natRule, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.GetNatRule(nsxClient.Context, router_id, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving nat rule ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if nat rule %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if display_name == natRule.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX nat rule %s wasn't found", display_name)
	}
}

func testAccNSXNATRuleCheckDestroy(state *terraform.State, display_name string) error {

	nsxClient := testAccProvider.Meta().(*nsxt.APIClient)

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_nat_rule" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		router_id := rs.Primary.Attributes["logical_router_id"]
		natRule, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.GetNatRule(nsxClient.Context, router_id, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving nat rule ID %s. Error: %v", resourceID, err)
		}

		if display_name == natRule.DisplayName {
			return fmt.Errorf("NSX nat rule %s still exists", display_name)
		}
	}
	return nil
}

func testAccNSXNATRulePreConditionTemplate(edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
	display_name = "%s"
}

resource "nsxt_logical_tier1_router" "RTR1" {
	display_name = "tier1_router"
	edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
}`, edgeClusterName)
}

func testAccNSXSNATRuleCreateTemplate(name string, edgeClusterName string) string {
	return testAccNSXNATRulePreConditionTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_nat_rule" "test" {
	logical_router_id = "${nsxt_logical_tier1_router.RTR1.id}"
    display_name = "%s"
	description = "Acceptance Test"
    action = "SNAT"
    translated_network = "4.4.4.0/24"
    match_destination_network = "3.3.3.0/24"
    match_source_network = "5.5.5.0/24"
    enabled = true
    logging = true
    nat_pass = false
	tags = [{scope = "scope1", tag = "tag1"}]
}`, name)
}

func testAccNSXSNATRuleUpdateTemplate(name string, edgeClusterName string) string {
	return testAccNSXNATRulePreConditionTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_nat_rule" "test" {
	logical_router_id = "${nsxt_logical_tier1_router.RTR1.id}"
    display_name = "%s"
	description = "Acceptance Test Update"
    action = "SNAT"
    translated_network = "4.4.4.0/24"
    match_destination_network = "3.3.3.0/24"
    match_source_network = "6.6.6.0/24"
    enabled = false
    logging = true
    nat_pass = true
	tags = [{scope = "scope1", tag = "tag1"},
	        {scope = "scope2", tag = "tag2"}]
}`, name)
}

func testAccNSXDNATRuleCreateTemplate(name string, edgeClusterName string) string {
	return testAccNSXNATRulePreConditionTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_nat_rule" "test" {
	logical_router_id = "${nsxt_logical_tier1_router.RTR1.id}"
    display_name = "%s"
	description = "Acceptance Test"
    action = "DNAT"
    translated_network = "4.4.4.4"
    match_destination_network = "3.3.3.0/24"
    enabled = true
    logging = true
    nat_pass = true
	tags = [{scope = "scope1", tag = "tag1"}]
}`, name)
}

func testAccNSXDNATRuleUpdateTemplate(name string, edgeClusterName string) string {
	return testAccNSXNATRulePreConditionTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_nat_rule" "test" {
	logical_router_id = "${nsxt_logical_tier1_router.RTR1.id}"
    display_name = "%s"
	description = "Acceptance Test Update"
    action = "DNAT"
    translated_network = "4.4.4.4"
    match_destination_network = "7.7.7.0/24"
    enabled = true
    logging = true
    nat_pass = true
	tags = [{scope = "scope1", tag = "tag1"},
	        {scope = "scope2", tag = "tag2"}]
}`, name)
}