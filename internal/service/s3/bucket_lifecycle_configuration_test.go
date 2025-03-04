package s3_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccS3BucketLifecycleConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      "1",
						"expiration.0.days": "365",
						"filter.#":          "0",
						"id":                rName,
						"status":            tfs3.LifecycleRuleStatusEnabled,
					}),
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

func TestAccS3BucketLifecycleConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfigurationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketLifecycleConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_filterWithPrefix(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	dateUpdated := time.Date(currTime.Year()+1, currTime.Month(), currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_Basic_updateConfig(rName, date, "logs/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      "1",
						"expiration.0.date": date,
						"filter.#":          "1",
						"filter.0.prefix":   "logs/",
						"id":                rName,
						"status":            tfs3.LifecycleRuleStatusEnabled,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfiguration_Basic_updateConfig(rName, dateUpdated, "tmp/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      "1",
						"expiration.0.date": dateUpdated,
						"filter.#":          "1",
						"filter.0.prefix":   "tmp/",
						"id":                rName,
						"status":            tfs3.LifecycleRuleStatusEnabled,
					}),
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

func TestAccS3BucketLifecycleConfiguration_Filter_ObjectSizeGreaterThan(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_filterWithObjectSizeGreaterThanConfig(rName, date, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":                      "1",
						"expiration.0.date":                 date,
						"filter.#":                          "1",
						"filter.0.object_size_greater_than": "100",
						"id":                                rName,
						"status":                            tfs3.LifecycleRuleStatusEnabled,
					}),
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

func TestAccS3BucketLifecycleConfiguration_Filter_ObjectSizeLessThan(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_filterWithObjectSizeLessThanConfig(rName, date, 500),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":                   "1",
						"expiration.0.date":              date,
						"filter.#":                       "1",
						"filter.0.object_size_less_than": "500",
						"id":                             rName,
						"status":                         tfs3.LifecycleRuleStatusEnabled,
					}),
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

func TestAccS3BucketLifecycleConfiguration_Filter_ObjectSizeRange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_filterWithObjectSizeRangeConfig(rName, date, 500, 64000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      "1",
						"expiration.0.date": date,
						"filter.#":          "1",
						"filter.0.and.#":    "1",
						"filter.0.and.0.object_size_greater_than": "500",
						"filter.0.and.0.object_size_less_than":    "64000",
						"id":                                      rName,
						"status":                                  tfs3.LifecycleRuleStatusEnabled,
					}),
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

func TestAccS3BucketLifecycleConfiguration_Filter_ObjectSizeRangeAndPrefix(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_filterWithObjectSizeRangeAndPrefixConfig(rName, date, 500, 64000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      "1",
						"expiration.0.date": date,
						"filter.#":          "1",
						"filter.0.and.#":    "1",
						"filter.0.and.0.object_size_greater_than": "500",
						"filter.0.and.0.object_size_less_than":    "64000",
						"filter.0.and.0.prefix":                   rName,
						"id":                                      rName,
						"status":                                  tfs3.LifecycleRuleStatusEnabled,
					}),
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

func TestAccS3BucketLifecycleConfiguration_disableRule(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_Basic_statusConfig(rName, tfs3.LifecycleRuleStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
				),
			},
			{
				Config: testAccBucketLifecycleConfiguration_Basic_statusConfig(rName, tfs3.LifecycleRuleStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"status": tfs3.LifecycleRuleStatusDisabled,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfiguration_Basic_statusConfig(rName, tfs3.LifecycleRuleStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"status": tfs3.LifecycleRuleStatusEnabled,
					}),
				),
			},
		},
	})
}

func TestAccS3BucketLifecycleConfiguration_multipleRules(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"
	date := time.Now()
	expirationDate := time.Date(date.Year(), date.Month(), date.Day()+14, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_multipleRulesConfig(rName, expirationDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                    "log",
						"expiration.#":          "1",
						"expiration.0.days":     "90",
						"filter.#":              "1",
						"filter.0.and.#":        "1",
						"filter.0.and.0.prefix": "log/",
						"filter.0.and.0.tags.%": "2",
						"status":                tfs3.LifecycleRuleStatusEnabled,
						"transition.#":          "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.transition.*", map[string]string{
						"days":          "30",
						"storage_class": s3.StorageClassStandardIa,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.transition.*", map[string]string{
						"days":          "60",
						"storage_class": s3.StorageClassGlacier,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"id":                "tmp",
						"expiration.#":      "1",
						"expiration.0.date": expirationDate,
						"filter.#":          "1",
						"filter.0.prefix":   "tmp/",
						"status":            tfs3.LifecycleRuleStatusEnabled,
					}),
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

func TestAccS3BucketLifecycleConfiguration_nonCurrentVersionExpiration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_nonCurrentVersionExpirationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"noncurrent_version_expiration.#":                 "1",
						"noncurrent_version_expiration.0.noncurrent_days": "90",
					}),
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

func TestAccS3BucketLifecycleConfiguration_nonCurrentVersionTransition(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_nonCurrentVersionTransitionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"noncurrent_version_transition.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.noncurrent_version_transition.*", map[string]string{
						"noncurrent_days": "30",
						"storage_class":   s3.StorageClassStandardIa,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.noncurrent_version_transition.*", map[string]string{
						"noncurrent_days": "60",
						"storage_class":   s3.StorageClassGlacier,
					}),
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

// Ensure backwards compatible with now-deprecated "prefix" configuration
func TestAccS3BucketLifecycleConfiguration_prefix(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_Basic_prefixConfig(rName, "path1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":      "1",
						"expiration.0.days": "365",
						"id":                rName,
						"prefix":            "path1/",
						"status":            tfs3.LifecycleRuleStatusEnabled,
					}),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23239
func TestAccS3BucketLifecycleConfiguration_Filter_Tag(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_Filter_TagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#":         "1",
						"expiration.0.days":    "365",
						"id":                   rName,
						"filter.#":             "1",
						"filter.0.tag.#":       "1",
						"filter.0.tag.0.key":   "key1",
						"filter.0.tag.0.value": "value1",
						"status":               tfs3.LifecycleRuleStatusEnabled,
					}),
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

func TestAccS3BucketLifecycleConfiguration_RuleExpiration_expireMarkerOnly(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_RuleExpiration_expiredDeleteMarkerConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#": "1",
						"expiration.0.expired_object_delete_marker": "true",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfiguration_RuleExpiration_expiredDeleteMarkerConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#": "1",
						"expiration.0.expired_object_delete_marker": "false",
					}),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11420
func TestAccS3BucketLifecycleConfiguration_RuleExpiration_emptyBlock(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_RuleExpiration_emptyConfigurationBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"expiration.#": "1",
					}),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15138
func TestAccS3BucketLifecycleConfiguration_ruleAbortIncompleteMultipartUpload(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_ruleAbortIncompleteMultipartUploadConfig(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"abort_incomplete_multipart_upload.#":                       "1",
						"abort_incomplete_multipart_upload.0.days_after_initiation": "7",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfiguration_ruleAbortIncompleteMultipartUploadConfig(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"abort_incomplete_multipart_upload.#":                       "1",
						"abort_incomplete_multipart_upload.0.days_after_initiation": "5",
					}),
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

// TestAccS3BucketLifecycleConfiguration_TransitionDate_standardIa validates the change to address
// https://github.com/hashicorp/terraform-provider-aws/issues/23117
// does not introduce a regression.
func TestAccS3BucketLifecycleConfiguration_TransitionDate_standardIa(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_dateTransitionConfig(rName, date, s3.TransitionStorageClassStandardIa),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
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

// TestAccS3BucketLifecycleConfiguration_TransitionDate_intelligentTiering validates the change to address
// https://github.com/hashicorp/terraform-provider-aws/issues/23117
// does not introduce a regression.
func TestAccS3BucketLifecycleConfiguration_TransitionDate_intelligentTiering(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_dateTransitionConfig(rName, date, s3.StorageClassIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23117
func TestAccS3BucketLifecycleConfiguration_TransitionStorageClassOnly_intelligentTiering(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_transitionStorageClassOnlyConfig(rName, s3.StorageClassIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.0.transition.*", map[string]string{
						"days":          "0",
						"date":          "",
						"storage_class": s3.StorageClassIntelligentTiering,
					}),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23117
func TestAccS3BucketLifecycleConfiguration_TransitionZeroDays_intelligentTiering(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_zeroDaysTransitionConfig(rName, s3.StorageClassIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
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

func TestAccS3BucketLifecycleConfiguration_TransitionUpdateBetweenDaysAndDate_intelligentTiering(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	currTime := time.Now()
	date := time.Date(currTime.Year(), currTime.Month()+1, currTime.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_zeroDaysTransitionConfig(rName, s3.StorageClassIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
				),
			},
			{
				Config: testAccBucketLifecycleConfiguration_dateTransitionConfig(rName, date, s3.StorageClassIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLifecycleConfiguration_zeroDaysTransitionConfig(rName, s3.StorageClassIntelligentTiering),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23228
func TestAccS3BucketLifecycleConfiguration_EmptyFilter_NonCurrentVersions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_lifecycle_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketLifecycleConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleConfiguration_EmptyFilter_NonCurrentVersionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLifecycleConfigurationExists(resourceName),
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

func testAccCheckBucketLifecycleConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_lifecycle_configuration" {
			continue
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketLifecycleConfigurationInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := verify.RetryOnAWSCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
			return conn.GetBucketLifecycleConfiguration(input)
		})

		if tfawserr.ErrCodeEquals(err, tfs3.ErrCodeNoSuchLifecycleConfiguration, s3.ErrCodeNoSuchBucket) {
			continue
		}

		if err != nil {
			return err
		}

		if config, ok := output.(*s3.GetBucketLifecycleConfigurationOutput); ok && config != nil && len(config.Rules) != 0 {
			return fmt.Errorf("S3 Lifecycle Configuration for bucket (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketLifecycleConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketLifecycleConfigurationInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := verify.RetryOnAWSCode(tfs3.ErrCodeNoSuchLifecycleConfiguration, func() (interface{}, error) {
			return conn.GetBucketLifecycleConfiguration(input)
		})

		if err != nil {
			return err
		}

		if config, ok := output.(*s3.GetBucketLifecycleConfigurationOutput); !ok || config == nil {
			return fmt.Errorf("S3 Bucket Replication Configuration for bucket (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketLifecycleConfigurationBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = %[1]q
    status = "Enabled"

    expiration {
      days = 365
    }
  }
}
`, rName)
}

func testAccBucketLifecycleConfiguration_Basic_statusConfig(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = %[1]q
    status = %[2]q

    expiration {
      days = 365
    }
  }
}
`, rName, status)
}

func testAccBucketLifecycleConfiguration_Basic_updateConfig(rName, date, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = %[1]q
    status = "Enabled"

    expiration {
      date = %[2]q
    }

    # One of prefix or filter required to ensure XML is well-formed
    filter {
      prefix = %[3]q
    }
  }
}
`, rName, date, prefix)
}

func testAccBucketLifecycleConfiguration_Basic_prefixConfig(rName, prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id = %[1]q

    # One of prefix or filter required to ensure XML is well-formed
    prefix = %[2]q
    status = "Enabled"

    expiration {
      days = 365
    }
  }
}
`, rName, prefix)
}

func testAccBucketLifecycleConfiguration_Filter_TagConfig(rName, key, value string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = %[1]q
    status = "Enabled"

    filter {
      tag {
        key   = %[2]q
        value = %[3]q
      }
    }

    expiration {
      days = 365
    }
  }
}
`, rName, key, value)
}

func testAccBucketLifecycleConfiguration_RuleExpiration_expiredDeleteMarkerConfig(rName string, expired bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id     = %[1]q
    status = "Enabled"

    expiration {
      expired_object_delete_marker = %[2]t
    }
  }
}
`, rName, expired)
}

func testAccBucketLifecycleConfiguration_RuleExpiration_emptyConfigurationBlockConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id     = %[1]q
    status = "Enabled"

    expiration {}
  }
}
`, rName)
}

func testAccBucketLifecycleConfiguration_ruleAbortIncompleteMultipartUploadConfig(rName string, days int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    abort_incomplete_multipart_upload {
      days_after_initiation = %[2]d
    }

    id     = %[1]q
    status = "Enabled"
  }
}
`, rName, days)
}

func testAccBucketLifecycleConfiguration_multipleRulesConfig(rName, date string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = "log"

    expiration {
      days = 90
    }

    filter {
      and {
        prefix = "log/"

        tags = {
          key1 = "value1"
          key2 = "value2"
        }
      }
    }

    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "GLACIER"
    }
  }

  rule {
    id = "tmp"

    filter {
      prefix = "tmp/"
    }

    expiration {
      date = %[2]q
    }

    status = "Enabled"
  }
}
`, rName, date)
}

func testAccBucketLifecycleConfiguration_nonCurrentVersionExpirationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    # One of prefix or filter required to ensure XML is well-formed
    filter {
      prefix = "config/"
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }

    status = "Enabled"
  }
}
`, rName)
}

func testAccBucketLifecycleConfiguration_nonCurrentVersionTransitionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    # One of prefix or filter required to ensure XML is well-formed
    filter {
      prefix = "config/"
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    noncurrent_version_transition {
      noncurrent_days = 60
      storage_class   = "GLACIER"
    }

    status = "Enabled"
  }
}
`, rName)
}

func testAccBucketLifecycleConfiguration_transitionStorageClassOnlyConfig(rName, storageClass string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }

    noncurrent_version_transition {
      noncurrent_days = 0
      storage_class   = "INTELLIGENT_TIERING"
    }

    transition {
      storage_class = %[2]q
    }

    status = "Enabled"
  }
}
`, rName, storageClass)
}

func testAccBucketLifecycleConfiguration_zeroDaysTransitionConfig(rName, storageClass string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }

    noncurrent_version_transition {
      noncurrent_days = 0
      storage_class   = "INTELLIGENT_TIERING"
    }

    transition {
      days          = 0
      storage_class = %[2]q
    }

    status = "Enabled"
  }
}
`, rName, storageClass)
}

func testAccBucketLifecycleConfiguration_dateTransitionConfig(rName, transitionDate, storageClass string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }

    noncurrent_version_transition {
      noncurrent_days = 0
      storage_class   = "INTELLIGENT_TIERING"
    }

    transition {
      date          = %[2]q
      storage_class = %[3]q
    }

    status = "Enabled"
  }
}
`, rName, transitionDate, storageClass)
}

func testAccBucketLifecycleConfiguration_EmptyFilter_NonCurrentVersionsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    filter {}

    noncurrent_version_expiration {
      newer_noncurrent_versions = 2
      noncurrent_days           = 30
    }

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }

    status = "Enabled"
  }
}
`, rName)
}

func testAccBucketLifecycleConfiguration_filterWithObjectSizeGreaterThanConfig(rName, date string, sizeGreaterThan int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    expiration {
      date = %[2]q
    }

    filter {
      object_size_greater_than = %[3]d
    }

    status = "Enabled"
  }
}
`, rName, date, sizeGreaterThan)
}

func testAccBucketLifecycleConfiguration_filterWithObjectSizeLessThanConfig(rName, date string, sizeLessThan int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    expiration {
      date = %[2]q
    }

    filter {
      object_size_less_than = %[3]d
    }

    status = "Enabled"
  }
}
`, rName, date, sizeLessThan)
}

func testAccBucketLifecycleConfiguration_filterWithObjectSizeRangeConfig(rName, date string, sizeGreaterThan, sizeLessThan int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    expiration {
      date = %[2]q
    }

    filter {
      and {
        object_size_greater_than = %[3]d
        object_size_less_than    = %[4]d
      }
    }

    status = "Enabled"
  }
}
`, rName, date, sizeGreaterThan, sizeLessThan)
}

func testAccBucketLifecycleConfiguration_filterWithObjectSizeRangeAndPrefixConfig(rName, date string, sizeGreaterThan, sizeLessThan int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    id = %[1]q

    expiration {
      date = %[2]q
    }

    filter {
      and {
        object_size_greater_than = %[3]d
        object_size_less_than    = %[4]d
        prefix                   = %[1]q
      }
    }

    status = "Enabled"
  }
}
`, rName, date, sizeGreaterThan, sizeLessThan)
}
