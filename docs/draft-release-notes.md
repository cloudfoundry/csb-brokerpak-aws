## Release notes for next release:

### New feature:
- S3: region updates for existing buckets are now blocked by the broker resulting in faster feedback and improved error message.

### Fix:
- minimum constraints on MySQL and PostreSQL storage_gb are now enforced
- adds lifecycle.prevent_destroy to all data services to provide extra layer of protection against data loss
