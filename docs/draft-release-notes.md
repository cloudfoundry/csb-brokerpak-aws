## Release notes for next release:

### Features
- region property as a text field instead of an enumerated enabling selection of any region available in the Cloud Provider
- S3: region updates for existing buckets are now blocked by the broker resulting in faster feedback and improved error message.
- S3: ACL can now be specified on creation/update if the plan does not specify a value for it. Previously it was a plan-only input and as such could only be specified in the plan definition.
- S3: Bucket Ownership controls can now be specified in a plan or on creation/update if the plan does not specify a value for it.
- S3: Blocking public access to Amazon S3 storage. This feature provides settings for buckets to help manage public access to Amazon S3 resources. S3 Block Public Access settings override policies and permissions so that it is possible to limit public access to these resources.
- S3: Server Side encryption can now be enabled and configured. This feature provides settings for configuring encryption of data in an S3 bucket.
- S3: Object Lock. This feature allows storing objects using a write-once-read-many (WORM) model. Object Lock can help prevent objects from being deleted or overwritten for a fixed amount of time.
- S3: There are no default plans defined. Plans must be configured through the environment variable: `GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS`.
- S3: Allow versioning updates. We add the ability to modify the versioning of an S3 bucket, to enable such functionality in step after its creation. Once versioning is enabled, it can no longer be disabled as the IaaS will throw an error.
- Beta tag: all service offerings tagged as beta and will not be displayed by default in the marketplace. Set the environment variable `GSB_COMPATIBILITY_ENABLE_BETA_SERVICES` to true to enable them. 
- PostgreSQL: when creating a binding, by default the PostgreSQL connection will be secured via the "verify-full" PosgreSQL configuration. This will require the AWS certificate bundle to be installed, or it can be disabled by setting "use_tls=false"
- PostgreSQL: a new "provider_verify_certificate" property allows for the PostgreSQL Terraform provider to skip the verification of the server certificate.
- PostgreSQL: server can reject non-SSL connections by default. Renamed "use_tls" to "require_ssl". Wheh the "require_ssl" property is true, it will make the server require SSL connections. When false (default), the server will accept SSL and non-SSL connections.
- Terraform upgrade (from 0.12.30 to 0.12.31) has been added
- PostgreSQL: Only "instance_class" are now exposed when provisioning or updating an instance. The previous “cores” abstraction is removed, in favor of using the underlying AWS instance class property.
- PostgreSQL: Automated backups can now be scheduled through "backup_window". By default, the automated backups are disabled.
- PostgreSQL: Automated backups can be customised through the following properties: "delete_automated_backups" - delete backups when deleting the instance, defaults to true; "copy_tags_to_snapshot" - copy all instance tags to snapshots, defaults to true. 

### Fix:
- minimum constraints on MySQL and PostreSQL storage_gb are now enforced
- adds lifecycle.prevent_destroy to all data services to provide extra layer of protection against data loss
- Modification of the region generates the same service without eliminating the existing one in the newly established region. Blocking updating operation of such property to avoid the generation of infrastructure unintentionally.
