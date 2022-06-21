package kendra_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFaq_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_kendra_faq.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFaqDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFaqConfig_basic(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFaqExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`index/.+/faq/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "language_code", "en"),
					resource.TestCheckResourceAttr(resourceName, "name", rName5),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test_faq", "arn"),
					resource.TestCheckResourceAttr(resourceName, "s3_path.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_path.0.bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_path.0.key", "aws_s3_object.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.FaqStatusActive)),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckFaqDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kendra_faq" {
			continue
		}

		id, indexId, err := tfkendra.FaqParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = tfkendra.FindFaqByID(context.TODO(), conn, id, indexId)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckFaqExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kendra Faq is set")
		}

		id, indexId, err := tfkendra.FaqParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

		_, err = tfkendra.FindFaqByID(context.TODO(), conn, id, indexId)

		if err != nil {
			return fmt.Errorf("Error describing Kendra Faq: %s", err.Error())
		}

		return nil
	}
}

func testAccFaqConfigBase(rName, rName2, rName3, rName4 string) string {
	// Kendra IAM policies: https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_kms_key" "this" {
  key_id = "alias/aws/kendra"
}
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["kendra.amazonaws.com"]
    }
  }
}
data "aws_iam_policy_document" "test_faq" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}
data "aws_iam_policy_document" "test_index" {
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData"
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "cloudwatch:namespace"

      values = [
        "Kendra"
      ]
    }
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogGroups"
    ]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*"
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogStreams",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*:log-stream:*"
    ]
  }
}

resource "aws_iam_policy" "test_faq" {
  name        = %[1]q
  description = "Allow Kendra to access S3"
  policy      = data.aws_iam_policy_document.test_faq.json
}

resource "aws_iam_policy" "test_index" {
  name        = %[2]q
  description = "Kendra Index IAM permissions"
  policy      = data.aws_iam_policy_document.test_index.json
}

resource "aws_iam_role_policy_attachment" "test_faq" {
  role       = aws_iam_role.test_faq.name
  policy_arn = aws_iam_policy.test_faq.arn
}

resource "aws_iam_role_policy_attachment" "test_index" {
  role       = aws_iam_role.test_index.name
  policy_arn = aws_iam_policy.test_index.arn
}

resource "aws_iam_role" "test_faq" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role" "test_index" {
  name               = %[2]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_kendra_index" "test" {
  depends_on = [aws_iam_role_policy_attachment.test_index]
  name       = %[3]q
  role_arn   = aws_iam_role.test_index.arn
}

resource "aws_s3_bucket" "test" {
  bucket        = %[4]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  source = "test-fixtures/basic.csv"
  key    = "test/basic.csv"
}
`, rName, rName2, rName3, rName4)
}

func testAccFaqConfig_basic(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccFaqConfigBase(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_kendra_faq" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test_faq.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}
`, rName5))
}
