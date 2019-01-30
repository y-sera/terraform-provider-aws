package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsBackupSelection_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsBackupSelectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupSelectionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsBackupSelectionExists("aws_backup_selection.test"),
				),
			},
		},
	})
}

func testAccCheckAwsBackupSelectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).backupconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_selection" {
			continue
		}

		input := &backup.GetBackupSelectionInput{
			BackupPlanId: aws.String(rs.Primary.Attributes["plan_id"]),
			SelectionId:  aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetBackupSelection(input)

		if err == nil {
			if *resp.SelectionId == rs.Primary.ID {
				return fmt.Errorf("Selection '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsBackupSelectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s, %v", name, s.RootModule().Resources)
		}
		return nil
	}
}

func testAccBackupSelectionConfig(randInt int) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_backup_vault" "test" {
	name = "tf_acc_test_backup_vault_%d"
}

resource "aws_backup_plan" "test" {
	name = "tf_acc_test_backup_plan_%d"

	rule {
		rule_name 			= "tf_acc_test_backup_rule_%d"
		target_vault_name 	= "${aws_backup_vault.test.name}"
		schedule			= "cron(0 12 * * ? *)"
	}
}

resource "aws_backup_selection" "test" {
	plan_id 	= "${aws_backup_plan.test.id}"

	name 		= "tf_acc_test_backup_selection_%d"
	iam_role 	= "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"

	tag {
		type = "STRINGEQUALS"
		key = "foo"
		value = "bar"
	}

	resources = [
		"arn:aws:ec2:us-east-1:${data.aws_caller_identity.current.account_id}:volume/"
	]
}
`, randInt, randInt, randInt, randInt)
}
