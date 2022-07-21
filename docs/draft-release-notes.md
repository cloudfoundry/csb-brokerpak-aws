## Release notes for next release:

### Features
- region property as a text field instead of an enumerated enabling selection of any region available in the Cloud Provider
- S3: region updates for existing buckets are now blocked by the broker resulting in faster feedback and improved error message.
- S3: ACL can now be specified on creation/update if the plan does not specify a value for it. Previously it was a plan-only input and as such could only be specified in the plan definition.
- S3: Bucket Ownership controls can now be specified in a plan or on creation/update if the plan does not specify a value for it.
- Beta tag: all service offerings tagged as beta and will not be displayed by default in the marketplace. Set the environment variable `GSB_COMPATIBILITY_ENABLE_BETA_SERVICES` to true to enable them. 

### Fix:
- minimum constraints on MySQL and PostreSQL storage_gb are now enforced
- adds lifecycle.prevent_destroy to all data services to provide extra layer of protection against data loss