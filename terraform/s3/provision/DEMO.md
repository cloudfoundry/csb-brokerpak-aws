# Agenda

I will focus on:
- Current state of the prototype (generic for any service, few changes required in csb-core)
- Solutions for things impossible/painful today
- Demoing real world use cases
- Sneak peak possible new opportunities

I won’t focus on:
- I will not focus:
  - Service specific improvements
  - Things that continue working the same
  - Implementation details or code changes

List of Demos:
- Creating a new instance (TF inputs autoremember)
- Updating an instance (surprising behaviour)
- Solution: Changing default value for an existing property
- Exposing a new property
- Solution: Deprecating a property
- Solution: `prohibit_update` corner-case


# Demos

## Creating a new instance

- terraform plan (fails)
- create terraform.tfstate
        ```
        {
          "version": 4,
          "check_results": null
        }
        ```
- terraform plan (succeeds)
- terraform apply


## Updating an instance

- Modify `pab_restrict_public_buckets` in terraform.tfvars
- terraform plan (shows some surprising changes)
  - prohibit_update validations
- terraform apply
- terraform apply
  (Default values assigned by the IAAS keep appearing. We’ll see how to fix this later)


## Changing default value for an existing property

- Set `require_tls: true` in variables.tf
- terraform plan (default value is ignored)
- rename terraform.tfstate & terraform.tfstate.backup
- create terraform.tfstate from template
- terraform plan (default value is observed)
- restore terraform.tfsate & terraform.tfstate.backup
- terraform apply


## Exposing a new property

- Pass `force_destroy` to `aws_s3_bucket` in `main.tf`
- terraform plan (fails: missing attribute)
- Add value in `terraform.tfvars`
- terraform plan (fails: unsupported attribute)
- Add type in `variables.tf`
- terraform plan (succeeds)
- Allow upgrading brokerpak without interaction:
  * Remove `force_destroy` from terraform.tfvars
  * Add `force_destroy` default in `variables.tf`
- terraform plan (succeeds)
- terraform apply


## Deprecating a property

The first two steps replicate the worst-case scenario where users specify the deprecated property in their plan:

- Add `pab_restrict_public_buckets` to terraform.tfvars
- terraform apply (succeeds)

————————————————————

- Remove it from variables.tf
- terraform plan (fails: unsupported properties specified)
- Remove it from terraform.tfvars
- terraform plan (fails: unsupported properties specified)
- Add to `deprecated_inputs` in variables.tf
- terraform plan (fails: missing attribute in main.tf)
- Replace with new functionality in main.tf
- terraform plan (succeeds)
- terraform apply


## Show fix for `prohibit_update` corner-case

- Add “boc_object_ownership”: “ObjectWriter” to terraform.tfvars
- terraform plan (fails: prohibit_update)
- Remove “boc_object_ownership” from `prohibit_updates` list in variables.tf
- terraform plan (succeeds)
- terraform apply
- Re-add “boc_object_ownership” to `prohibit_updates` list in variables.tf
- terraform apply
 (explain why this fixes the existing corner-case)


# Future opportunities

More fine-grain defaults:
- service_defaults
- *plan_defaults
- plan_enforced

————————————————————

Alternatives to current custom mysql:
- Terraform pg backend

Cross Foundation, Truly Multi-Cloud services:
- Terraform remote state, pg backend, locking and workspaces

