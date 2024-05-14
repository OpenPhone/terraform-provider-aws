// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rum_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchrum "github.com/hashicorp/terraform-provider-aws/internal/service/rum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRUMAppMonitor_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon cloudwatchrum.AppMonitor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttrSet(resourceName, "app_monitor_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "localhost"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", acctest.CtOne),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppMonitorConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttrSet(resourceName, "app_monitor_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "cw_log_group"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, "localhost"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", acctest.CtOne),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_customEvents(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon cloudwatchrum.AppMonitor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_customEvents(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "custom_events.0.status", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppMonitorConfig_customEvents(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "custom_events.0.status", "DISABLED"),
				),
			},
			{
				Config: testAccAppMonitorConfig_customEvents(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "custom_events.#", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "custom_events.0.status", "ENABLED"),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon cloudwatchrum.AppMonitor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppMonitorConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppMonitorConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", acctest.CtOne),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var appMon cloudwatchrum.AppMonitor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RUMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppMonitorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(ctx, resourceName, &appMon),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatchrum.ResourceAppMonitor(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatchrum.ResourceAppMonitor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppMonitorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RUMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rum_app_monitor" {
				continue
			}

			_, err := tfcloudwatchrum.FindAppMonitorByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch RUM App Monitor %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppMonitorExists(ctx context.Context, n string, v *cloudwatchrum.AppMonitor) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch RUM App Monitor ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RUMConn(ctx)

		output, err := tfcloudwatchrum.FindAppMonitorByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAppMonitorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"
}
`, rName)
}

func testAccAppMonitorConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name           = %[1]q
  domain         = "localhost"
  cw_log_enabled = true
}
`, rName)
}

func testAccAppMonitorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppMonitorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAppMonitorConfig_customEvents(rName, enabled string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"

  custom_events {
    status = %[2]q
  }
}
`, rName, enabled)
}
