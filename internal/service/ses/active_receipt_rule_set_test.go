package ses_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
)

// Only one SES Receipt RuleSet can be active at a time, so run serially
// locally and in TeamCity.
func TestAccSESActiveReceiptRuleSet_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Resource": {
			"basic":      testAccActiveReceiptRuleSet_basic,
			"disappears": testAccActiveReceiptRuleSet_disappears,
		},
		"DataSource": {
			"basic":           testAccActiveReceiptRuleSetDataSource_basic,
			"noActiveRuleSet": testAccActiveReceiptRuleSetDataSource_noActiveRuleSet,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccActiveReceiptRuleSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_active_receipt_rule_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ses.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActiveReceiptRuleSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActiveReceiptRuleSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveReceiptRuleSetExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-rule-set/%s", rName)),
				),
			},
		},
	})
}

func testAccActiveReceiptRuleSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_active_receipt_rule_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(ctx, t)
			testAccPreCheckReceiptRule(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ses.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckActiveReceiptRuleSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccActiveReceiptRuleSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveReceiptRuleSetExists(ctx, resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfses.ResourceActiveReceiptRuleSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckActiveReceiptRuleSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_active_receipt_rule_set" {
				continue
			}

			response, err := conn.DescribeActiveReceiptRuleSetWithContext(ctx, &ses.DescribeActiveReceiptRuleSetInput{})
			if err != nil {
				return err
			}

			if response.Metadata != nil && (aws.StringValue(response.Metadata.Name) == rs.Primary.ID) {
				return fmt.Errorf("Active receipt rule set still exists")
			}
		}

		return nil
	}
}

func testAccCheckActiveReceiptRuleSetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Active Receipt Rule Set not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Active Receipt Rule Set name not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn()

		response, err := conn.DescribeActiveReceiptRuleSetWithContext(ctx, &ses.DescribeActiveReceiptRuleSetInput{})
		if err != nil {
			return err
		}

		if response.Metadata != nil && (aws.StringValue(response.Metadata.Name) != rs.Primary.ID) {
			return fmt.Errorf("The active receipt rule set (%s) was not set to %s", aws.StringValue(response.Metadata.Name), rs.Primary.ID)
		}

		return nil
	}
}

func testAccActiveReceiptRuleSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_active_receipt_rule_set" "test" {
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
}
`, name)
}
