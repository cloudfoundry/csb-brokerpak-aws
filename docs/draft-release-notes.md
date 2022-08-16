## Release notes for next release:

### Breaking
- PostgreSQL: the default storage type is now set as 'io1' (provisioned IOPS SSD). Previously the default used 'gp2' (general purpose SSD). Users who previously had custom plans should add the property `"storage_type":"gp2"` to the plan definition, to ensure the storage type is not amended on any update. 
  As part of this work, the default storage size has also been increased to 100GB, as this is the smallest storage supported by the 'io1' storage type.

### Features
- region property as a text field instead of an enumerated enabling selection of any region available in the Cloud Provider
- S3: region updates for existing buckets are now blocked by the broker resulting in faster feedback and improved error message.
- S3: ACL can now be specified on creation if the plan does not specify a value for it. Previously it was a plan-only input and as such could only be specified in the plan definition.
- S3: Bucket Ownership controls can now be specified in a plan or on creation if the plan does not specify a value for it. It defaults to `ObjectOwnershipEnforced` and this disables ACLs by default. If you have custom plans refer to the upgrading instructions for information regarding this change.
- S3: Blocking public access to Amazon S3 storage. This feature provides settings for buckets to help manage public access to Amazon S3 resources. S3 Block Public Access settings override policies and permissions so that it is possible to limit public access to these resources.
- S3: Server Side encryption can now be enabled and configured. This feature provides settings for configuring encryption of data in an S3 bucket.
- S3: Object Lock. This feature allows storing objects using a write-once-read-many (WORM) model. Object Lock can help prevent objects from being deleted or overwritten for a fixed amount of time.
- S3: There are no default plans defined. Plans must be configured through the environment variable: `GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS`.
- S3: Allow versioning updates. We add the ability to modify the versioning of an S3 bucket, to enable such functionality in step after its creation. Once versioning is enabled, it can no longer be disabled as the IaaS will throw an error.
- Beta tag: all service offerings tagged as beta and will not be displayed by default in the marketplace. Set the environment variable `GSB_COMPATIBILITY_ENABLE_BETA_SERVICES` to true to enable them. 
- PostgreSQL: when creating a binding, by default the PostgreSQL connection will be secured via the "verify-full" PostgreSQL configuration. This will require the AWS certificate bundle to be installed, or it can be disabled by setting "use_tls=false"
- PostgreSQL: a new "provider_verify_certificate" property allows for the PostgreSQL Terraform provider to skip the verification of the server certificate.
- PostgreSQL: server can reject non-SSL connections by default. Renamed "use_tls" to "require_ssl". When the "require_ssl" property is true, it will make the server require SSL connections. When false (default), the server will accept SSL and non-SSL connections.
- PostgreSQL: Enhanced Monitoring. Amazon RDS provides metrics in real time for the operating system (OS) of the DB instance. Enhanced Monitoring enables all the system metrics and process information for the RDS DB instances on the console.
- PostgreSQL: Only "instance_class" are now exposed when provisioning or updating an instance. The previous “cores” abstraction is deprecated, in favour of using the underlying AWS instance class property.
- PostgreSQL: Automated backups can now be scheduled through "backup_window". By default, the automated backups are disabled.
- PostgreSQL: Automated backups can be customised through the following properties: "delete_automated_backups" - delete backups when deleting the instance, defaults to true; "copy_tags_to_snapshot" - copy all instance tags to snapshots, defaults to true. 
- PostgreSQL: Enable encryption with a custom key. Amazon RDS encrypted DB instances provide an additional layer of data protection by securing data from unauthorized access to the underlying storage. Amazon RDS uses an AWS KMS key to encrypt these resources, and now a custom key with the desired configuration can be used.
- PostgreSQL: Added deprecation warning to `cores` property and made it optional. It is recommended to use the `instance_class` property instead. 
- PostgreSQL: Performance Insights can now be enabled and a kms key can be provided to encrypt the performance insights data. Performance insights is disabled by default.
- PostgreSQL: The storage type can now be defined through the property "storage_type". In addition to this, if using the provisioned IOPS SSD (io1) storage type, then the 'iops' value can also be defined through the property "iops".
- Terraform upgrade (from 0.12.30 to 1.1.9) has been added

### Fix:
- minimum constraints on MySQL and PostgreSQL storage_gb are now enforced
- adds lifecycle.prevent_destroy to all data services to provide extra layer of protection against data loss
- Modification of the region generates the same service without eliminating the existing one in the newly established region. Blocking updating operation of such property to avoid the generation of infrastructure unintentionally.
- PostgreSQL role is now always cleanly deleted during unbind
